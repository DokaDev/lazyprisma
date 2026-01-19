package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dokadev/lazyprisma/pkg/commands"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
)

// MigrateDeploy runs npx prisma migrate deploy
func (a *App) MigrateDeploy() {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Migrate Deploy") {
		a.logCommandBlocked("Migrate Deploy")
		return
	}

	// 1. Refresh first to ensure DB connection is current
	a.RefreshAll()

	// 2. Check DB connection
	migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel)
	if !ok {
		a.finishCommand()
		modal := NewMessageModal(a.g, "Error",
			"Failed to access migrations panel.",
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Check if DB is connected
	if !migrationsPanel.dbConnected {
		a.finishCommand()
		modal := NewMessageModal(a.g, "Database Connection Required",
			"No database connection detected.",
			"Please ensure your database is running and accessible.",
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction("Migrate Deploy Error", "Failed to get working directory: "+err.Error())
		modal := NewMessageModal(a.g, "Migrate Deploy Error",
			"Failed to get working directory:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	outputPanel.LogAction("Migrate Deploy", "Running prisma migrate deploy...")

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma migrate deploy command
	deployCmd := builder.New("npx", "prisma", "migrate", "deploy").
		WithWorkingDir(cwd).
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
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnComplete(func(exitCode int) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				a.finishCommand() // Finish command
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					if exitCode == 0 {
						out.LogAction("Migrate Deploy Complete", "Migrations applied successfully")
						// Refresh all panels to show updated migration status
						a.RefreshAll()
						// Show success modal
						modal := NewMessageModal(a.g, "Migrate Deploy Successful",
							"Migrations applied successfully!",
						).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
						a.OpenModal(modal)
					} else {
						out.LogAction("Migrate Deploy Failed", fmt.Sprintf("Migrate deploy failed with exit code: %d", exitCode))
						// Refresh even on failure - DB state may have changed
						a.RefreshAll()
						modal := NewMessageModal(a.g, "Migrate Deploy Failed",
							fmt.Sprintf("Prisma migrate deploy failed with exit code: %d", exitCode),
							"Check output panel for details.",
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
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					out.LogAction("Migrate Deploy Error", err.Error())
					modal := NewMessageModal(a.g, "Migrate Deploy Error",
						"Failed to run prisma migrate deploy:",
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
		outputPanel.LogAction("Migrate Deploy Error", "Failed to start migrate deploy: "+err.Error())
		modal := NewMessageModal(a.g, "Migrate Deploy Error",
			"Failed to start migrate deploy:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// MigrateDev opens a list modal to choose migration type
func (a *App) MigrateDev() {
	items := []ListModalItem{
		{
			Label:       "Schema diff-based migration",
			Description: "Create a migration from changes in Prisma schema, apply it to the database, trigger generators (e.g. Prisma Client)",
			OnSelect: func() error {
				a.CloseModal()
				a.SchemaDiffMigration()
				return nil
			},
		},
		{
			Label:       "Manual migration",
			Description: "This tool creates manual migrations for database changes that cannot be expressed through Prisma schema diff. It is used to explicitly record and version control database-specific logic such as triggers, functions, and DML operations that cannot be managed at the Prisma schema level.",
			OnSelect: func() error {
				a.CloseModal()
				a.showManualMigrationInput()
				return nil
			},
		},
	}

	modal := NewListModal(a.g, "Migrate Dev", items,
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

	outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction("Migration Error", "Failed to get working directory: "+err.Error())
		modal := NewMessageModal(a.g, "Migration Error",
			"Failed to get working directory:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	outputPanel.LogAction("Migrate Dev", fmt.Sprintf("Creating migration: %s", migrationName))

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

				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					if exitCode == 0 {
						out.LogAction("Migrate Complete", "Migration created successfully")
						// Show success modal
						modal := NewMessageModal(a.g, "Migration Created",
							fmt.Sprintf("Migration '%s' created successfully!", migrationName),
							"You can find it in the prisma/migrations directory.",
						).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
						a.OpenModal(modal)
					} else {
						out.LogAction("Migrate Failed", fmt.Sprintf("Migration creation failed with exit code: %d", exitCode))
						modal := NewMessageModal(a.g, "Migration Failed",
							fmt.Sprintf("Prisma migrate dev failed with exit code: %d", exitCode),
							"Check output panel for details.",
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
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					out.LogAction("Migration Error", err.Error())
					modal := NewMessageModal(a.g, "Migration Error",
						"Failed to run prisma migrate dev:",
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
		outputPanel.LogAction("Migration Error", "Failed to start migrate dev: "+err.Error())
		modal := NewMessageModal(a.g, "Migration Error",
			"Failed to start migrate dev:",
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
		migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel)
		if !ok {
			modal := NewMessageModal(a.g, "Error",
				"Failed to access migrations panel.",
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return
		}

		// Check if DB is connected
		if !migrationsPanel.dbConnected {
			modal := NewMessageModal(a.g, "Database Connection Required",
				"No database connection detected.",
				"Please ensure your database is running and accessible.",
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return
		}

		// 3. Check for DB-Only migrations
		if len(migrationsPanel.category.DBOnly) > 0 {
			modal := NewMessageModal(a.g, "DB-Only Migrations Detected",
				"Cannot create new migration whilst DB-Only migrations exist.",
				"Please resolve DB-Only migrations first.",
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(modal)
			return
		}

		// 4. Check for Checksum Mismatch
		for _, m := range migrationsPanel.category.Local {
			if m.ChecksumMismatch {
				modal := NewMessageModal(a.g, "Checksum Mismatch Detected",
					"Cannot create new migration whilst checksum mismatch exists.",
					fmt.Sprintf("Migration '%s' has been modified locally.", m.Name),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return
			}
		}

		// 5. Check for Pending migrations
		if len(migrationsPanel.category.Pending) > 0 {
			// Check if any pending migration is empty
			for _, m := range migrationsPanel.category.Pending {
				if m.IsEmpty {
					modal := NewMessageModal(a.g, "Empty Pending Migration Detected",
						"Cannot create new migration whilst empty pending migrations exist.",
						fmt.Sprintf("Migration '%s' is pending and empty.", m.Name),
						"Please delete it or add SQL content.",
					).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
					a.OpenModal(modal)
					return
				}
			}

			// Show confirmation modal for normal pending migrations
			modal := NewConfirmModal(a.g, "Pending Migrations Detected",
				"Prisma automatically applies pending migrations before creating new ones. This may cause unintended behaviour in the future. Do you wish to continue?",
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
		modal := NewMessageModal(a.g, "Operation Blocked",
			"Another operation is currently running.",
			"Please wait for it to complete.",
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// createManualMigration creates a manual migration folder and file
func (a *App) createManualMigration(migrationName string) {
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

	// Generate timestamp (YYYYMMDDHHmmss format) in UTC to match Prisma CLI behavior
	timestamp := time.Now().UTC().Format("20060102150405")
	folderName := fmt.Sprintf("%s_%s", timestamp, migrationName)

	// Migration folder path (prisma/migrations/{timestamp}_{name})
	migrationsDir := fmt.Sprintf("%s/prisma/migrations", cwd)
	migrationFolder := fmt.Sprintf("%s/%s", migrationsDir, folderName)

	// Create migration folder
	if err := os.MkdirAll(migrationFolder, 0755); err != nil {
		modal := NewMessageModal(a.g, "Error",
			"Failed to create migration folder:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Create migration.sql file with initial comment
	migrationFile := fmt.Sprintf("%s/migration.sql", migrationFolder)
	initialContent := "-- This migration was manually created via lazyprisma\n\n"

	if err := os.WriteFile(migrationFile, []byte(initialContent), 0644); err != nil {
		modal := NewMessageModal(a.g, "Error",
			"Failed to create migration.sql:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Success - show result and refresh
	a.RefreshAll()

	modal := NewMessageModal(a.g, "Manual Migration Created",
		fmt.Sprintf("Created: %s", folderName),
		fmt.Sprintf("Location: %s", migrationFolder),
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	a.OpenModal(modal)
}

// showMigrationNameInput shows input modal for migration name
func (a *App) showMigrationNameInput() {
	modal := NewInputModal(a.g, "Enter migration name",
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
		WithSubtitle("Spaces will be replaced with underscores").
		WithRequired(true).
		OnValidationFail(func(reason string) {
			// Validation failed - show error
			a.CloseModal()
			errorModal := NewMessageModal(a.g, "Validation Failed",
				reason,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(errorModal)
		})

	a.OpenModal(modal)
}

// showManualMigrationInput shows input modal for manual migration name
func (a *App) showManualMigrationInput() {
	modal := NewInputModal(a.g, "Enter migration name",
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
		WithSubtitle("Spaces will be replaced with underscores").
		WithRequired(true).
		OnValidationFail(func(reason string) {
			// Validation failed - show error
			a.CloseModal()
			errorModal := NewMessageModal(a.g, "Validation Failed",
				reason,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			a.OpenModal(errorModal)
		})

	a.OpenModal(modal)
}

// Generate runs prisma generate and shows result in modal
func (a *App) Generate() {
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Generate") {
		a.logCommandBlocked("Generate")
		return
	}

	outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction("Generate Error", "Failed to get working directory: "+err.Error())
		modal := NewMessageModal(a.g, "Generate Error",
			"Failed to get working directory:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	outputPanel.LogAction("Generate", "Running prisma generate...")

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma generate command
	generateCmd := builder.New("npx", "prisma", "generate").
		WithWorkingDir(cwd).
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
					out.AppendOutput("  " + line)
				}
				return nil
			})
		}).
		OnComplete(func(exitCode int) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					if exitCode == 0 {
						a.finishCommand() // Finish immediately on success
						out.LogAction("Generate Complete", "Prisma Client generated successfully")
						// Show success modal
						modal := NewMessageModal(a.g, "Generate Successful",
							"Prisma Client generated successfully!",
						).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
						a.OpenModal(modal)
					} else {
						// Failed - run validate to check schema (keep spinner running)
						out.LogAction("Generate Failed", "Checking schema for errors...")

						// Run validate in goroutine to not block UI updates
						go func() {
							validateResult, err := prisma.Validate(cwd)

							// Update UI on main thread after validate completes
							a.g.Update(func(g *gocui.Gui) error {
								a.finishCommand() // Finish after validate completes

								if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
									if err == nil && !validateResult.Valid {
										// Schema has validation errors - show them
										out.LogAction("Schema Validation Failed", fmt.Sprintf("Found %d schema errors", len(validateResult.Errors)))

										// Show validation errors in modal
										modal := NewMessageModal(a.g, "Schema Validation Failed",
											"Generate failed due to schema errors.",
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									} else {
										// Schema is valid but generate failed for other reasons
										out.LogAction("Generate Failed", fmt.Sprintf("Generate failed with exit code: %d", exitCode))
										modal := NewMessageModal(a.g, "Generate Failed",
											fmt.Sprintf("Prisma generate failed with exit code: %d", exitCode),
											"Schema is valid. Check output panel for details.",
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									}
								}
								return nil
							})
						}()
					}
				}
				return nil
			})
		}).
		OnError(func(err error) {
			// Update UI on main thread
			a.g.Update(func(g *gocui.Gui) error {
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					// Check if it's an exit status error (command ran but failed)
					if strings.Contains(err.Error(), "exit status") {
						// Failed - run validate to check schema (keep spinner running)
						out.LogAction("Generate Failed", "Checking schema for errors...")

						// Run validate in goroutine to not block UI updates
						go func() {
							validateResult, validateErr := prisma.Validate(cwd)

							// Update UI on main thread after validate completes
							a.g.Update(func(g *gocui.Gui) error {
								a.finishCommand() // Finish after validate completes

								if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
									if validateErr == nil && !validateResult.Valid {
										// Schema has validation errors - show them
										out.LogAction("Schema Validation Failed", fmt.Sprintf("Found %d schema errors", len(validateResult.Errors)))

										// Show validation errors in modal
										modal := NewMessageModal(a.g, "Schema Validation Failed",
											"Generate failed due to schema errors.",
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									} else {
										// Schema is valid but generate failed for other reasons
										out.LogAction("Generate Failed", err.Error())
										modal := NewMessageModal(a.g, "Generate Failed",
											"Prisma generate failed:",
											"Schema is valid. Check output panel for details.",
										).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
										a.OpenModal(modal)
									}
								}
								return nil
							})
						}()
					} else {
						// Other error (command couldn't start, etc.)
						a.finishCommand() // Finish immediately on startup error
						out.LogAction("Generate Error", err.Error())
						modal := NewMessageModal(a.g, "Generate Error",
							"Failed to run prisma generate:",
							err.Error(),
						).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
						a.OpenModal(modal)
					}
				}
				return nil
			})
		})

	// Run async to avoid blocking UI (spinner will show automatically)
	if err := generateCmd.RunAsync(); err != nil {
		a.finishCommand() // Clean up if command fails to start
		outputPanel.LogAction("Generate Error", "Failed to start generate: "+err.Error())
		modal := NewMessageModal(a.g, "Generate Error",
			"Failed to start generate:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// MigrateResolve resolves a failed migration
func (a *App) MigrateResolve() {
	// Get migrations panel
	migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel)
	if !ok {
		modal := NewMessageModal(a.g, "Error",
			"Failed to access migrations panel.",
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Get selected migration
	selectedMigration := migrationsPanel.GetSelectedMigration()
	if selectedMigration == nil {
		modal := NewMessageModal(a.g, "No Migration Selected",
			"Please select a migration to resolve.",
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		a.OpenModal(modal)
		return
	}

	// Check if migration is failed (only In-Transaction migrations can be resolved)
	if !selectedMigration.IsFailed {
		modal := NewMessageModal(a.g, "Cannot Resolve Migration",
			"Only migrations in 'In-Transaction' state can be resolved.",
			fmt.Sprintf("Migration '%s' is not in a failed state.", selectedMigration.Name),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Show ListModal with resolve options
	migrationName := selectedMigration.Name

	items := []ListModalItem{
		{
			Label:       "Mark as applied",
			Description: "Mark this migration as successfully applied to the database. Use this if you have manually fixed the issue and the migration changes are now present in the database.",
			OnSelect: func() error {
				a.CloseModal()
				a.executeResolve(migrationName, "applied")
				return nil
			},
		},
		{
			Label:       "Mark as rolled back",
			Description: "Mark this migration as rolled back (reverted from the database). Use this if you have manually reverted the changes and the migration is no longer applied to the database.",
			OnSelect: func() error {
				a.CloseModal()
				a.executeResolve(migrationName, "rolled-back")
				return nil
			},
		},
	}

	modal := NewListModal(a.g, "Resolve Migration: "+migrationName, items,
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

	outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel)
	if !ok {
		a.finishCommand() // Clean up if panel not found
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction("Migrate Resolve Error", "Failed to get working directory: "+err.Error())
		modal := NewMessageModal(a.g, "Migrate Resolve Error",
			"Failed to get working directory:",
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
	outputPanel.LogAction("Migrate Resolve", fmt.Sprintf("Marking migration as %s: %s", actionLabel, migrationName))

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma migrate resolve command
	resolveCmd := builder.New("npx", "prisma", "migrate", "resolve", "--"+action, migrationName).
		WithWorkingDir(cwd).
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
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					if exitCode == 0 {
						out.LogAction("Migrate Resolve Complete", fmt.Sprintf("Migration marked as %s successfully", actionLabel))
						// Show success modal
						modal := NewMessageModal(a.g, "Migrate Resolve Successful",
							fmt.Sprintf("Migration marked as %s successfully!", actionLabel),
						).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
						a.OpenModal(modal)
					} else {
						out.LogAction("Migrate Resolve Failed", fmt.Sprintf("Migrate resolve failed with exit code: %d", exitCode))
						modal := NewMessageModal(a.g, "Migrate Resolve Failed",
							fmt.Sprintf("Prisma migrate resolve failed with exit code: %d", exitCode),
							"Check output panel for details.",
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
				if out, ok := a.panels[ViewOutputs].(*OutputPanel); ok {
					out.LogAction("Migrate Resolve Error", err.Error())
					modal := NewMessageModal(a.g, "Migrate Resolve Error",
						"Failed to run prisma migrate resolve:",
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
		outputPanel.LogAction("Migrate Resolve Error", "Failed to start migrate resolve: "+err.Error())
		modal := NewMessageModal(a.g, "Migrate Resolve Error",
			"Failed to start migrate resolve:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
	}
}

// Studio toggles Prisma Studio
func (a *App) Studio() {
	outputPanel, ok := a.panels[ViewOutputs].(*OutputPanel)
	if !ok {
		return
	}

	// Check if Studio is already running
	if a.studioRunning {
		// Stop Studio
		if a.studioCmd != nil {
			if err := a.studioCmd.Kill(); err != nil {
				outputPanel.LogAction("Studio Error", "Failed to stop Prisma Studio: "+err.Error())
				modal := NewMessageModal(a.g, "Studio Error",
					"Failed to stop Prisma Studio:",
					err.Error(),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				a.OpenModal(modal)
				return
			}
			a.studioCmd = nil
		}
		a.studioRunning = false
		outputPanel.LogAction("Studio Stopped", "Prisma Studio has been stopped")

		// Clear subtitle
		outputPanel.SetSubtitle("")

		// Update UI
		a.g.Update(func(g *gocui.Gui) error {
			// Trigger redraw of status bar
			return nil
		})

		modal := NewMessageModal(a.g, "Studio Stopped",
			"Prisma Studio has been stopped.",
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		a.OpenModal(modal)
		return
	}

	// Start Studio
	// Try to start command - if another command is running, block
	if !a.tryStartCommand("Start Studio") {
		a.logCommandBlocked("Start Studio")
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		a.finishCommand()
		outputPanel.LogAction("Studio Error", "Failed to get working directory: "+err.Error())
		modal := NewMessageModal(a.g, "Studio Error",
			"Failed to get working directory:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Log action start
	outputPanel.LogAction("Studio", "Starting Prisma Studio...")

	// Create command builder
	builder := commands.NewCommandBuilder(commands.NewPlatform())

	// Build prisma studio command
	// Note: We don't use StreamOutput here because Studio is a long-running process
	// and we want to capture the command object to kill it later
	studioCmd := builder.New("npx", "prisma", "studio").
		WithWorkingDir(cwd)

	// Start async
	if err := studioCmd.RunAsync(); err != nil {
		a.finishCommand()
		outputPanel.LogAction("Studio Error", "Failed to start Prisma Studio: "+err.Error())
		modal := NewMessageModal(a.g, "Studio Error",
			"Failed to start Prisma Studio:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Wait a bit to ensure it started, then finish the "starting" command
	// The process continues running in background
	go func() {
		time.Sleep(2 * time.Second)
		a.g.Update(func(g *gocui.Gui) error {
			a.finishCommand() // Finish "starting" command
			a.studioRunning = true
			a.studioCmd = studioCmd // Save Command object

			outputPanel.LogAction("Studio Started", "Prisma Studio is running at http://localhost:5555")
			outputPanel.SetSubtitle("Prisma Studio listening on http://localhost:5555")

			// Show info modal
			modal := NewMessageModal(a.g, "Prisma Studio Started",
				"Prisma Studio is running at http://localhost:5555",
				"Press 'S' again to stop it.",
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			a.OpenModal(modal)
			return nil
		})
	}()
}

// DeleteMigration deletes a pending migration
func (a *App) DeleteMigration() {
	// Get migrations panel
	migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel)
	if !ok {
		return
	}

	// Get selected migration
	selected := migrationsPanel.GetSelectedMigration()
	if selected == nil {
		modal := NewMessageModal(a.g, "No Selection",
			"Please select a migration to delete.",
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		a.OpenModal(modal)
		return
	}

	// Validate: Can only delete if it exists locally
	if selected.Path == "" {
		modal := NewMessageModal(a.g, "Cannot Delete",
			"This migration exists only in the database (DB-Only).",
			"Cannot delete a migration that has no local file.",
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Validate: Can only delete pending migrations (not applied to DB)
	// Exception: If DB is not connected, we assume it's safe to delete local files (user responsibility)
	if migrationsPanel.dbConnected && selected.AppliedAt != nil {
		modal := NewMessageModal(a.g, "Cannot Delete",
			"This migration has already been applied to the database.",
			"Deleting it locally will cause inconsistency.",
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Confirm deletion
	modal := NewConfirmModal(a.g, "Delete Migration",
		fmt.Sprintf("Are you sure you want to delete this migration?\n\n%s\n\nThis action cannot be undone.", selected.Name),
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
		outputPanel, _ := a.panels[ViewOutputs].(*OutputPanel)
		if outputPanel != nil {
			outputPanel.LogActionRed("Delete Error", "Failed to delete migration: "+err.Error())
		}

		modal := NewMessageModal(a.g, "Delete Error",
			"Failed to delete migration folder:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Success
	outputPanel, _ := a.panels[ViewOutputs].(*OutputPanel)
	if outputPanel != nil {
		outputPanel.LogAction("Deleted", fmt.Sprintf("Migration '%s' deleted", name))
	}

	// Refresh to update list
	a.RefreshAll()

	modal := NewMessageModal(a.g, "Deleted",
		"Migration deleted successfully.",
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	a.OpenModal(modal)
}

// CopyMigrationInfo copies migration info to clipboard
func (a *App) CopyMigrationInfo() {
	// Get migrations panel
	migrationsPanel, ok := a.panels[ViewMigrations].(*MigrationsPanel)
	if !ok {
		return
	}

	// Get selected migration
	selected := migrationsPanel.GetSelectedMigration()
	if selected == nil {
		return
	}

	items := []ListModalItem{
		{
			Label:       "Copy Name",
			Description: selected.Name,
			OnSelect: func() error {
				a.CloseModal()
				a.copyTextToClipboard(selected.Name, "Migration Name")
				return nil
			},
		},
		{
			Label:       "Copy Path",
			Description: selected.Path,
			OnSelect: func() error {
				a.CloseModal()
				a.copyTextToClipboard(selected.Path, "Migration Path")
				return nil
			},
		},
	}

	// If it has a checksum, allow copying it
	if selected.Checksum != "" {
		items = append(items, ListModalItem{
			Label:       "Copy Checksum",
			Description: selected.Checksum,
			OnSelect: func() error {
				a.CloseModal()
				a.copyTextToClipboard(selected.Checksum, "Checksum")
				return nil
			},
		})
	}

	modal := NewListModal(a.g, "Copy to Clipboard", items,
		func() {
			a.CloseModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	a.OpenModal(modal)
}

func (a *App) copyTextToClipboard(text, label string) {
	if err := CopyToClipboard(text); err != nil {
		modal := NewMessageModal(a.g, "Clipboard Error",
			"Failed to copy to clipboard:",
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		a.OpenModal(modal)
		return
	}

	// Show toast/notification via modal for now
	// Ideally we would have a toast system
	modal := NewMessageModal(a.g, "Copied",
		fmt.Sprintf("%s copied to clipboard!", label),
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	a.OpenModal(modal)
}
