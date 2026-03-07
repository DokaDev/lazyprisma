package i18n

type TranslationSet struct {
	// Panel Titles
	PanelTitleOutput    string
	PanelTitleWorkspace string
	PanelTitleDetails   string

	// Tab Labels
	TabLocal        string
	TabPending      string
	TabDBOnly       string
	TabDetails      string
	TabActionNeeded string

	// Error Messages (general)
	ErrorFailedGetWorkingDirectory   string
	ErrorLoadingLocalMigrations      string
	ErrorNoMigrationsFound           string
	ErrorFailedAccessMigrationsPanel string
	ErrorNoDBConnectionDetected      string
	ErrorEnsureDBAccessible          string
	ErrorFailedGetWorkingDir         string
	ErrorCannotExecuteCommand        string
	ErrorCommandCurrentlyRunning     string
	ErrorOperationBlocked            string

	// Modal Titles
	ModalTitleError                     string
	ModalTitleDBConnectionRequired      string
	ModalTitleMigrationError            string
	ModalTitleMigrationCreated          string
	ModalTitleMigrationFailed           string
	ModalTitleMigrateDeploySuccess      string
	ModalTitleMigrateDeployFailed       string
	ModalTitleMigrateDeployError        string
	ModalTitleGenerateSuccess           string
	ModalTitleGenerateFailed            string
	ModalTitleGenerateError             string
	ModalTitleSchemaValidationFailed    string
	ModalTitleNoMigrationSelected       string
	ModalTitleCannotResolveMigration    string
	ModalTitleMigrateResolveSuccess     string
	ModalTitleMigrateResolveFailed      string
	ModalTitleMigrateResolveError       string
	ModalTitleStudioError               string
	ModalTitleStudioStopped             string
	ModalTitleStudioStarted             string
	ModalTitleNoSelection               string
	ModalTitleCannotDelete              string
	ModalTitleDeleteError               string
	ModalTitleDeleted                   string
	ModalTitleClipboardError            string
	ModalTitleCopied                    string
	ModalTitlePendingMigrationsDetected string
	ModalTitleDBOnlyMigrationsDetected  string
	ModalTitleChecksumMismatchDetected  string
	ModalTitleEmptyPendingDetected      string
	ModalTitleOperationBlocked          string
	ModalTitleDeleteMigration           string
	ModalTitleValidationFailed          string
	ModalTitleMigrateDev                string
	ModalTitleResolveMigration          string
	ModalTitleCopyToClipboard           string
	ModalTitleEnterMigrationName        string

	// Modal Messages
	ModalMsgMigrationCreatedSuccess     string
	ModalMsgMigrationCreatedDetail      string
	ModalMsgMigrationFailedWithCode     string
	ModalMsgCheckOutputPanel            string
	ModalMsgMigrationsAppliedSuccess    string
	ModalMsgMigrateDeployFailedWithCode string
	ModalMsgFailedRunMigrateDeploy      string
	ModalMsgFailedStartMigrateDeploy    string
	ModalMsgPrismaClientGenerated       string
	ModalMsgGenerateFailedSchemaErrors  string
	ModalMsgGenerateFailedWithCode      string
	ModalMsgSchemaValidCheckOutput      string
	ModalMsgFailedRunGenerate           string
	ModalMsgFailedStartGenerate         string
	ModalMsgSelectMigrationResolve      string
	ModalMsgOnlyInTransactionResolve    string
	ModalMsgMigrationNotFailed          string
	ModalMsgMigrationMarkedSuccess      string
	ModalMsgMigrateResolveFailedWithCode string
	ModalMsgFailedRunMigrateResolve     string
	ModalMsgFailedStartMigrateResolve   string
	ModalMsgFailedStopStudio            string
	ModalMsgStudioStopped               string
	ModalMsgFailedStartStudio           string
	ModalMsgStudioRunningAt             string
	ModalMsgPressStopStudio             string
	ModalMsgSelectMigrationDelete       string
	ModalMsgMigrationDBOnly             string
	ModalMsgCannotDeleteNoLocalFile     string
	ModalMsgMigrationAlreadyApplied     string
	ModalMsgDeleteLocalInconsistency    string
	ModalMsgFailedCreateFolder          string
	ModalMsgFailedDeleteFolder          string
	ModalMsgFailedWriteMigrationFile    string
	ModalMsgMigrationDeletedSuccess     string
	ModalMsgFailedCopyClipboard         string
	ModalMsgCopiedToClipboard           string
	ModalMsgPendingMigrationsWarning    string
	ModalMsgCannotCreateWithDBOnly      string
	ModalMsgResolveDBOnlyFirst          string
	ModalMsgCannotCreateWithMismatch    string
	ModalMsgMigrationModifiedLocally    string
	ModalMsgCannotCreateWithEmpty       string
	ModalMsgMigrationPendingEmpty       string
	ModalMsgDeleteOrAddContent          string
	ModalMsgAnotherOperationRunning     string
	ModalMsgWaitComplete                string
	ModalMsgConfirmDeleteMigration      string
	ModalMsgSpacesReplaced              string
	ModalMsgInputRequired               string
	ModalMsgManualMigrationCreated      string
	ModalMsgManualMigrationLocation     string
	CopyLabelMigrationName              string
	CopyLabelMigrationPath              string
	CopyLabelChecksum                   string

	// Modal Footers
	ModalFooterInputSubmitCancel string
	ModalFooterListNavigate      string
	ModalFooterMessageClose      string
	ModalFooterConfirmYesNo      string

	// Status Bar
	StatusStudioOn string
	KeyHintRefresh string
	KeyHintDev     string
	KeyHintDeploy  string
	KeyHintGenerate string
	KeyHintResolve string
	KeyHintStudio  string
	KeyHintCopy    string

	// Log Actions
	LogActionMigrateDeploy         string
	LogMsgRunningMigrateDeploy     string
	LogActionMigrateDeployComplete string
	LogMsgMigrationsAppliedSuccess string
	LogActionMigrateDeployFailed   string
	LogMsgMigrateDeployFailedCode  string
	LogActionMigrateResolve        string
	LogMsgMarkingMigration         string
	LogActionMigrateResolveComplete string
	LogMsgMigrationMarked          string
	LogActionMigrateResolveFailed  string
	LogMsgMigrateResolveFailedCode string
	LogActionMigrateResolveError   string
	LogActionGenerate              string
	LogMsgRunningGenerate          string
	LogActionGenerateComplete      string
	LogMsgPrismaClientGeneratedSuccess string
	LogActionGenerateFailed        string
	LogMsgCheckingSchemaErrors     string
	LogActionSchemaValidationFailed string
	LogMsgFoundSchemaErrors        string
	LogActionGenerateError         string
	LogActionStudio                string
	LogMsgStartingStudio           string
	LogActionStudioStarted         string
	LogMsgStudioListeningAt        string
	LogActionStudioStopped         string
	LogMsgStudioHasStopped         string
	LogActionMigrateDev            string
	LogMsgCreatingMigration        string
	LogActionMigrateComplete       string
	LogMsgMigrationCreatedSuccess  string
	LogActionMigrateFailed         string
	LogMsgMigrationCreationFailedCode string
	LogActionMigrationError        string
	LogMsgFailedDeleteMigration    string
	LogActionDeleted               string
	LogMsgMigrationDeleted         string
	SuccessAllPanelsRefreshed      string
	ActionRefresh                  string

	// List Modal Items
	ListItemSchemaDiffMigration     string
	ListItemDescSchemaDiffMigration string
	ListItemManualMigration         string
	ListItemDescManualMigration     string
	ListItemMarkApplied             string
	ListItemDescMarkApplied         string
	ListItemMarkRolledBack          string
	ListItemDescMarkRolledBack      string
	ListItemCopyName                string
	ListItemCopyPath                string
	ListItemCopyChecksum            string

	// Details Panel - Migration Status
	MigrationStatusInTransaction  string
	MigrationStatusDBOnly         string
	MigrationStatusChecksumMismatch string
	MigrationStatusApplied        string
	MigrationStatusEmptyMigration string
	MigrationStatusPending        string

	// Details Panel - Labels & Descriptions
	DetailsPanelInitialPlaceholder      string
	DetailsNameLabel                    string
	DetailsTimestampLabel               string
	DetailsPathLabel                    string
	DetailsStatusLabel                  string
	DetailsAppliedAtLabel               string
	DetailsDownMigrationLabel           string
	DetailsDownMigrationAvailable       string
	DetailsDownMigrationNotAvailable    string
	DetailsStartedAtLabel               string
	DetailsInTransactionWarning         string
	DetailsNoAdditionalMigrationsWarning string
	DetailsResolveManuallyInstruction   string
	DetailsErrorLogsLabel               string
	DetailsDBOnlyDescription            string
	DetailsChecksumModifiedDescription  string
	DetailsChecksumIssuesWarning        string
	DetailsLocalChecksumLabel           string
	DetailsHistoryChecksumLabel         string
	DetailsEmptyMigrationDescription    string
	DetailsEmptyMigrationWarning        string
	DetailsDownMigrationSQLLabel        string
	ErrorReadingMigrationSQL            string

	// Details Panel - Action Needed
	ActionNeededNoIssuesMessage                string
	ActionNeededHeader                         string
	ActionNeededIssueSingular                   string
	ActionNeededIssuePlural                     string
	ActionNeededEmptyMigrationsHeader           string
	ActionNeededEmptyDescription                string
	ActionNeededAffectedLabel                   string
	ActionNeededRecommendedLabel                string
	ActionNeededAddMigrationSQL                 string
	ActionNeededDeleteEmptyFolders              string
	ActionNeededMarkAsBaseline                  string
	ActionNeededChecksumMismatchHeader          string
	ActionNeededChecksumModifiedDescription     string
	ActionNeededWarningPrefix                   string
	ActionNeededEditingInconsistenciesWarning   string
	ActionNeededRevertLocalChanges              string
	ActionNeededCreateNewInstead                string
	ActionNeededContactTeamIfNeeded             string
	ActionNeededSchemaValidationErrorsHeader    string
	ActionNeededSchemaValidationFailedDesc      string
	ActionNeededFixBeforeMigration              string
	ActionNeededValidationOutputLabel           string
	ActionNeededRecommendedActionsLabel         string
	ActionNeededFixSchemaErrors                 string
	ActionNeededCheckLineNumbers                string
	ActionNeededReferPrismaDocumentation        string

	// Workspace Panel
	WorkspaceVersionLine             string
	WorkspacePrismaGlobalIndicator   string
	WorkspaceGitLine                 string
	WorkspaceSchemaModifiedIndicator string
	WorkspaceBranchFormat            string
	WorkspaceNotGitRepository        string
	WorkspaceConnected               string
	WorkspaceNotConfigured           string
	WorkspaceDisconnected            string
	WorkspaceProviderLine            string
	WorkspaceHardcodedIndicator      string
	WorkspaceNotSet                  string
	WorkspaceErrorFormat             string
	WorkspaceErrorGetWorkingDirectory string
	WorkspaceErrorSchemaNotFound     string
	WorkspaceNotConfiguredSuffix     string
	WorkspaceDatabaseURLNotConfigured string
	WorkspaceNoDatabaseURL           string
	WorkspaceVersionNotFound         string

	// Migrations Panel
	MigrationsFooterFormat string

	// main.go strings
	VersionOutput              string
	ErrorFailedGetCurrentDir   string
	ErrorNotPrismaWorkspace    string
	ErrorExpectedOneOf         string
	ErrorExpectedConfigV7Plus  string
	ErrorExpectedSchemaV7Minus string
	ErrorFailedCreateApp       string
	ErrorFailedRegisterKeybindings string
	ErrorAppRuntime            string
}

func EnglishTranslationSet() *TranslationSet {
	return &TranslationSet{
		// Panel Titles
		PanelTitleOutput:    "Output",
		PanelTitleWorkspace: "Workspace",
		PanelTitleDetails:   "Details",

		// Tab Labels
		TabLocal:        "Local",
		TabPending:      "Pending",
		TabDBOnly:       "DB-Only",
		TabDetails:      "Details",
		TabActionNeeded: "Action-Needed",

		// Error Messages (general)
		ErrorFailedGetWorkingDirectory:   "Error: Failed to get working directory",
		ErrorLoadingLocalMigrations:      "Error loading local migrations: %v",
		ErrorNoMigrationsFound:           "No migrations found",
		ErrorFailedAccessMigrationsPanel: "Failed to access migrations panel.",
		ErrorNoDBConnectionDetected:      "No database connection detected.",
		ErrorEnsureDBAccessible:          "Please ensure your database is running and accessible.",
		ErrorFailedGetWorkingDir:         "Failed to get working directory:",
		ErrorCannotExecuteCommand:        "Cannot execute '%s'",
		ErrorCommandCurrentlyRunning:     " — '%s' is currently running",
		ErrorOperationBlocked:            "Operation Blocked",

		// Modal Titles
		ModalTitleError:                     "Error",
		ModalTitleDBConnectionRequired:      "Database Connection Required",
		ModalTitleMigrationError:            "Migration Error",
		ModalTitleMigrationCreated:          "Migration Created",
		ModalTitleMigrationFailed:           "Migration Failed",
		ModalTitleMigrateDeploySuccess:      "Migrate Deploy Successful",
		ModalTitleMigrateDeployFailed:       "Migrate Deploy Failed",
		ModalTitleMigrateDeployError:        "Migrate Deploy Error",
		ModalTitleGenerateSuccess:           "Generate Successful",
		ModalTitleGenerateFailed:            "Generate Failed",
		ModalTitleGenerateError:             "Generate Error",
		ModalTitleSchemaValidationFailed:    "Schema Validation Failed",
		ModalTitleNoMigrationSelected:       "No Migration Selected",
		ModalTitleCannotResolveMigration:    "Cannot Resolve Migration",
		ModalTitleMigrateResolveSuccess:     "Migrate Resolve Successful",
		ModalTitleMigrateResolveFailed:      "Migrate Resolve Failed",
		ModalTitleMigrateResolveError:       "Migrate Resolve Error",
		ModalTitleStudioError:               "Studio Error",
		ModalTitleStudioStopped:             "Studio Stopped",
		ModalTitleStudioStarted:             "Prisma Studio Started",
		ModalTitleNoSelection:               "No Selection",
		ModalTitleCannotDelete:              "Cannot Delete",
		ModalTitleDeleteError:               "Delete Error",
		ModalTitleDeleted:                   "Deleted",
		ModalTitleClipboardError:            "Clipboard Error",
		ModalTitleCopied:                    "Copied",
		ModalTitlePendingMigrationsDetected: "Pending Migrations Detected",
		ModalTitleDBOnlyMigrationsDetected:  "DB-Only Migrations Detected",
		ModalTitleChecksumMismatchDetected:  "Checksum Mismatch Detected",
		ModalTitleEmptyPendingDetected:      "Empty Pending Migration Detected",
		ModalTitleOperationBlocked:          "Operation Blocked",
		ModalTitleDeleteMigration:           "Delete Migration",
		ModalTitleValidationFailed:          "Validation Failed",
		ModalTitleMigrateDev:                "Migrate Dev",
		ModalTitleResolveMigration:          "Resolve Migration: %s",
		ModalTitleCopyToClipboard:           "Copy to Clipboard",
		ModalTitleEnterMigrationName:        "Enter migration name",

		// Modal Messages
		ModalMsgMigrationCreatedSuccess:      "Migration '%s' created successfully!",
		ModalMsgMigrationCreatedDetail:       "You can find it in the prisma/migrations directory.",
		ModalMsgMigrationFailedWithCode:      "Prisma migrate dev failed with exit code: %d",
		ModalMsgCheckOutputPanel:             "Check output panel for details.",
		ModalMsgMigrationsAppliedSuccess:     "Migrations applied successfully!",
		ModalMsgMigrateDeployFailedWithCode:  "Prisma migrate deploy failed with exit code: %d",
		ModalMsgFailedRunMigrateDeploy:       "Failed to run prisma migrate deploy:",
		ModalMsgFailedStartMigrateDeploy:     "Failed to start migrate deploy:",
		ModalMsgPrismaClientGenerated:        "Prisma Client generated successfully!",
		ModalMsgGenerateFailedSchemaErrors:   "Generate failed due to schema errors.",
		ModalMsgGenerateFailedWithCode:       "Prisma generate failed with exit code: %d",
		ModalMsgSchemaValidCheckOutput:       "Schema is valid. Check output panel for details.",
		ModalMsgFailedRunGenerate:            "Failed to run prisma generate:",
		ModalMsgFailedStartGenerate:          "Failed to start generate:",
		ModalMsgSelectMigrationResolve:       "Please select a migration to resolve.",
		ModalMsgOnlyInTransactionResolve:     "Only migrations in 'In-Transaction' state can be resolved.",
		ModalMsgMigrationNotFailed:           "Migration '%s' is not in a failed state.",
		ModalMsgMigrationMarkedSuccess:       "Migration marked as %s successfully!",
		ModalMsgMigrateResolveFailedWithCode: "Prisma migrate resolve failed with exit code: %d",
		ModalMsgFailedRunMigrateResolve:      "Failed to run prisma migrate resolve:",
		ModalMsgFailedStartMigrateResolve:    "Failed to start migrate resolve:",
		ModalMsgFailedStopStudio:             "Failed to stop Prisma Studio:",
		ModalMsgStudioStopped:                "Prisma Studio has been stopped.",
		ModalMsgFailedStartStudio:            "Failed to start Prisma Studio:",
		ModalMsgStudioRunningAt:              "Prisma Studio is running at http://localhost:5555",
		ModalMsgPressStopStudio:              "Press 'S' again to stop it.",
		ModalMsgSelectMigrationDelete:        "Please select a migration to delete.",
		ModalMsgMigrationDBOnly:              "This migration exists only in the database (DB-Only).",
		ModalMsgCannotDeleteNoLocalFile:      "Cannot delete a migration that has no local file.",
		ModalMsgMigrationAlreadyApplied:      "This migration has already been applied to the database.",
		ModalMsgDeleteLocalInconsistency:     "Deleting it locally will cause inconsistency.",
		ModalMsgFailedCreateFolder:           "Failed to create migration folder:",
		ModalMsgFailedDeleteFolder:           "Failed to delete migration folder:",
		ModalMsgFailedWriteMigrationFile:     "Failed to write migration file:",
		ModalMsgMigrationDeletedSuccess:      "Migration deleted successfully.",
		ModalMsgFailedCopyClipboard:          "Failed to copy to clipboard:",
		ModalMsgCopiedToClipboard:            "%s copied to clipboard!",
		ModalMsgPendingMigrationsWarning:     "Prisma automatically applies pending migrations before creating new ones. This may cause unintended behaviour in the future. Do you wish to continue?",
		ModalMsgCannotCreateWithDBOnly:       "Cannot create new migration whilst DB-Only migrations exist.",
		ModalMsgResolveDBOnlyFirst:           "Please resolve DB-Only migrations first.",
		ModalMsgCannotCreateWithMismatch:     "Cannot create new migration whilst checksum mismatch exists.",
		ModalMsgMigrationModifiedLocally:     "Migration '%s' has been modified locally.",
		ModalMsgCannotCreateWithEmpty:        "Cannot create new migration whilst empty pending migrations exist.",
		ModalMsgMigrationPendingEmpty:        "Migration '%s' is pending and empty.",
		ModalMsgDeleteOrAddContent:           "Please delete it or add SQL content.",
		ModalMsgAnotherOperationRunning:      "Another operation is currently running.",
		ModalMsgWaitComplete:                 "Please wait for it to complete.",
		ModalMsgConfirmDeleteMigration:       "Are you sure you want to delete this migration?\n\n%s\n\nThis action cannot be undone.",
		ModalMsgSpacesReplaced:               "Spaces will be replaced with underscores",
		ModalMsgInputRequired:                "Input is required",
		ModalMsgManualMigrationCreated:       "Created: %s",
		ModalMsgManualMigrationLocation:      "Location: %s",
		CopyLabelMigrationName:               "Migration Name",
		CopyLabelMigrationPath:               "Migration Path",
		CopyLabelChecksum:                    "Checksum",

		// Modal Footers
		ModalFooterInputSubmitCancel: "[Enter] Submit [ESC] Cancel",
		ModalFooterListNavigate:      "[↑/↓] Navigate [Enter] Select [ESC] Cancel",
		ModalFooterMessageClose:      " [Enter/q/ESC] Close ",
		ModalFooterConfirmYesNo:      " [Y] Yes [N] No [ESC] Cancel ",

		// Status Bar
		StatusStudioOn:  "[Studio: ON]",
		KeyHintRefresh:  "efresh",
		KeyHintDev:      "ev",
		KeyHintDeploy:   "eploy",
		KeyHintGenerate: "enerate",
		KeyHintResolve:  "resolve",
		KeyHintStudio:   "tudio",
		KeyHintCopy:     "opy",

		// Log Actions
		LogActionMigrateDeploy:            "Migrate Deploy",
		LogMsgRunningMigrateDeploy:        "Running prisma migrate deploy...",
		LogActionMigrateDeployComplete:    "Migrate Deploy Complete",
		LogMsgMigrationsAppliedSuccess:    "Migrations applied successfully",
		LogActionMigrateDeployFailed:      "Migrate Deploy Failed",
		LogMsgMigrateDeployFailedCode:     "Migrate deploy failed with exit code: %d",
		LogActionMigrateResolve:           "Migrate Resolve",
		LogMsgMarkingMigration:            "Marking migration as %s: %s",
		LogActionMigrateResolveComplete:   "Migrate Resolve Complete",
		LogMsgMigrationMarked:             "Migration marked as %s successfully",
		LogActionMigrateResolveFailed:     "Migrate Resolve Failed",
		LogMsgMigrateResolveFailedCode:    "Migrate resolve failed with exit code: %d",
		LogActionMigrateResolveError:      "Migrate Resolve Error",
		LogActionGenerate:                 "Generate",
		LogMsgRunningGenerate:             "Running prisma generate...",
		LogActionGenerateComplete:         "Generate Complete",
		LogMsgPrismaClientGeneratedSuccess: "Prisma Client generated successfully",
		LogActionGenerateFailed:           "Generate Failed",
		LogMsgCheckingSchemaErrors:        "Checking schema for errors...",
		LogActionSchemaValidationFailed:   "Schema Validation Failed",
		LogMsgFoundSchemaErrors:           "Found %d schema errors",
		LogActionGenerateError:            "Generate Error",
		LogActionStudio:                   "Studio",
		LogMsgStartingStudio:              "Starting Prisma Studio...",
		LogActionStudioStarted:            "Studio Started",
		LogMsgStudioListeningAt:           "Prisma Studio is running at http://localhost:5555",
		LogActionStudioStopped:            "Studio Stopped",
		LogMsgStudioHasStopped:            "Prisma Studio has been stopped",
		LogActionMigrateDev:               "Migrate Dev",
		LogMsgCreatingMigration:           "Creating migration: %s",
		LogActionMigrateComplete:          "Migrate Complete",
		LogMsgMigrationCreatedSuccess:     "Migration created successfully",
		LogActionMigrateFailed:            "Migrate Failed",
		LogMsgMigrationCreationFailedCode: "Migration creation failed with exit code: %d",
		LogActionMigrationError:           "Migration Error",
		LogMsgFailedDeleteMigration:       "Failed to delete migration: %s",
		LogActionDeleted:                  "Deleted",
		LogMsgMigrationDeleted:            "Migration '%s' deleted",
		SuccessAllPanelsRefreshed:         "All panels have been refreshed",
		ActionRefresh:                     "Refresh",

		// List Modal Items
		ListItemSchemaDiffMigration:     "Schema diff-based migration",
		ListItemDescSchemaDiffMigration: "Create a migration from changes in Prisma schema, apply it to the database, trigger generators (e.g. Prisma Client)",
		ListItemManualMigration:         "Manual migration",
		ListItemDescManualMigration:     "This tool creates manual migrations for database changes that cannot be expressed through Prisma schema diff. It is used to explicitly record and version control database-specific logic such as triggers, functions, and DML operations that cannot be managed at the Prisma schema level.",
		ListItemMarkApplied:             "Mark as applied",
		ListItemDescMarkApplied:         "Mark this migration as successfully applied to the database. Use this if you have manually fixed the issue and the migration changes are now present in the database.",
		ListItemMarkRolledBack:          "Mark as rolled back",
		ListItemDescMarkRolledBack:      "Mark this migration as rolled back (reverted from the database). Use this if you have manually reverted the changes and the migration is no longer applied to the database.",
		ListItemCopyName:                "Copy Name",
		ListItemCopyPath:                "Copy Path",
		ListItemCopyChecksum:            "Copy Checksum",

		// Details Panel - Migration Status
		MigrationStatusInTransaction:    "⚠ In-Transaction",
		MigrationStatusDBOnly:           "✗ DB Only",
		MigrationStatusChecksumMismatch: "⚠ Checksum Mismatch",
		MigrationStatusApplied:          "✓ Applied",
		MigrationStatusEmptyMigration:   "⚠ Empty Migration",
		MigrationStatusPending:          "⚠ Pending",

		// Details Panel - Labels & Descriptions
		DetailsPanelInitialPlaceholder:       "Details\n\nSelect a migration to view details...",
		DetailsNameLabel:                     "Name: %s\n",
		DetailsTimestampLabel:                "Timestamp: %s\n",
		DetailsPathLabel:                     "Path: %s\n",
		DetailsStatusLabel:                   "Status: ",
		DetailsAppliedAtLabel:                "Applied at: %s",
		DetailsDownMigrationLabel:            "Down Migration: ",
		DetailsDownMigrationAvailable:        "✓ Available",
		DetailsDownMigrationNotAvailable:     "✗ Not available",
		DetailsStartedAtLabel:                "Started At: ",
		DetailsInTransactionWarning:          "⚠ WARNING: This migration is stuck in an incomplete state.",
		DetailsNoAdditionalMigrationsWarning: "No additional migrations can be applied until this is resolved.",
		DetailsResolveManuallyInstruction:    "Please resolve this migration manually before proceeding.\n",
		DetailsErrorLogsLabel:                "Error Logs:",
		DetailsDBOnlyDescription:             "This migration exists in the database but not in local files.",
		DetailsChecksumModifiedDescription:   "The local migration file has been modified after being applied to the database.\n",
		DetailsChecksumIssuesWarning:         "This can cause issues during deployment.\n\n",
		DetailsLocalChecksumLabel:            "Local Checksum:   ",
		DetailsHistoryChecksumLabel:          "History Checksum: ",
		DetailsEmptyMigrationDescription:     "This migration folder is empty or missing migration.sql.\n",
		DetailsEmptyMigrationWarning:         "This may cause issues during deployment.",
		DetailsDownMigrationSQLLabel:         "Down Migration SQL:",
		ErrorReadingMigrationSQL:             "Error reading migration.sql:\n%v",

		// Details Panel - Action Needed
		ActionNeededNoIssuesMessage:                "No action required\n\nAll migrations are in good state and schema is valid.",
		ActionNeededHeader:                         "⚠ Action Needed",
		ActionNeededIssueSingular:                   " issue",
		ActionNeededIssuePlural:                     "s",
		ActionNeededEmptyMigrationsHeader:           "Empty Migrations",
		ActionNeededEmptyDescription:                "These migrations have no SQL content.\n\n",
		ActionNeededAffectedLabel:                   "Affected:\n",
		ActionNeededRecommendedLabel:                "Recommended Actions:\n",
		ActionNeededAddMigrationSQL:                 "  → Add migration.sql manually\n",
		ActionNeededDeleteEmptyFolders:              "  → Delete empty migration folders\n",
		ActionNeededMarkAsBaseline:                  "  → Mark as baseline migration\n\n",
		ActionNeededChecksumMismatchHeader:          "Checksum Mismatch",
		ActionNeededChecksumModifiedDescription:     "Migration content was modified after\nbeing applied to database.\n\n",
		ActionNeededWarningPrefix:                   "⚠ WARNING: ",
		ActionNeededEditingInconsistenciesWarning:   "Editing applied migrations\ncan cause inconsistencies.\n\n",
		ActionNeededRevertLocalChanges:              "  → Revert local changes\n",
		ActionNeededCreateNewInstead:                "  → Create new migration instead\n",
		ActionNeededContactTeamIfNeeded:             "  → Contact team if needed\n\n",
		ActionNeededSchemaValidationErrorsHeader:    "Schema Validation Errors",
		ActionNeededSchemaValidationFailedDesc:      "Schema validation failed.\n",
		ActionNeededFixBeforeMigration:              "Fix these issues before running migrations.\n\n",
		ActionNeededValidationOutputLabel:           "Validation Output:",
		ActionNeededRecommendedActionsLabel:         "Recommended Actions:",
		ActionNeededFixSchemaErrors:                 "  → Fix schema.prisma errors\n",
		ActionNeededCheckLineNumbers:                "  → Check line numbers in output above\n",
		ActionNeededReferPrismaDocumentation:        "  → Refer to Prisma documentation\n",

		// Workspace Panel
		WorkspaceVersionLine:              "Node: %s | Prisma: %s",
		WorkspacePrismaGlobalIndicator:    " (Global)",
		WorkspaceGitLine:                  "Git: %s",
		WorkspaceSchemaModifiedIndicator:  " (schema modified)",
		WorkspaceBranchFormat:             "(%s)",
		WorkspaceNotGitRepository:         "Git: Not a git repository",
		WorkspaceConnected:                "✓ Connected",
		WorkspaceNotConfigured:            "✗ Not configured",
		WorkspaceDisconnected:             "✗ Disconnected",
		WorkspaceProviderLine:             "Provider: %s  %s",
		WorkspaceHardcodedIndicator:       " (Hard coded)",
		WorkspaceNotSet:                   "Not set",
		WorkspaceErrorFormat:              "Error: %s",
		WorkspaceErrorGetWorkingDirectory: "Error getting working directory",
		WorkspaceErrorSchemaNotFound:      "Schema file not found",
		WorkspaceNotConfiguredSuffix:      " not configured",
		WorkspaceDatabaseURLNotConfigured: "DATABASE_URL not configured",
		WorkspaceNoDatabaseURL:            "No DATABASE_URL",
		WorkspaceVersionNotFound:          "Not found",

		// Migrations Panel
		MigrationsFooterFormat: "%d of %d",

		// main.go strings
		VersionOutput:              "LazyPrisma %s (%s)\n",
		ErrorFailedGetCurrentDir:   "Error: Failed to get current directory: %v\n",
		ErrorNotPrismaWorkspace:    "Error: Current directory is not a Prisma workspace.\n",
		ErrorExpectedOneOf:         "\nExpected one of:\n",
		ErrorExpectedConfigV7Plus:  "  - prisma.config.ts (Prisma v7.0+)\n",
		ErrorExpectedSchemaV7Minus: "  - prisma/schema.prisma (Prisma < v7.0)\n",
		ErrorFailedCreateApp:       "Failed to create app: %v\n",
		ErrorFailedRegisterKeybindings: "Failed to register keybindings: %v\n",
		ErrorAppRuntime:            "App error: %v\n",
	}
}
