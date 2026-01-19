package app

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
)

const (
	spinnerTickInterval = 50 * time.Millisecond
)

var spinnerFrames = []rune{'|', '/', '-', '\\'}

type App struct {
	g            *gocui.Gui
	config       AppConfig
	panels       map[string]Panel
	focusOrder   []string
	currentFocus int

	// Modal management
	activeModal Modal
	savedFocus  int

	// Command execution tracking
	commandRunning     atomic.Bool   // Thread-safe flag for command execution
	runningCommandName atomic.Value  // Name of currently running command (string)
	spinnerFrame       atomic.Uint32 // Current spinner frame index (0-3)
	stopSpinnerCh      chan struct{} // Channel to stop spinner goroutine

	// Studio process management
	studioCmd     *commands.Command // Running studio command
	studioRunning bool              // True if studio is running
}

type AppConfig struct {
	DebugMode bool
	AppName   string
	Version   string
	Developer string
}

func NewApp(config AppConfig) (*App, error) {
	g, err := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if err != nil {
		return nil, err
	}

	app := &App{
		g:             g,
		config:        config,
		panels:        make(map[string]Panel),
		focusOrder:    []string{ViewWorkspace, ViewMigrations, ViewDetails, ViewOutputs},
		currentFocus:  0,
		stopSpinnerCh: make(chan struct{}),
	}

	g.SetManagerFunc(gocui.ManagerFunc(app.layoutManager))
	g.Mouse = true
	g.ShowListFooter = true

	// Start spinner update goroutine
	app.startSpinnerUpdater()

	return app, nil
}

func (a *App) Run() error {
	defer a.g.Close()
	defer close(a.stopSpinnerCh) // Stop spinner goroutine
	defer func() {
		// Kill studio process if running
		if a.studioCmd != nil {
			a.studioCmd.Kill()
		}
	}()

	// 초기 포커스
	if len(a.focusOrder) > 0 {
		if panel, ok := a.panels[a.focusOrder[0]]; ok {
			panel.OnFocus()
		}
	}

	if err := a.g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (a *App) RegisterPanel(panel Panel) {
	a.panels[panel.ID()] = panel
}

func (a *App) GetGui() *gocui.Gui {
	return a.g
}

// OpenModal opens a modal and saves current focus state
func (a *App) OpenModal(modal Modal) {
	// Save current focus
	a.savedFocus = a.currentFocus

	// Blur all panels
	for _, id := range a.focusOrder {
		if panel, ok := a.panels[id]; ok {
			panel.OnBlur()
		}
	}

	a.activeModal = modal
}

// CloseModal closes the active modal and restores focus
func (a *App) CloseModal() {
	if a.activeModal != nil {
		a.activeModal.OnClose()
		a.activeModal = nil
	}

	// Restore focus
	if a.savedFocus >= 0 && a.savedFocus < len(a.focusOrder) {
		if panel, ok := a.panels[a.focusOrder[a.savedFocus]]; ok {
			panel.OnFocus()
		}
	}
}

// HasActiveModal returns true if a modal is currently active
func (a *App) HasActiveModal() bool {
	return a.activeModal != nil
}

// GetCurrentPanel returns the currently focused panel
func (a *App) GetCurrentPanel() Panel {
	if a.currentFocus >= 0 && a.currentFocus < len(a.focusOrder) {
		return a.panels[a.focusOrder[a.currentFocus]]
	}
	return nil
}

// tryStartCommand attempts to start a command execution
// Returns true if command can start, false if another command is already running
func (a *App) tryStartCommand(commandName string) bool {
	// CompareAndSwap atomically: if false, set to true and return true
	// if already true, return false
	if a.commandRunning.CompareAndSwap(false, true) {
		a.runningCommandName.Store(commandName)
		return true
	}
	return false
}

// finishCommand marks command execution as complete
func (a *App) finishCommand() {
	a.runningCommandName.Store("")
	a.commandRunning.Store(false)
	a.spinnerFrame.Store(0) // Reset spinner to first frame
}

// logCommandBlocked logs a message when command execution is blocked
func (a *App) logCommandBlocked(commandName string) {
	a.g.Update(func(g *gocui.Gui) error {
		if outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
			runningTask := ""
			if val := a.runningCommandName.Load(); val != nil {
				runningTask = val.(string)
			}

			message := fmt.Sprintf("Cannot execute '%s'", commandName)
			if runningTask != "" {
				message += fmt.Sprintf(" ('%s' is currently running)", runningTask)
			}

			outputPanel.LogActionRed("Operation Blocked", message)
		}
		return nil
	})
}

// startSpinnerUpdater starts a background goroutine that updates the spinner frame
func (a *App) startSpinnerUpdater() {
	go func() {
		ticker := time.NewTicker(spinnerTickInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Only update if command is running
				if a.commandRunning.Load() {
					// Increment frame index (0-3, wrapping around)
					currentFrame := a.spinnerFrame.Load()
					nextFrame := (currentFrame + 1) % uint32(len(spinnerFrames))
					a.spinnerFrame.Store(nextFrame)

					// Trigger UI update (thread-safe)
					a.g.Update(func(g *gocui.Gui) error {
						// StatusBar will be redrawn by layout manager
						return nil
					})
				}
			case <-a.stopSpinnerCh:
				return
			}
		}
	}()
}

// handlePanelClick handles mouse click on a panel to switch focus
func (a *App) handlePanelClick(viewID string) error {
	// Ignore if modal is active
	if a.HasActiveModal() {
		return nil
	}

	// Find the index of the clicked panel in focus order
	targetIndex := -1
	for i, id := range a.focusOrder {
		if id == viewID {
			targetIndex = i
			break
		}
	}

	// If panel not found or already focused, do nothing
	if targetIndex == -1 || targetIndex == a.currentFocus {
		return nil
	}

	// Blur current panel
	if panel, ok := a.panels[a.focusOrder[a.currentFocus]]; ok {
		panel.OnBlur()
	}

	// Update focus index
	a.currentFocus = targetIndex

	// Focus new panel
	if panel, ok := a.panels[a.focusOrder[a.currentFocus]]; ok {
		panel.OnFocus()
	}

	return nil
}

// registerMouseClickForFocus registers a mouse click handler to switch focus
func (a *App) registerMouseClickForFocus(viewID string) {
	a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
		ViewName: viewID,
		Key:      gocui.MouseLeft,
		Modifier: gocui.ModNone,
		Handler: func(opts gocui.ViewMouseBindingOpts) error {
			return a.handlePanelClick(viewID)
		},
	})
}

// RegisterMouseBindings registers mouse click handlers for all panels
func (a *App) RegisterMouseBindings() {
	// Register click handlers for all panels except MigrationsPanel and DetailsPanel
	for _, viewID := range a.focusOrder {
		if viewID != ViewMigrations && viewID != ViewDetails {
			a.registerMouseClickForFocus(viewID)
		}
	}

	// Register special bindings for MigrationsPanel
	if migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel); ok {
		migrationsPanel.SetApp(a)

		// Tab click binding
		a.g.SetTabClickBinding(ViewMigrations, func(tabIndex int) error {
			return migrationsPanel.handleTabClick(tabIndex)
		})

		// List item click binding
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewMigrations,
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				return migrationsPanel.handleListClick(opts.Y)
			},
		})
	}

	// Register special bindings for DetailsPanel
	if detailsPanel, ok := a.panels[ViewDetails].(*DetailsPanel); ok {
		// Tab click binding
		a.g.SetTabClickBinding(ViewDetails, func(tabIndex int) error {
			return detailsPanel.handleTabClick(tabIndex)
		})

		// Panel focus click binding (for content area)
		a.registerMouseClickForFocus(ViewDetails)
	}

	// Register mouse wheel bindings for all panels
	a.registerMouseWheelBindings()
}

// registerMouseWheelBindings registers mouse wheel handlers for all panels
func (a *App) registerMouseWheelBindings() {
	// Workspace panel
	if workspacePanel, ok := a.panels[ViewWorkspace].(*WorkspacePanel); ok {
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewWorkspace,
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				workspacePanel.ScrollUpByWheel()
				return nil
			},
		})
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewWorkspace,
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				workspacePanel.ScrollDownByWheel()
				return nil
			},
		})
	}

	// Migrations panel
	if migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel); ok {
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewMigrations,
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				migrationsPanel.ScrollUpByWheel()
				return nil
			},
		})
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewMigrations,
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				migrationsPanel.ScrollDownByWheel()
				return nil
			},
		})
	}

	// Details panel
	if detailsPanel, ok := a.panels[ViewDetails].(*DetailsPanel); ok {
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewDetails,
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				detailsPanel.ScrollUpByWheel()
				return nil
			},
		})
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewDetails,
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				detailsPanel.ScrollDownByWheel()
				return nil
			},
		})
	}

	// Output panel
	if outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewOutputs,
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				outputPanel.ScrollUpByWheel()
				return nil
			},
		})
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewOutputs,
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				outputPanel.ScrollDownByWheel()
				return nil
			},
		})
	}
}

// RefreshAll refreshes all panels asynchronously
func (a *App) RefreshAll(onComplete ...func()) bool {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Refresh All") {
		a.logCommandBlocked("Refresh All")
		return false
	}

	// Run refresh in background to avoid blocking UI
	go func() {
		defer a.finishCommand() // Always mark command as complete

		// Refresh workspace panel
		if workspacePanel, ok := a.panels[ViewWorkspace].(*WorkspacePanel); ok {
			workspacePanel.Refresh()
		}

		// Refresh migrations panel
		if migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel); ok {
			migrationsPanel.Refresh()
		}

		// Refresh Details panel Action-Needed data
		if detailsPanel, ok := a.panels[ViewDetails].(*DetailsPanel); ok {
			detailsPanel.LoadActionNeededData()
		}

		// Update UI on main thread (thread-safe)
		a.g.Update(func(g *gocui.Gui) error {
			// Add refresh notification to output panel
			if outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
				outputPanel.LogAction("Refresh", "All panels have been refreshed")
			}

			// Execute callbacks
			for _, callback := range onComplete {
				callback()
			}
			return nil
		})
	}()

	return true
}

// TestModal opens a test modal (temporary for testing)
func (a *App) TestModal() {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		modal := NewMessageModal(a.g, "Error",
			"Failed to get working directory:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Run validation
	result, err := prisma.Validate(cwd)
	if err != nil {
		modal := NewMessageModal(a.g, "Validation Error",
			"Failed to run validation:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Show result
	if result.Valid {
		// Validation passed
		modal := NewMessageModal(a.g, "Schema Validation Passed",
			"Your Prisma schema is valid!",
		).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})
		a.OpenModal(modal)
	} else {
		// Validation failed - show errors
		lines := []string{"Schema validation failed with the following errors:"}
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				styledErr := Stylize(err, Style{FgColor: ColorRed, Bold: true})
				lines = append(lines, styledErr)
			}
		} else {
			styledOutput := Stylize(result.Output, Style{FgColor: ColorRed, Bold: true})
			lines = append(lines, styledOutput)
		}

		modal := NewMessageModal(a.g, "Schema Validation Failed", lines...).
			WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// TestInputModal opens a test input modal (temporary for testing)
func (a *App) TestInputModal() {
	modal := NewInputModal(a.g, "Enter migration name",
		func(input string) {
			// Close input modal
			a.CloseModal()

			// Show result in message modal
			resultModal := NewMessageModal(a.g, "Input Received",
				"You entered:",
				input,
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			a.OpenModal(resultModal)
		},
		func() {
			// Cancel - just close modal
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan}).
		WithRequired(true).
		OnValidationFail(func(reason string) {
			// Close input modal and show error modal
			a.CloseModal()

			errorModal := NewMessageModal(a.g, "Validation Failed",
				reason,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(errorModal)
		})

	a.OpenModal(modal)
}

// TestListModal opens a test list modal (temporary for testing)
func (a *App) TestListModal() {
	items := []ListModalItem{
		{
			Label:       "Create Migration",
			Description: "Create a new migration file.\n\nThis will:\n• Generate a new migration file in prisma/migrations\n• Include timestamp in the filename\n• Prompt for migration name",
			OnSelect: func() error {
				a.CloseModal()
				resultModal := NewMessageModal(a.g, "Action Selected",
					"You selected: Create Migration",
				).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
				a.OpenModal(resultModal)
				return nil
			},
		},
		{
			Label:       "Run Migrations",
			Description: "Apply pending migrations to the database.\n\nThis will:\n• Execute all pending migrations in order\n• Update _prisma_migrations table\n• May modify database schema",
			OnSelect: func() error {
				a.CloseModal()
				resultModal := NewMessageModal(a.g, "Action Selected",
					"You selected: Run Migrations",
				).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
				a.OpenModal(resultModal)
				return nil
			},
		},
		{
			Label:       "Reset Database",
			Description: "Reset the database to a clean state.\n\nWARNING: This will:\n• Drop all tables and data\n• Re-run all migrations from scratch\n• Cannot be undone",
			OnSelect: func() error {
				a.CloseModal()
				resultModal := NewMessageModal(a.g, "Action Selected",
					"You selected: Reset Database",
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(resultModal)
				return nil
			},
		},
		{
			Label:       "Validate Schema",
			Description: "Validate the Prisma schema file.\n\nThis will:\n• Check for syntax errors\n• Verify model relationships\n• Validate field types\n• Report any issues",
			OnSelect: func() error {
				a.CloseModal()
				resultModal := NewMessageModal(a.g, "Action Selected",
					"You selected: Validate Schema",
				).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
				a.OpenModal(resultModal)
				return nil
			},
		},
	}

	modal := NewListModal(a.g, "Select Action", items,
		func() {
			// Cancel - just close modal
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	a.OpenModal(modal)
}

// TestConfirmModal opens a test confirm modal (temporary for testing)
func (a *App) TestConfirmModal() {
	modal := NewConfirmModal(a.g, "Confirm Action",
		"Are you sure you want to proceed with this action? This cannot be undone.",
		func() {
			// Yes callback - close confirm modal and show result
			a.CloseModal()
			resultModal := NewMessageModal(a.g, "Confirmed",
				"You clicked Yes!",
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			a.OpenModal(resultModal)
		},
		func() {
			// No callback - close confirm modal and show result
			a.CloseModal()
			resultModal := NewMessageModal(a.g, "Cancelled",
				"You clicked No!",
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(resultModal)
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})

	a.OpenModal(modal)
}

// TestPing tests network connectivity by pinging google.com
func (a *App) TestPing() {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Network Test") {
		a.logCommandBlocked("Network Test")
		return
	}

	outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Log action start
	outputPanel.LogAction("Network Test", "Pinging google.com...")

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build ping command (4 pings)
	pingCmd := builder.New("ping", "-c", "4", "google.com").
		StreamOutput().
		OnStdout(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnStderr(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					out.AppendOutput("  [ERROR] " + line)
				}
				return nil
			})
		}).
		OnComplete(func(exitCode int) {
			defer a.finishCommand() // Mark command as complete

			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					if exitCode == 0 {
						out.LogAction("Network Test Complete", "Ping successful")
					} else {
						out.LogAction("Network Test Failed", fmt.Sprintf("Ping failed with exit code: %d", exitCode))
					}
				}
				return nil
			})
		}).
		OnError(func(err error) {
			defer a.finishCommand() // Mark command as complete even on error

			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					out.LogAction("Network Test Error", err.Error())
				}
				return nil
			})
		})

	// Run async to avoid blocking UI
	if err := pingCmd.RunAsync(); err != nil {
		a.finishCommand() // Clean up if command fails to start
		outputPanel.LogAction("Network Test Error", "Failed to start ping: "+err.Error())
	}
}
