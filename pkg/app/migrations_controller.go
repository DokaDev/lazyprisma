package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dokadev/lazyprisma/pkg/gui/context"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/jesseduffield/gocui"
)

// MigrationsController handles migration-related operations.
type MigrationsController struct {
	c             types.IControllerHost
	g             *gocui.Gui
	migrationsCtx *context.MigrationsContext
	outputCtx     *context.OutputContext
	openModal     func(Modal)
	closeModal    func()
	runStreamCmd  func(AsyncCommandOpts) bool
}

// NewMigrationsController creates a new MigrationsController.
func NewMigrationsController(
	c types.IControllerHost,
	g *gocui.Gui,
	migrationsCtx *context.MigrationsContext,
	outputCtx *context.OutputContext,
	openModal func(Modal),
	closeModal func(),
	runStreamCmd func(AsyncCommandOpts) bool,
) *MigrationsController {
	return &MigrationsController{
		c:             c,
		g:             g,
		migrationsCtx: migrationsCtx,
		outputCtx:     outputCtx,
		openModal:     openModal,
		closeModal:    closeModal,
		runStreamCmd:  runStreamCmd,
	}
}

// MigrateDeploy runs npx prisma migrate deploy
func (mc *MigrationsController) MigrateDeploy() {
	tr := mc.c.GetTranslationSet()

	// Try to start command - if another command is running, block
	if !mc.c.TryStartCommand("Migrate Deploy") {
		mc.c.LogCommandBlocked("Migrate Deploy")
		return
	}

	// Run everything in background to avoid blocking UI during refresh/checks
	go func() {
		// 1. Refresh first to ensure DB connection is current
		mc.c.RefreshPanels()

		// 2. Check DB connection
		if !mc.migrationsCtx.IsDBConnected() {
			mc.c.FinishCommand()
			mc.c.OnUIThread(func() error {
				modal := NewMessageModal(mc.g, tr, tr.ModalTitleDBConnectionRequired,
					tr.ErrorNoDBConnectionDetected,
					tr.ErrorEnsureDBAccessible,
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				mc.openModal(modal)
				return nil
			})
			return
		}

		// Pre-flight checks passed -- run the streaming command
		mc.runStreamCmd(AsyncCommandOpts{
			Name:          "Migrate Deploy",
			SkipTryStart:  true, // already called above
			Args:          []string{"npx", "prisma", "migrate", "deploy"},
			LogAction:     tr.LogActionMigrateDeploy,
			LogDetail:     tr.LogMsgRunningMigrateDeploy,
			ErrorTitle:    tr.ModalTitleMigrateDeployError,
			ErrorStartMsg: tr.ModalMsgFailedStartMigrateDeploy,
			OnSuccess: func(out *context.OutputContext, cwd string) {
				mc.c.FinishCommand()
				out.LogAction(tr.LogActionMigrateDeployComplete, tr.LogMsgMigrationsAppliedSuccess)
				mc.c.RefreshAll()
				modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrateDeploySuccess,
					tr.ModalMsgMigrationsAppliedSuccess,
				).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
				mc.openModal(modal)
			},
			OnFailure: func(out *context.OutputContext, cwd string, exitCode int) {
				mc.c.FinishCommand()
				out.LogAction(tr.LogActionMigrateDeployFailed, fmt.Sprintf(tr.LogMsgMigrateDeployFailedCode, exitCode))
				mc.c.RefreshAll()
				modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrateDeployFailed,
					fmt.Sprintf(tr.ModalMsgMigrateDeployFailedWithCode, exitCode),
					tr.ModalMsgCheckOutputPanel,
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				mc.openModal(modal)
			},
			OnError: func(out *context.OutputContext, cwd string, err error) {
				mc.c.FinishCommand()
				out.LogAction(tr.LogActionMigrateDeployFailed, err.Error())
				modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrateDeployError,
					tr.ModalMsgFailedRunMigrateDeploy,
					err.Error(),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				mc.openModal(modal)
			},
		})
	}()
}

// MigrateDev opens a list modal to choose migration type
func (mc *MigrationsController) MigrateDev() {
	tr := mc.c.GetTranslationSet()

	items := []ListModalItem{
		{
			Label:       tr.ListItemSchemaDiffMigration,
			Description: tr.ListItemDescSchemaDiffMigration,
			OnSelect: func() error {
				mc.closeModal()
				mc.SchemaDiffMigration()
				return nil
			},
		},
		{
			Label:       tr.ListItemManualMigration,
			Description: tr.ListItemDescManualMigration,
			OnSelect: func() error {
				mc.closeModal()
				mc.showManualMigrationInput()
				return nil
			},
		},
	}

	modal := NewListModal(mc.g, tr, tr.ModalTitleMigrateDev, items,
		func() {
			mc.closeModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	mc.openModal(modal)
}

// executeCreateMigration runs npx prisma migrate dev --name <name> --create-only
func (mc *MigrationsController) executeCreateMigration(migrationName string) {
	tr := mc.c.GetTranslationSet()

	mc.runStreamCmd(AsyncCommandOpts{
		Name:          "Create Migration",
		Args:          []string{"npx", "prisma", "migrate", "dev", "--name", migrationName, "--create-only"},
		LogAction:     tr.LogActionMigrateDev,
		LogDetail:     fmt.Sprintf(tr.LogMsgCreatingMigration, migrationName),
		ErrorTitle:    tr.ModalTitleMigrationError,
		ErrorStartMsg: tr.ModalMsgFailedStartMigrateDeploy,
		OnSuccess: func(out *context.OutputContext, cwd string) {
			mc.c.FinishCommand()
			mc.c.RefreshAll()
			out.LogAction(tr.LogActionMigrateComplete, tr.LogMsgMigrationCreatedSuccess)
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrationCreated,
				fmt.Sprintf(tr.ModalMsgMigrationCreatedSuccess, migrationName),
				tr.ModalMsgMigrationCreatedDetail,
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			mc.openModal(modal)
		},
		OnFailure: func(out *context.OutputContext, cwd string, exitCode int) {
			mc.c.FinishCommand()
			mc.c.RefreshAll()
			out.LogAction(tr.LogActionMigrateFailed, fmt.Sprintf(tr.LogMsgMigrationCreationFailedCode, exitCode))
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrationFailed,
				fmt.Sprintf(tr.ModalMsgMigrationFailedWithCode, exitCode),
				tr.ModalMsgCheckOutputPanel,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(modal)
		},
		OnError: func(out *context.OutputContext, cwd string, err error) {
			mc.c.FinishCommand()
			out.LogAction(tr.LogActionMigrationError, err.Error())
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrationError,
				tr.ModalMsgFailedRunMigrateDeploy,
				err.Error(),
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(modal)
		},
	})
}

// SchemaDiffMigration performs schema diff-based migration with validation checks
func (mc *MigrationsController) SchemaDiffMigration() {
	tr := mc.c.GetTranslationSet()

	// 1. Refresh first (with callback to ensure data is loaded before checking)
	started := mc.c.RefreshAll(func() {
		// 2. Check DB connection
		if !mc.migrationsCtx.IsDBConnected() {
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleDBConnectionRequired,
				tr.ErrorNoDBConnectionDetected,
				tr.ErrorEnsureDBAccessible,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(modal)
			return
		}

		// 3. Check for DB-Only migrations
		if len(mc.migrationsCtx.GetCategory().DBOnly) > 0 {
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleDBOnlyMigrationsDetected,
				tr.ModalMsgCannotCreateWithDBOnly,
				tr.ModalMsgResolveDBOnlyFirst,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(modal)
			return
		}

		// 4. Check for Checksum Mismatch
		for _, m := range mc.migrationsCtx.GetCategory().Local {
			if m.ChecksumMismatch {
				modal := NewMessageModal(mc.g, tr, tr.ModalTitleChecksumMismatchDetected,
					tr.ModalMsgCannotCreateWithMismatch,
					fmt.Sprintf(tr.ModalMsgMigrationModifiedLocally, m.Name),
				).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
				mc.openModal(modal)
				return
			}
		}

		// 5. Check for Pending migrations
		if len(mc.migrationsCtx.GetCategory().Pending) > 0 {
			// Check if any pending migration is empty
			for _, m := range mc.migrationsCtx.GetCategory().Pending {
				if m.IsEmpty {
					modal := NewMessageModal(mc.g, tr, tr.ModalTitleEmptyPendingDetected,
						tr.ModalMsgCannotCreateWithEmpty,
						fmt.Sprintf(tr.ModalMsgMigrationPendingEmpty, m.Name),
						tr.ModalMsgDeleteOrAddContent,
					).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
					mc.openModal(modal)
					return
				}
			}

			// Show confirmation modal for normal pending migrations
			modal := NewConfirmModal(mc.g, tr, tr.ModalTitlePendingMigrationsDetected,
				tr.ModalMsgPendingMigrationsWarning,
				func() {
					// Yes - proceed with migration name input
					mc.closeModal()
					mc.showMigrationNameInput()
				},
				func() {
					// No - cancel
					mc.closeModal()
				},
			).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
			mc.openModal(modal)
			return
		}

		// All checks passed - show migration name input
		mc.showMigrationNameInput()
	})

	if !started {
		// If refresh failed to start (e.g., another command running), show error
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleOperationBlocked,
			tr.ModalMsgAnotherOperationRunning,
			tr.ModalMsgWaitComplete,
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
	}
}

// createManualMigration creates a manual migration folder and file
func (mc *MigrationsController) createManualMigration(migrationName string) {
	tr := mc.c.GetTranslationSet()

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleError,
			tr.ErrorFailedGetWorkingDir,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
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
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleError,
			tr.ModalMsgFailedCreateFolder,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
		return
	}

	// Create migration.sql file with initial comment
	migrationFile := fmt.Sprintf("%s/migration.sql", migrationFolder)
	initialContent := "-- This migration was manually created via lazyprisma\n\n"

	if err := os.WriteFile(migrationFile, []byte(initialContent), 0644); err != nil {
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleError,
			tr.ModalMsgFailedWriteMigrationFile,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
		return
	}

	// Success - show result and refresh
	mc.c.RefreshAll()

	modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrationCreated,
		fmt.Sprintf(tr.ModalMsgManualMigrationCreated, folderName),
		fmt.Sprintf(tr.ModalMsgManualMigrationLocation, migrationFolder),
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	mc.openModal(modal)
}

// showMigrationNameInput shows input modal for migration name
func (mc *MigrationsController) showMigrationNameInput() {
	tr := mc.c.GetTranslationSet()

	modal := NewInputModal(mc.g, tr, tr.ModalTitleEnterMigrationName,
		func(input string) {
			// Replace spaces with underscores
			migrationName := strings.ReplaceAll(strings.TrimSpace(input), " ", "_")

			// Close input modal
			mc.closeModal()

			// Execute actual migration creation
			mc.executeCreateMigration(migrationName)
		},
		func() {
			// Cancel - just close modal
			mc.closeModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan}).
		WithSubtitle(tr.ModalMsgSpacesReplaced).
		WithRequired(true).
		OnValidationFail(func(reason string) {
			// Validation failed - show error
			mc.closeModal()
			errorModal := NewMessageModal(mc.g, tr, tr.ModalTitleValidationFailed,
				reason,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(errorModal)
		})

	mc.openModal(modal)
}

// showManualMigrationInput shows input modal for manual migration name
func (mc *MigrationsController) showManualMigrationInput() {
	tr := mc.c.GetTranslationSet()

	modal := NewInputModal(mc.g, tr, tr.ModalTitleEnterMigrationName,
		func(input string) {
			// Replace spaces with underscores
			migrationName := strings.ReplaceAll(strings.TrimSpace(input), " ", "_")

			// Close input modal
			mc.closeModal()

			// Create manual migration
			mc.createManualMigration(migrationName)
		},
		func() {
			// Cancel - just close modal
			mc.closeModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan}).
		WithSubtitle(tr.ModalMsgSpacesReplaced).
		WithRequired(true).
		OnValidationFail(func(reason string) {
			// Validation failed - show error
			mc.closeModal()
			errorModal := NewMessageModal(mc.g, tr, tr.ModalTitleValidationFailed,
				reason,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(errorModal)
		})

	mc.openModal(modal)
}

// MigrateResolve resolves a failed migration
func (mc *MigrationsController) MigrateResolve() {
	tr := mc.c.GetTranslationSet()

	// Get selected migration
	selectedMigration := mc.migrationsCtx.GetSelectedMigration()
	if selectedMigration == nil {
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleNoMigrationSelected,
			tr.ModalMsgSelectMigrationResolve,
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		mc.openModal(modal)
		return
	}

	// Check if migration is failed (only In-Transaction migrations can be resolved)
	if !selectedMigration.IsFailed {
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleCannotResolveMigration,
			tr.ModalMsgOnlyInTransactionResolve,
			fmt.Sprintf(tr.ModalMsgMigrationNotFailed, selectedMigration.Name),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
		return
	}

	// Show ListModal with resolve options
	migrationName := selectedMigration.Name

	items := []ListModalItem{
		{
			Label:       tr.ListItemMarkApplied,
			Description: tr.ListItemDescMarkApplied,
			OnSelect: func() error {
				mc.closeModal()
				mc.executeResolve(migrationName, "applied")
				return nil
			},
		},
		{
			Label:       tr.ListItemMarkRolledBack,
			Description: tr.ListItemDescMarkRolledBack,
			OnSelect: func() error {
				mc.closeModal()
				mc.executeResolve(migrationName, "rolled-back")
				return nil
			},
		},
	}

	modal := NewListModal(mc.g, tr, fmt.Sprintf(tr.ModalTitleResolveMigration, migrationName), items,
		func() { mc.closeModal() },
	).WithStyle(MessageModalStyle{TitleColor: ColorCyan, BorderColor: ColorCyan})

	mc.openModal(modal)
}

// executeResolve runs npx prisma migrate resolve with the specified action
func (mc *MigrationsController) executeResolve(migrationName string, action string) {
	tr := mc.c.GetTranslationSet()

	actionLabel := tr.ActionLabelApplied
	if action == "rolled-back" {
		actionLabel = tr.ActionLabelRolledBack
	}

	mc.runStreamCmd(AsyncCommandOpts{
		Name:          "Migrate Resolve",
		Args:          []string{"npx", "prisma", "migrate", "resolve", "--" + action, migrationName},
		LogAction:     tr.LogActionMigrateResolve,
		LogDetail:     fmt.Sprintf(tr.LogMsgMarkingMigration, actionLabel, migrationName),
		ErrorTitle:    tr.ModalTitleMigrateResolveError,
		ErrorStartMsg: tr.ModalMsgFailedStartMigrateResolve,
		OnSuccess: func(out *context.OutputContext, cwd string) {
			mc.c.FinishCommand()
			mc.c.RefreshAll()
			out.LogAction(tr.LogActionMigrateResolveComplete, fmt.Sprintf(tr.LogMsgMigrationMarked, actionLabel))
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrateResolveSuccess,
				fmt.Sprintf(tr.ModalMsgMigrationMarkedSuccess, actionLabel),
			).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
			mc.openModal(modal)
		},
		OnFailure: func(out *context.OutputContext, cwd string, exitCode int) {
			mc.c.FinishCommand()
			mc.c.RefreshAll()
			out.LogAction(tr.LogActionMigrateResolveFailed, fmt.Sprintf(tr.LogMsgMigrateResolveFailedCode, exitCode))
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrateResolveFailed,
				fmt.Sprintf(tr.ModalMsgMigrateResolveFailedWithCode, exitCode),
				tr.ModalMsgCheckOutputPanel,
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(modal)
		},
		OnError: func(out *context.OutputContext, cwd string, err error) {
			mc.c.FinishCommand()
			out.LogAction(tr.LogActionMigrateResolveError, err.Error())
			modal := NewMessageModal(mc.g, tr, tr.ModalTitleMigrateResolveError,
				tr.ModalMsgFailedRunMigrateResolve,
				err.Error(),
			).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
			mc.openModal(modal)
		},
	})
}

// DeleteMigration deletes a pending migration
func (mc *MigrationsController) DeleteMigration() {
	tr := mc.c.GetTranslationSet()

	// Get selected migration
	selected := mc.migrationsCtx.GetSelectedMigration()
	if selected == nil {
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleNoSelection,
			tr.ModalMsgSelectMigrationDelete,
		).WithStyle(MessageModalStyle{TitleColor: ColorYellow, BorderColor: ColorYellow})
		mc.openModal(modal)
		return
	}

	// Validate: Can only delete if it exists locally
	if selected.Path == "" {
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleCannotDelete,
			tr.ModalMsgMigrationDBOnly,
			tr.ModalMsgCannotDeleteNoLocalFile,
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
		return
	}

	// Validate: Can only delete pending migrations (not applied to DB)
	if mc.migrationsCtx.IsDBConnected() && selected.AppliedAt != nil {
		modal := NewMessageModal(mc.g, tr, tr.ModalTitleCannotDelete,
			tr.ModalMsgMigrationAlreadyApplied,
			tr.ModalMsgDeleteLocalInconsistency,
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
		return
	}

	// Confirm deletion
	modal := NewConfirmModal(mc.g, tr, tr.ModalTitleDeleteMigration,
		fmt.Sprintf(tr.ModalMsgConfirmDeleteMigration, selected.Name),
		func() {
			mc.closeModal()
			mc.executeDeleteMigration(selected.Path, selected.Name)
		},
		func() {
			mc.closeModal()
		},
	).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
	mc.openModal(modal)
}

// executeDeleteMigration performs the actual deletion
func (mc *MigrationsController) executeDeleteMigration(path, name string) {
	tr := mc.c.GetTranslationSet()

	if err := os.RemoveAll(path); err != nil {
		mc.outputCtx.LogActionRed(tr.ModalTitleDeleteError, fmt.Sprintf(tr.LogMsgFailedDeleteMigration, err.Error()))

		modal := NewMessageModal(mc.g, tr, tr.ModalTitleDeleteError,
			tr.ModalMsgFailedDeleteFolder,
			err.Error(),
		).WithStyle(MessageModalStyle{TitleColor: ColorRed, BorderColor: ColorRed})
		mc.openModal(modal)
		return
	}

	// Success
	mc.outputCtx.LogAction(tr.LogActionDeleted, fmt.Sprintf(tr.LogMsgMigrationDeleted, name))

	// Refresh to update list
	mc.c.RefreshAll()

	modal := NewMessageModal(mc.g, tr, tr.ModalTitleDeleted,
		tr.ModalMsgMigrationDeletedSuccess,
	).WithStyle(MessageModalStyle{TitleColor: ColorGreen, BorderColor: ColorGreen})
	mc.openModal(modal)
}
