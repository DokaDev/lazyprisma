package app

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/common"
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
)

const (
	spinnerTickInterval = 50 * time.Millisecond
)

type App struct {
	g            *gocui.Gui
	config       AppConfig
	Common       *common.Common
	Tr           *i18n.TranslationSet
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
	Language  string
}

func NewApp(config AppConfig) (*App, error) {
	g, err := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if err != nil {
		return nil, err
	}

	cmn := common.NewCommon(i18n.NewTranslationSet(config.Language))

	app := &App{
		g:             g,
		config:        config,
		Common:        cmn,
		Tr:            cmn.Tr,
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

	// Initial focus
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

// StatusBarState returns callbacks for the StatusBarContext to access App state.
func (a *App) StatusBarState() context.StatusBarState {
	return context.StatusBarState{
		IsCommandRunning: func() bool {
			return a.commandRunning.Load()
		},
		GetSpinnerFrame: func() uint32 {
			return a.spinnerFrame.Load()
		},
		IsStudioRunning: func() bool {
			return a.studioRunning
		},
		GetCommandName: func() string {
			if val := a.runningCommandName.Load(); val != nil {
				return val.(string)
			}
			return ""
		},
	}
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
		if outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
			runningTask := ""
			if val := a.runningCommandName.Load(); val != nil {
				runningTask = val.(string)
			}

			message := fmt.Sprintf(a.Tr.ErrorCannotExecuteCommand, commandName)
			if runningTask != "" {
				message += fmt.Sprintf(a.Tr.ErrorCommandCurrentlyRunning, runningTask)
			}

			outputPanel.LogActionRed(a.Tr.ErrorOperationBlocked, message)
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
					nextFrame := (currentFrame + 1) % context.SpinnerFrameCount()
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

// HandlePanelClick is the public wrapper for panel-click focus switching.
// It is used as a callback by contexts that manage their own mouse events.
func (a *App) HandlePanelClick(viewID string) {
	_ = a.handlePanelClick(viewID) // error intentionally ignored: click handler fallback
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

	// Register special bindings for MigrationsContext
	if migrationsCtx, ok := a.panels[ViewMigrations].(*context.MigrationsContext); ok {
		// Tab click binding
		a.g.SetTabClickBinding(ViewMigrations, func(tabIndex int) error {
			return migrationsCtx.HandleTabClick(tabIndex)
		})

		// List item click binding
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewMigrations,
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				return migrationsCtx.HandleListClick(opts.Y)
			},
		})
	}

	// Register special bindings for DetailsContext
	if detailsCtx, ok := a.panels[ViewDetails].(*context.DetailsContext); ok {
		// Tab click binding
		a.g.SetTabClickBinding(ViewDetails, func(tabIndex int) error {
			return detailsCtx.HandleTabClick(tabIndex)
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
	if workspaceCtx, ok := a.panels[ViewWorkspace].(*context.WorkspaceContext); ok {
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewWorkspace,
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				workspaceCtx.ScrollUpByWheel()
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
				workspaceCtx.ScrollDownByWheel()
				return nil
			},
		})
	}

	// Migrations context
	if migrationsCtx, ok := a.panels[ViewMigrations].(*context.MigrationsContext); ok {
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewMigrations,
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				migrationsCtx.ScrollUpByWheel()
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
				migrationsCtx.ScrollDownByWheel()
				return nil
			},
		})
	}

	// Details context
	if detailsCtx, ok := a.panels[ViewDetails].(*context.DetailsContext); ok {
		a.g.SetViewClickBinding(&gocui.ViewMouseBinding{
			ViewName: ViewDetails,
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if a.HasActiveModal() {
					return nil
				}
				detailsCtx.ScrollUpByWheel()
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
				detailsCtx.ScrollDownByWheel()
				return nil
			},
		})
	}

	// Output panel
	if outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
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

		a.refreshPanels()

		// Update UI on main thread (thread-safe)
		a.g.Update(func(g *gocui.Gui) error {
			// Add refresh notification to output panel
			if outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
				outputPanel.LogAction(a.Tr.ActionRefresh, a.Tr.SuccessAllPanelsRefreshed)
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

// refreshPanels refreshes all panels (blocking, internal)
func (a *App) refreshPanels() {
	// Refresh workspace panel
	if workspaceCtx, ok := a.panels[ViewWorkspace].(*context.WorkspaceContext); ok {
		workspaceCtx.Refresh()
	}

	// Refresh migrations context
	if migrationsCtx, ok := a.panels[ViewMigrations].(*context.MigrationsContext); ok {
		migrationsCtx.Refresh()

		// Wire action-needed data from migrations to details
		if detailsCtx, ok := a.panels[ViewDetails].(*context.DetailsContext); ok {
			// Collect action-needed migrations from Local category
			var actionNeeded []prisma.Migration
			for _, mig := range migrationsCtx.GetCategory().Local {
				if mig.IsEmpty || mig.ChecksumMismatch {
					actionNeeded = append(actionNeeded, mig)
				}
			}
			detailsCtx.SetActionNeededMigrations(actionNeeded)
			detailsCtx.LoadActionNeededData()
		}
	}
}
