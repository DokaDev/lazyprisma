package app

import (
	"fmt"
	"os"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
)

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
