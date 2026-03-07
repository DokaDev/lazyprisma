package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/jesseduffield/gocui"
)

// MigrateDeploy runs npx prisma migrate deploy
func (a *App) MigrateDeploy() {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Migrate Deploy") {
		a.logCommandBlocked("Migrate Deploy")
		return
	}

	// Run everything in background to avoid blocking UI during refresh/checks
	go func() {
		// 1. Refresh first to ensure DB connection is current
		a.refreshPanels()

		// 2. Check DB connection
		migrationsPanel, ok := a.panels[ViewMigrations].(*context.MigrationsContext)
		if !ok {
			a.finishCommand()
			a.g.Update(func(g *gocui.Gui) error {
				modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleError,
					a.Tr.ErrorFailedAccessMigrationsPanel,
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return nil
			})
			return
		}

		// Check if DB is connected
		if !migrationsPanel.IsDBConnected() {
			a.finishCommand()
			a.g.Update(func(g *gocui.Gui) error {
				modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleDBConnectionRequired,
					a.Tr.ErrorNoDBConnectionDetected,
					a.Tr.ErrorEnsureDBAccessible,
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return nil
			})
			return
		}

		outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext)
		if !ok {
			a.finishCommand() // Clean up if panel not found
			return
		}

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			a.finishCommand()
			a.g.Update(func(g *gocui.Gui) error {
				outputPanel.LogAction(a.Tr.LogActionMigrateDeployFailed, a.Tr.ErrorFailedGetWorkingDir+" "+err.Error())
				modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateDeployError,
					a.Tr.ErrorFailedGetWorkingDir,
					err.Error(),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return nil
			})
			return
		}

		// Log action start
		a.g.Update(func(g *gocui.Gui) error {
			outputPanel.LogAction(a.Tr.LogActionMigrateDeploy, a.Tr.LogMsgRunningMigrateDeploy)
			return nil
		})

		// Create command builder
		builder := commands.NewCommandBuilder(commands.NewPlatform())

		// Build prisma migrate deploy command
		deployCmd := builder.New("npx", "prisma", "migrate", "deploy").
			WithWorkingDir(cwd).
			StreamOutput().
			OnStdout(func(line string) {
				// Update UI on main thread
				a.g.Update(func(g *gocui.Gui) error {
					if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
						out.AppendOutput("  " + line)
					}
					return nil
				})
			}).
			OnStderr(func(line string) {
				// Update UI on main thread
				a.g.Update(func(g *gocui.Gui) error {
					if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
						out.AppendOutput("  " + line)
					}
					return nil
				})
			}).
			OnComplete(func(exitCode int) {
				// Update UI on main thread
				a.g.Update(func(g *gocui.Gui) error {
					a.finishCommand() // Finish command
					if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
						if exitCode == 0 {
							out.LogAction(a.Tr.LogActionMigrateDeployComplete, a.Tr.LogMsgMigrationsAppliedSuccess)
							// Refresh all panels to show updated migration status
							a.RefreshAll()
							// Show success modal
							modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateDeploySuccess,
								a.Tr.ModalMsgMigrationsAppliedSuccess,
							).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
							a.OpenModal(modal)
						} else {
							out.LogAction(a.Tr.LogActionMigrateDeployFailed, fmt.Sprintf(a.Tr.LogMsgMigrateDeployFailedCode, exitCode))
							// Refresh even on failure - DB state may have changed
							a.RefreshAll()
							modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateDeployFailed,
								fmt.Sprintf(a.Tr.ModalMsgMigrateDeployFailedWithCode, exitCode),
								a.Tr.ModalMsgCheckOutputPanel,
							).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
							a.OpenModal(modal)
						}
					}
					return nil
				})
			}).
			OnError(func(err error) {
				// Update UI on main thread
				a.g.Update(func(g *gocui.Gui) error {
					a.finishCommand() // Finish command
					if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
						out.LogAction(a.Tr.LogActionMigrateDeployFailed, err.Error())
						modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateDeployError,
							a.Tr.ModalMsgFailedRunMigrateDeploy,
							err.Error(),
						).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
						a.OpenModal(modal)
					}
					return nil
				})
			})

		// Run async to avoid blocking UI (spinner will show automatically)
		if err := deployCmd.RunAsync(); err != nil {
			a.finishCommand() // Clean up if command fails to start
			a.g.Update(func(g *gocui.Gui) error {
				outputPanel.LogAction(a.Tr.LogActionMigrateDeployFailed, a.Tr.ModalMsgFailedStartMigrateDeploy+" "+err.Error())
				modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateDeployError,
					a.Tr.ModalMsgFailedStartMigrateDeploy,
					err.Error(),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return nil
			})
		}
	}()
}

// MigrateDev opens a list modal to choose migration type
func (a *App) MigrateDev() {
	items := []ListModalItem{
		{
			Label:       a.Tr.ListItemSchemaDiffMigration,
			Description: a.Tr.ListItemDescSchemaDiffMigration,
			OnSelect: func() error {
				a.CloseModal()
				a.SchemaDiffMigration()
				return nil
			},
		},
		{
			Label:       a.Tr.ListItemManualMigration,
			Description: a.Tr.ListItemDescManualMigration,
			OnSelect: func() error {
				a.CloseModal()
				a.showManualMigrationInput()
				return nil
			},
		},
	}

	modal := NewListModal(a.g, a.Tr, a.Tr.ModalTitleMigrateDev, items,
		func() {
			// Cancel - just close modal
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	a.OpenModal(modal)
}

// executeCreateMigration runs npx prisma migrate dev --name <name> --create-only
func (a *App) executeCreateMigration(migrationName string) {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Create Migration") {
		a.logCommandBlocked("Create Migration")
		return
	}

	outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction(a.Tr.LogActionMigrationError, a.Tr.ErrorFailedGetWorkingDir+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrationError,
			a.Tr.ErrorFailedGetWorkingDir,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	outputPanel.LogAction(a.Tr.LogActionMigrateDev, fmt.Sprintf(a.Tr.LogMsgCreatingMigration, migrationName))

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma migrate dev --create-only command
	// Note: --create-only flag creates the migration without applying it to the database
	createCmd := builder.New("npx", "prisma", "migrate", "dev", "--name", migrationName, "--create-only").
		WithWorkingDir(cwd).
		StreamOutput().
		OnStdout(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnStderr(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnComplete(func(exitCode int) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				a.finishCommand() // Finish command
				// Refresh all panels to show the new migration
				a.RefreshAll()

				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					if exitCode == 0 {
						out.LogAction(a.Tr.LogActionMigrateComplete, a.Tr.LogMsgMigrationCreatedSuccess)
						// Show success modal
						modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrationCreated,
							fmt.Sprintf(a.Tr.ModalMsgMigrationCreatedSuccess, migrationName),
							a.Tr.ModalMsgMigrationCreatedDetail,
						).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
						a.OpenModal(modal)
					} else {
						out.LogAction(a.Tr.LogActionMigrateFailed, fmt.Sprintf(a.Tr.LogMsgMigrationCreationFailedCode, exitCode))
						modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrationFailed,
							fmt.Sprintf(a.Tr.ModalMsgMigrationFailedWithCode, exitCode),
							a.Tr.ModalMsgCheckOutputPanel,
						).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
						a.OpenModal(modal)
					}
				}
				return nil
			})
		}).
		OnError(func(err error) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				a.finishCommand() // Finish command
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.LogAction(a.Tr.LogActionMigrationError, err.Error())
					modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrationError,
						a.Tr.ModalMsgFailedRunMigrateDeploy,
						err.Error(),
					).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
					a.OpenModal(modal)
				}
				return nil
			})
		})

	// Run async to avoid blocking UI (spinner will show automatically)
	if err := createCmd.RunAsync(); err != nil {
		a.finishCommand() // Clean up if command fails to start
		outputPanel.LogAction(a.Tr.LogActionMigrationError, a.Tr.ModalMsgFailedStartMigrateDeploy+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrationError,
			a.Tr.ModalMsgFailedStartMigrateDeploy,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// SchemaDiffMigration performs schema diff-based migration with validation checks
func (a *App) SchemaDiffMigration() {
	// 1. Refresh first (with callback to ensure data is loaded before checking)
	started := a.RefreshAll(func() {
		// 2. Check DB connection
		migrationsPanel, ok := a.panels[ViewMigrations].(*context.MigrationsContext)
		if !ok {
			modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleError,
				a.Tr.ErrorFailedAccessMigrationsPanel,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return
		}

		// Check if DB is connected
		if !migrationsPanel.IsDBConnected() {
			modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleDBConnectionRequired,
				a.Tr.ErrorNoDBConnectionDetected,
				a.Tr.ErrorEnsureDBAccessible,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return
		}

		// 3. Check for DB-Only migrations
		if len(migrationsPanel.GetCategory().DBOnly) > 0 {
			modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleDBOnlyMigrationsDetected,
				a.Tr.ModalMsgCannotCreateWithDBOnly,
				a.Tr.ModalMsgResolveDBOnlyFirst,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return
		}

		// 4. Check for Checksum Mismatch
		for _, m := range migrationsPanel.GetCategory().Local {
			if m.ChecksumMismatch {
				modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleChecksumMismatchDetected,
					a.Tr.ModalMsgCannotCreateWithMismatch,
					fmt.Sprintf(a.Tr.ModalMsgMigrationModifiedLocally, m.Name),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return
			}
		}

		// 5. Check for Pending migrations
		if len(migrationsPanel.GetCategory().Pending) > 0 {
			// Check if any pending migration is empty
			for _, m := range migrationsPanel.GetCategory().Pending {
				if m.IsEmpty {
					modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleEmptyPendingDetected,
						a.Tr.ModalMsgCannotCreateWithEmpty,
						fmt.Sprintf(a.Tr.ModalMsgMigrationPendingEmpty, m.Name),
						a.Tr.ModalMsgDeleteOrAddContent,
					).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
					a.OpenModal(modal)
					return
				}
			}

			// Show confirmation modal for normal pending migrations
			modal := NewConfirmModal(a.g, a.Tr, a.Tr.ModalTitlePendingMigrationsDetected,
				a.Tr.ModalMsgPendingMigrationsWarning,
				func() {
					// Yes - proceed with migration name input
					a.CloseModal()
					a.showMigrationNameInput()
				},
				func() {
					// No - cancel
					a.CloseModal()
				},
			).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
			a.OpenModal(modal)
			return
		}

		// All checks passed - show migration name input
		a.showMigrationNameInput()
	})

	if !started {
		// If refresh failed to start (e.g., another command running), show error
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleOperationBlocked,
			a.Tr.ModalMsgAnotherOperationRunning,
			a.Tr.ModalMsgWaitComplete,
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// createManualMigration creates a manual migration folder and file
func (a *App) createManualMigration(migrationName string) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleError,
			a.Tr.ErrorFailedGetWorkingDir,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Generate timestamp (YYYYMMDDHHmmss format) in UTC to match Prisma CLI behavior
	timestamp := time.Now().UTC().Format("20060102150405")
	folderName := fmt.Sprintf("%s_%s", timestamp, migrationName)

	// Migration folder path (prisma/migrations/{timestamp}_{name})
	migrationsDir := fmt.Sprintf("%s/prisma/migrations", cwd)
	migrationFolder := fmt.Sprintf("%s/%s", migrationsDir, folderName)

	// Create migration folder
	if err := os.MkdirAll(migrationFolder, 0755); err != nil {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleError,
			a.Tr.ModalMsgFailedCreateFolder,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Create migration.sql file with initial comment
	migrationFile := fmt.Sprintf("%s/migration.sql", migrationFolder)
	initialContent := "-- This migration was manually created via lazyprisma\n\n"

	if err := os.WriteFile(migrationFile, []byte(initialContent), 0644); err != nil {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleError,
			a.Tr.ModalMsgFailedWriteMigrationFile,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Success - show result and refresh
	a.RefreshAll()

	modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrationCreated,
		fmt.Sprintf(a.Tr.ModalMsgManualMigrationCreated, folderName),
		fmt.Sprintf(a.Tr.ModalMsgManualMigrationLocation, migrationFolder),
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	a.OpenModal(modal)
}

// showMigrationNameInput shows input modal for migration name
func (a *App) showMigrationNameInput() {
	modal := NewInputModal(a.g, a.Tr, a.Tr.ModalTitleEnterMigrationName,
		func(input string) {
			// Replace spaces with underscores
			migrationName := strings.ReplaceAll(strings.TrimSpace(input), " ", "_")

			// Close input modal
			a.CloseModal()

			// Execute actual migration creation
			a.executeCreateMigration(migrationName)
		},
		func() {
			// Cancel - just close modal
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan}).
		WithSubtitle(a.Tr.ModalMsgSpacesReplaced).
		WithRequired(true).
		OnValidationFail(func(reason string) {
			// Validation failed - show error
			a.CloseModal()
			errorModal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleValidationFailed,
				reason,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(errorModal)
		})

	a.OpenModal(modal)
}

// showManualMigrationInput shows input modal for manual migration name
func (a *App) showManualMigrationInput() {
	modal := NewInputModal(a.g, a.Tr, a.Tr.ModalTitleEnterMigrationName,
		func(input string) {
			// Replace spaces with underscores
			migrationName := strings.ReplaceAll(strings.TrimSpace(input), " ", "_")

			// Close input modal
			a.CloseModal()

			// Create manual migration
			a.createManualMigration(migrationName)
		},
		func() {
			// Cancel - just close modal
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan}).
		WithSubtitle(a.Tr.ModalMsgSpacesReplaced).
		WithRequired(true).
		OnValidationFail(func(reason string) {
			// Validation failed - show error
			a.CloseModal()
			errorModal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleValidationFailed,
				reason,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(errorModal)
		})

	a.OpenModal(modal)
}

// MigrateResolve resolves a failed migration
func (a *App) MigrateResolve() {
	// Get migrations panel
	migrationsPanel, ok := a.panels[ViewMigrations].(*context.MigrationsContext)
	if !ok {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleError,
			a.Tr.ErrorFailedAccessMigrationsPanel,
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Get selected migration
	selectedMigration := migrationsPanel.GetSelectedMigration()
	if selectedMigration == nil {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleNoMigrationSelected,
			a.Tr.ModalMsgSelectMigrationResolve,
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		a.OpenModal(modal)
		return
	}

	// Check if migration is failed (only In-Transaction migrations can be resolved)
	if !selectedMigration.IsFailed {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleCannotResolveMigration,
			a.Tr.ModalMsgOnlyInTransactionResolve,
			fmt.Sprintf(a.Tr.ModalMsgMigrationNotFailed, selectedMigration.Name),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Show ListModal with resolve options
	migrationName := selectedMigration.Name

	items := []ListModalItem{
		{
			Label:       a.Tr.ListItemMarkApplied,
			Description: a.Tr.ListItemDescMarkApplied,
			OnSelect: func() error {
				a.CloseModal()
				a.executeResolve(migrationName, "applied")
				return nil
			},
		},
		{
			Label:       a.Tr.ListItemMarkRolledBack,
			Description: a.Tr.ListItemDescMarkRolledBack,
			OnSelect: func() error {
				a.CloseModal()
				a.executeResolve(migrationName, "rolled-back")
				return nil
			},
		},
	}

	modal := NewListModal(a.g, a.Tr, fmt.Sprintf(a.Tr.ModalTitleResolveMigration, migrationName), items,
		func() { a.CloseModal() },
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	a.OpenModal(modal)
}

// executeResolve runs npx prisma migrate resolve with the specified action
func (a *App) executeResolve(migrationName string, action string) {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Migrate Resolve") {
		a.logCommandBlocked("Migrate Resolve")
		return
	}

	outputPanel, ok := a.panels[ViewOutputs].(*context.OutputContext)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction(a.Tr.LogActionMigrateResolveError, a.Tr.ErrorFailedGetWorkingDir+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateResolveError,
			a.Tr.ErrorFailedGetWorkingDir,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	actionLabel := "applied"
	if action == "rolled-back" {
		actionLabel = "rolled back"
	}
	outputPanel.LogAction(a.Tr.LogActionMigrateResolve, fmt.Sprintf(a.Tr.LogMsgMarkingMigration, actionLabel, migrationName))

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma migrate resolve command
	resolveCmd := builder.New("npx", "prisma", "migrate", "resolve", "--"+action, migrationName).
		WithWorkingDir(cwd).
		StreamOutput().
		OnStdout(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnStderr(func(line string) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnComplete(func(exitCode int) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				a.finishCommand() // Finish command
				// Refresh all panels to show updated migration status
				a.RefreshAll()
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					if exitCode == 0 {
						out.LogAction(a.Tr.LogActionMigrateResolveComplete, fmt.Sprintf(a.Tr.LogMsgMigrationMarked, actionLabel))
						// Show success modal
						modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateResolveSuccess,
							fmt.Sprintf(a.Tr.ModalMsgMigrationMarkedSuccess, actionLabel),
						).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
						a.OpenModal(modal)
					} else {
						out.LogAction(a.Tr.LogActionMigrateResolveFailed, fmt.Sprintf(a.Tr.LogMsgMigrateResolveFailedCode, exitCode))
						modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateResolveFailed,
							fmt.Sprintf(a.Tr.ModalMsgMigrateResolveFailedWithCode, exitCode),
							a.Tr.ModalMsgCheckOutputPanel,
						).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
						a.OpenModal(modal)
					}
				}
				return nil
			})
		}).
		OnError(func(err error) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				a.finishCommand() // Finish command
				if out, ok := a.panels[ViewOutputs].(*context.OutputContext); ok {
					out.LogAction(a.Tr.LogActionMigrateResolveError, err.Error())
					modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateResolveError,
						a.Tr.ModalMsgFailedRunMigrateResolve,
						err.Error(),
					).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
					a.OpenModal(modal)
				}
				return nil
			})
		})

	// Run async to avoid blocking UI (spinner will show automatically)
	if err := resolveCmd.RunAsync(); err != nil {
		a.finishCommand() // Clean up if command fails to start
		outputPanel.LogAction(a.Tr.LogActionMigrateResolveError, a.Tr.ModalMsgFailedStartMigrateResolve+" "+err.Error())
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleMigrateResolveError,
			a.Tr.ModalMsgFailedStartMigrateResolve,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// DeleteMigration deletes a pending migration
func (a *App) DeleteMigration() {
	// Get migrations panel
	migrationsPanel, ok := a.panels[ViewMigrations].(*context.MigrationsContext)
	if !ok {
		return
	}

	// Get selected migration
	selected := migrationsPanel.GetSelectedMigration()
	if selected == nil {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleNoSelection,
			a.Tr.ModalMsgSelectMigrationDelete,
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		a.OpenModal(modal)
		return
	}

	// Validate: Can only delete if it exists locally
	if selected.Path == "" {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleCannotDelete,
			a.Tr.ModalMsgMigrationDBOnly,
			a.Tr.ModalMsgCannotDeleteNoLocalFile,
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Validate: Can only delete pending migrations (not applied to DB)
	// Exception: If DB is not connected, we assume it's safe to delete local files (user responsibility)
	if migrationsPanel.IsDBConnected() && selected.AppliedAt != nil {
		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleCannotDelete,
			a.Tr.ModalMsgMigrationAlreadyApplied,
			a.Tr.ModalMsgDeleteLocalInconsistency,
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Confirm deletion
	modal := NewConfirmModal(a.g, a.Tr, a.Tr.ModalTitleDeleteMigration,
		fmt.Sprintf(a.Tr.ModalMsgConfirmDeleteMigration, selected.Name),
		func() {
			a.CloseModal()
			a.executeDeleteMigration(selected.Path, selected.Name)
		},
		func() {
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
	a.OpenModal(modal)
}

// executeDeleteMigration performs the actual deletion
func (a *App) executeDeleteMigration(path, name string) {
	if err := os.RemoveAll(path); err != nil {
		outputPanel, _ := a.panels[ViewOutputs].(*context.OutputContext)
		if outputPanel != nil {
			outputPanel.LogActionRed(a.Tr.ModalTitleDeleteError, fmt.Sprintf(a.Tr.LogMsgFailedDeleteMigration, err.Error()))
		}

		modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleDeleteError,
			a.Tr.ModalMsgFailedDeleteFolder,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Success
	outputPanel, _ := a.panels[ViewOutputs].(*context.OutputContext)
	if outputPanel != nil {
		outputPanel.LogAction(a.Tr.LogActionDeleted, fmt.Sprintf(a.Tr.LogMsgMigrationDeleted, name))
	}

	// Refresh to update list
	a.RefreshAll()

	modal := NewMessageModal(a.g, a.Tr, a.Tr.ModalTitleDeleted,
		a.Tr.ModalMsgMigrationDeletedSuccess,
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	a.OpenModal(modal)
}
