package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

type DetailsPanel struct {
	BasePanel
	content              string
	originY              int    // Scroll position
	currentMigrationName string // Currently displayed migration name
	migrationsPanel      *MigrationsPanel

	// Tab management
	tabs                    []string                // Tab names (Details, Action-Needed)
	tabIndex                int                     // Current tab index
	actionNeededMigrations  []prisma.Migration      // Migrations requiring action (Empty + Mismatch)
	validationResult        *prisma.ValidateResult  // Schema validation result
	tabOriginY              map[string]int          // Scroll position per tab

	// Translation set (available from construction time)
	tr *i18n.TranslationSet

	// App reference for modal check (tab click events)
	app *App
}

func NewDetailsPanel(g *gocui.Gui, tr *i18n.TranslationSet) *DetailsPanel {
	return &DetailsPanel{
		BasePanel:              NewBasePanel(ViewDetails, g),
		tr:                     tr,
		content:                tr.DetailsPanelInitialPlaceholder,
		tabs:                   []string{tr.TabDetails},
		tabIndex:               0,
		actionNeededMigrations: []prisma.Migration{},
		tabOriginY:             make(map[string]int),
	}
}

func (d *DetailsPanel) Draw(dim boxlayout.Dimensions) error {
	v, err := d.g.SetView(d.id, dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	// Setup view WITHOUT title (tabs replace title)
	d.v = v
	v.Clear()
	v.Frame = true
	v.FrameRunes = d.frameRunes
	v.Wrap = true // Enable word wrap for long lines

	// Set tabs
	v.Tabs = d.tabs
	v.TabIndex = d.tabIndex

	// Set frame and tab colors based on focus
	if d.focused {
		v.FrameColor = FocusedFrameColor
		v.TitleColor = FocusedTitleColor
		if len(d.tabs) == 1 {
			v.SelFgColor = FocusedTitleColor // Single tab: treat like title
		} else {
			v.SelFgColor = FocusedActiveTabColor // Multiple tabs: use active tab color
		}
	} else {
		v.FrameColor = PrimaryFrameColor
		v.TitleColor = PrimaryTitleColor
		if len(d.tabs) == 1 {
			v.SelFgColor = PrimaryTitleColor // Single tab: treat like title
		} else {
			v.SelFgColor = PrimaryActiveTabColor // Multiple tabs: use active tab color
		}
	}

	// Render content based on current tab
	if d.tabIndex < len(d.tabs) {
		tabName := d.tabs[d.tabIndex]
		if d.app != nil && tabName == d.tr.TabActionNeeded {
			fmt.Fprint(v, d.buildActionNeededContent())
		} else {
			fmt.Fprint(v, d.content)
		}
	}

	// Adjust origin to ensure it's within valid bounds
	AdjustOrigin(v, &d.originY)
	v.SetOrigin(0, d.originY)

	return nil
}

func (d *DetailsPanel) SetContent(content string) {
	d.content = content
}

// buildActionNeededContent builds the content for the Action-Needed tab
func (d *DetailsPanel) buildActionNeededContent() string {
	// Count all issues
	emptyCount := 0
	mismatchCount := 0
	var emptyMigrations []prisma.Migration
	var mismatchMigrations []prisma.Migration

	for _, mig := range d.actionNeededMigrations {
		if mig.IsEmpty {
			emptyCount++
			emptyMigrations = append(emptyMigrations, mig)
		}
		if mig.ChecksumMismatch {
			mismatchCount++
			mismatchMigrations = append(mismatchMigrations, mig)
		}
	}

	validationErrorCount := 0
	if d.validationResult != nil && !d.validationResult.Valid {
		validationErrorCount = len(d.validationResult.Errors)
		if validationErrorCount == 0 {
			validationErrorCount = 1 // At least one error if validation failed
		}
	}

	totalCount := emptyCount + mismatchCount + validationErrorCount

	if totalCount == 0 {
		return d.tr.ActionNeededNoIssuesMessage
	}

	var content strings.Builder

	// Header
	content.WriteString(fmt.Sprintf("%s (%d%s", Yellow(d.tr.ActionNeededHeader), totalCount, d.tr.ActionNeededIssueSingular))
	if totalCount > 1 {
		content.WriteString(d.tr.ActionNeededIssuePlural)
	}
	content.WriteString(")\n\n")

	// Empty Migrations Section
	if emptyCount > 0 {
		content.WriteString(strings.Repeat("━", 40) + "\n")
		content.WriteString(fmt.Sprintf("%s (%d)\n", Red(d.tr.ActionNeededEmptyMigrationsHeader), emptyCount))
		content.WriteString(strings.Repeat("━", 40) + "\n\n")

		content.WriteString(d.tr.ActionNeededEmptyDescription)

		content.WriteString(d.tr.ActionNeededAffectedLabel)
		for _, mig := range emptyMigrations {
			_, name := parseMigrationName(mig.Name)
			content.WriteString(fmt.Sprintf("  • %s\n", Red(name)))
		}

		content.WriteString("\n" + d.tr.ActionNeededRecommendedLabel)
		content.WriteString(d.tr.ActionNeededAddMigrationSQL)
		content.WriteString(d.tr.ActionNeededDeleteEmptyFolders)
		content.WriteString(d.tr.ActionNeededMarkAsBaseline)
	}

	// Checksum Mismatch Section
	if mismatchCount > 0 {
		content.WriteString(strings.Repeat("━", 40) + "\n")
		content.WriteString(fmt.Sprintf("%s (%d)\n", Orange(d.tr.ActionNeededChecksumMismatchHeader), mismatchCount))
		content.WriteString(strings.Repeat("━", 40) + "\n\n")

		content.WriteString(d.tr.ActionNeededChecksumModifiedDescription)

		content.WriteString(Yellow(d.tr.ActionNeededWarningPrefix))
		content.WriteString(d.tr.ActionNeededEditingInconsistenciesWarning)

		content.WriteString(d.tr.ActionNeededAffectedLabel)
		for _, mig := range mismatchMigrations {
			_, name := parseMigrationName(mig.Name)
			content.WriteString(fmt.Sprintf("  • %s\n", Orange(name)))
		}

		content.WriteString("\n" + d.tr.ActionNeededRecommendedLabel)
		content.WriteString(d.tr.ActionNeededRevertLocalChanges)
		content.WriteString(d.tr.ActionNeededCreateNewInstead)
		content.WriteString(d.tr.ActionNeededContactTeamIfNeeded)
	}

	// Schema Validation Section
	if validationErrorCount > 0 {
		content.WriteString(strings.Repeat("━", 40) + "\n")
		content.WriteString(fmt.Sprintf("%s (%d)\n", Red(d.tr.ActionNeededSchemaValidationErrorsHeader), validationErrorCount))
		content.WriteString(strings.Repeat("━", 40) + "\n\n")

		content.WriteString(d.tr.ActionNeededSchemaValidationFailedDesc)
		content.WriteString(d.tr.ActionNeededFixBeforeMigration)

		// Show full validation output (contains detailed error info)
		if d.validationResult.Output != "" {
			content.WriteString(Stylize(d.tr.ActionNeededValidationOutputLabel, Style{FgColor: ColorYellow, Bold: true}) + "\n")
			// Display the full output with proper formatting (preserve all line breaks)
			outputLines := strings.Split(d.validationResult.Output, "\n")
			for _, line := range outputLines {
				// Highlight error lines
				if strings.Contains(line, "Error:") || strings.Contains(line, "error:") {
					content.WriteString(Red(line) + "\n")
				} else if strings.Contains(line, "-->") {
					content.WriteString(Yellow(line) + "\n")
				} else {
					// Preserve empty lines and all other content as-is
					content.WriteString(line + "\n")
				}
			}
			content.WriteString("\n")
		}

		content.WriteString(Stylize(d.tr.ActionNeededRecommendedActionsLabel, Style{FgColor: ColorYellow, Bold: true}) + "\n")
		content.WriteString(d.tr.ActionNeededFixSchemaErrors)
		content.WriteString(d.tr.ActionNeededCheckLineNumbers)
		content.WriteString(d.tr.ActionNeededReferPrismaDocumentation)
	}

	return content.String()
}

// highlightSQL applies syntax highlighting to SQL code with line numbers
func highlightSQL(code string) string {
	// Get SQL lexer
	lexer := lexers.Get("sql")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Get style (monokai is a popular dark theme)
	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	// Get terminal formatter with 256 colors
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize and format
	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code // Return original if highlighting fails
	}

	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code // Return original if highlighting fails
	}

	// Add line numbers
	highlighted := buf.String()
	lines := strings.Split(highlighted, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		// Line number in gray color, right-aligned to 4 digits
		result.WriteString(fmt.Sprintf("\033[90m%4d │\033[0m %s", i+1, line))
	}

	return result.String()
}

// ScrollUp scrolls the details panel up
func (d *DetailsPanel) ScrollUp() {
	if d.originY > 0 {
		d.originY--
	}
}

// ScrollDown scrolls the details panel down
func (d *DetailsPanel) ScrollDown() {
	d.originY++
	// AdjustOrigin will be called in Draw() to ensure bounds
}

// ScrollUpByWheel scrolls the details panel up by 2 lines (mouse wheel)
func (d *DetailsPanel) ScrollUpByWheel() {
	if d.originY > 0 {
		d.originY -= 2
		if d.originY < 0 {
			d.originY = 0
		}
	}
}

// ScrollDownByWheel scrolls the details panel down by 2 lines (mouse wheel)
func (d *DetailsPanel) ScrollDownByWheel() {
	if d.v == nil {
		return
	}

	// Get actual content lines from the rendered view buffer
	contentLines := len(d.v.ViewBufferLines())
	_, viewHeight := d.v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	// Only scroll if we haven't reached the bottom
	if d.originY < maxOrigin {
		d.originY += 2
		if d.originY > maxOrigin {
			d.originY = maxOrigin
		}
	}
}

// ScrollToTop scrolls to the top of the details panel
func (d *DetailsPanel) ScrollToTop() {
	d.originY = 0
}

// ScrollToBottom scrolls to the bottom of the details panel
func (d *DetailsPanel) ScrollToBottom() {
	if d.v == nil {
		return
	}

	// Get actual content lines from the rendered view buffer
	contentLines := len(d.v.ViewBufferLines())
	_, viewHeight := d.v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	d.originY = maxOrigin
}

// UpdateFromMigration updates the details panel with migration information
func (d *DetailsPanel) UpdateFromMigration(migration *prisma.Migration, tabName string) {
	// Only reset scroll position for Details tab if viewing a different migration
	if migration != nil && d.currentMigrationName != migration.Name {
		// Reset Details tab scroll position only
		d.tabOriginY[d.tr.TabDetails] = 0
		// If currently on Details tab, also update originY
		if d.tabIndex < len(d.tabs) && d.tabs[d.tabIndex] == d.tr.TabDetails {
			d.originY = 0
		}
		d.currentMigrationName = migration.Name
	} else if migration == nil {
		// Reset Details tab scroll position only
		d.tabOriginY[d.tr.TabDetails] = 0
		// If currently on Details tab, also update originY
		if d.tabIndex < len(d.tabs) && d.tabs[d.tabIndex] == d.tr.TabDetails {
			d.originY = 0
		}
		d.currentMigrationName = ""
	}

	if migration == nil {
		d.content = d.tr.DetailsPanelInitialPlaceholder
		return
	}

	// Handle different cases (priority: Failed > DB-Only > Checksum Mismatch > Empty)

	// In-Transaction migrations (highest priority)
	if migration.IsFailed {
		timestamp, name := parseMigrationName(migration.Name)
		header := fmt.Sprintf(d.tr.DetailsNameLabel, Cyan(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		if migration.Path != "" {
			header += fmt.Sprintf(d.tr.DetailsPathLabel, getRelativePath(migration.Path))
		}
		header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n", Cyan(d.tr.MigrationStatusInTransaction))

		// Show down migration availability
		if migration.HasDownSQL {
			header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Green(d.tr.DetailsDownMigrationAvailable))
		} else {
			header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Red(d.tr.DetailsDownMigrationNotAvailable))
		}

		// Show started_at if available
		if migration.StartedAt != nil {
			header += fmt.Sprintf(d.tr.DetailsStartedAtLabel+"%s\n", migration.StartedAt.Format("2006-01-02 15:04:05"))
		}

		header += "\n" + Yellow(d.tr.DetailsInTransactionWarning)
		header += "\n" + Yellow(d.tr.DetailsNoAdditionalMigrationsWarning)
		header += "\n\n" + d.tr.DetailsResolveManuallyInstruction

		// Show logs if available
		if migration.Logs != nil && *migration.Logs != "" {
			header += "\n" + d.tr.DetailsErrorLogsLabel + "\n" + Red(*migration.Logs)
		}

		// Read and show migration.sql content (if Path is available - not DB-Only)
		if migration.Path != "" {
			sqlPath := filepath.Join(migration.Path, "migration.sql")
			content, err := os.ReadFile(sqlPath)
			if err == nil {
				highlightedSQL := highlightSQL(string(content))
				d.content = header + "\n\n" + highlightedSQL

				// Show down.sql if available
				if migration.HasDownSQL {
					downSQLPath := filepath.Join(migration.Path, "down.sql")
					downContent, err := os.ReadFile(downSQLPath)
					if err == nil {
						highlightedDownSQL := highlightSQL(string(downContent))
						d.content += "\n\n" + Yellow(d.tr.DetailsDownMigrationSQLLabel) + "\n\n" + highlightedDownSQL
					}
				}
			} else {
				d.content = header
			}
		} else {
			d.content = header
		}
		return
	}

	if tabName == "DB-Only" {
		timestamp, name := parseMigrationName(migration.Name)
		header := fmt.Sprintf(d.tr.DetailsNameLabel, Yellow(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n\n", Red(d.tr.MigrationStatusDBOnly))
		header += d.tr.DetailsDBOnlyDescription
		d.content = header
		return
	}

	// Checksum mismatch
	if migration.ChecksumMismatch {
		timestamp, name := parseMigrationName(migration.Name)
		header := fmt.Sprintf(d.tr.DetailsNameLabel, Orange(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		if migration.Path != "" {
			header += fmt.Sprintf(d.tr.DetailsPathLabel, getRelativePath(migration.Path))
		}
		// Show Applied status with Checksum Mismatch warning
		statusLine := fmt.Sprintf(d.tr.DetailsStatusLabel+"%s", Green(d.tr.MigrationStatusApplied))
		if migration.AppliedAt != nil {
			statusLine += fmt.Sprintf(" (%s)", fmt.Sprintf(d.tr.DetailsAppliedAtLabel, migration.AppliedAt.Format("2006-01-02 15:04:05")))
		}
		statusLine += fmt.Sprintf(" - %s\n", Orange(d.tr.MigrationStatusChecksumMismatch))
		header += statusLine

		// Show down migration availability
		if migration.HasDownSQL {
			header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Green(d.tr.DetailsDownMigrationAvailable))
		} else {
			header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Red(d.tr.DetailsDownMigrationNotAvailable))
		}

		header += "\n" + d.tr.DetailsChecksumModifiedDescription
		header += d.tr.DetailsChecksumIssuesWarning

		// Show checksum values (in orange for emphasis)
		header += fmt.Sprintf(d.tr.DetailsLocalChecksumLabel+"%s\n", Orange(migration.Checksum))
		header += fmt.Sprintf(d.tr.DetailsHistoryChecksumLabel+"%s\n", Orange(migration.DBChecksum))

		// Read and show migration.sql content
		sqlPath := filepath.Join(migration.Path, "migration.sql")
		content, err := os.ReadFile(sqlPath)
		if err == nil {
			highlightedSQL := highlightSQL(string(content))
			d.content = header + "\n" + highlightedSQL

			// Show down.sql if available
			if migration.HasDownSQL {
				downSQLPath := filepath.Join(migration.Path, "down.sql")
				downContent, err := os.ReadFile(downSQLPath)
				if err == nil {
					highlightedDownSQL := highlightSQL(string(downContent))
					d.content += "\n\n" + Yellow(d.tr.DetailsDownMigrationSQLLabel) + "\n\n" + highlightedDownSQL
				}
			}
		} else {
			d.content = header
		}
		return
	}

	if migration.IsEmpty {
		timestamp, name := parseMigrationName(migration.Name)
		header := fmt.Sprintf(d.tr.DetailsNameLabel, Magenta(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		if migration.Path != "" {
			header += fmt.Sprintf(d.tr.DetailsPathLabel, getRelativePath(migration.Path))
		}
		header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n", Red(d.tr.MigrationStatusEmptyMigration))

		// Show down migration availability (even for empty migrations)
		if migration.HasDownSQL {
			header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Green(d.tr.DetailsDownMigrationAvailable))
		} else {
			header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Red(d.tr.DetailsDownMigrationNotAvailable))
		}

		header += "\n" + d.tr.DetailsEmptyMigrationDescription
		header += d.tr.DetailsEmptyMigrationWarning
		d.content = header
		return
	}

	// Read migration.sql content
	sqlPath := filepath.Join(migration.Path, "migration.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		timestamp, name := parseMigrationName(migration.Name)
		d.content = fmt.Sprintf(d.tr.DetailsNameLabel, name) +
			fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp) +
			"\n" + fmt.Sprintf(d.tr.ErrorReadingMigrationSQL, err)
		return
	}

	// Build header with status
	timestamp, name := parseMigrationName(migration.Name)
	var header string
	if migration.AppliedAt != nil {
		header = fmt.Sprintf(d.tr.DetailsNameLabel, Green(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		if migration.Path != "" {
			header += fmt.Sprintf(d.tr.DetailsPathLabel, getRelativePath(migration.Path))
		}
		header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s (%s)\n",
			Green(d.tr.MigrationStatusApplied),
			fmt.Sprintf(d.tr.DetailsAppliedAtLabel, migration.AppliedAt.Format("2006-01-02 15:04:05")))
	} else {
		header = fmt.Sprintf(d.tr.DetailsNameLabel, Yellow(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		if migration.Path != "" {
			header += fmt.Sprintf(d.tr.DetailsPathLabel, getRelativePath(migration.Path))
		}
		header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n", Yellow(d.tr.MigrationStatusPending))
	}

	// Show down migration availability
	if migration.HasDownSQL {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Green(d.tr.DetailsDownMigrationAvailable))
	} else {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", Red(d.tr.DetailsDownMigrationNotAvailable))
	}

	// Apply syntax highlighting to SQL content
	highlightedSQL := highlightSQL(string(content))

	d.content = header + "\n" + highlightedSQL

	// Show down.sql if available
	if migration.HasDownSQL {
		downSQLPath := filepath.Join(migration.Path, "down.sql")
		downContent, err := os.ReadFile(downSQLPath)
		if err == nil {
			highlightedDownSQL := highlightSQL(string(downContent))
			d.content += "\n\n" + Yellow(d.tr.DetailsDownMigrationSQLLabel) + "\n\n" + highlightedDownSQL
		}
	}
}

// parseMigrationName parses a Prisma migration name into timestamp and description
// Expected format: YYYYMMDDHHMMSS_description
// Example: 20231123052950_create_career_table -> "2023-11-23 05:29:50", "create_career_table"
func parseMigrationName(fullName string) (timestamp, name string) {
	// Check if name matches expected format (at least 15 chars with underscore at position 14)
	if len(fullName) > 15 && fullName[14] == '_' {
		timestampStr := fullName[:14] // "20231123052950"
		name = fullName[15:]          // "create_career_table"

		// Parse timestamp: YYYYMMDDHHMMSS -> YYYY-MM-DD HH:MM:SS
		if len(timestampStr) == 14 {
			timestamp = fmt.Sprintf("%s-%s-%s %s:%s:%s",
				timestampStr[0:4],   // YYYY
				timestampStr[4:6],   // MM
				timestampStr[6:8],   // DD
				timestampStr[8:10],  // HH
				timestampStr[10:12], // mm
				timestampStr[12:14]) // ss
			return timestamp, name
		}
	}

	// Fallback: couldn't parse, return as-is
	return "N/A", fullName
}

// getRelativePath converts absolute path to relative path from current working directory
func getRelativePath(absPath string) string {
	if absPath == "" {
		return ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		return absPath // Fallback to absolute path
	}

	relPath, err := filepath.Rel(cwd, absPath)
	if err != nil {
		return absPath // Fallback to absolute path
	}

	return relPath
}

// LoadActionNeededData loads migrations that require action (Empty + Mismatch) and validates schema
func (d *DetailsPanel) LoadActionNeededData() {
	if d.migrationsPanel == nil {
		d.actionNeededMigrations = []prisma.Migration{}
		d.validationResult = nil
		d.updateTabs()
		return
	}

	// Collect Empty and Mismatch migrations from Local category
	var actionNeeded []prisma.Migration
	for _, mig := range d.migrationsPanel.category.Local {
		if mig.IsEmpty || mig.ChecksumMismatch {
			actionNeeded = append(actionNeeded, mig)
		}
	}

	d.actionNeededMigrations = actionNeeded

	// Run schema validation
	cwd, err := os.Getwd()
	if err == nil {
		validateResult, err := prisma.Validate(cwd)
		if err == nil {
			d.validationResult = validateResult
		} else {
			d.validationResult = nil
		}
	} else {
		d.validationResult = nil
	}

	d.updateTabs()
}

// updateTabs rebuilds the tabs list based on available data
func (d *DetailsPanel) updateTabs() {
	// Always have Details tab
	d.tabs = []string{d.tr.TabDetails}

	// Add Action-Needed tab if there are migration issues or validation errors
	hasIssues := len(d.actionNeededMigrations) > 0
	hasValidationErrors := d.validationResult != nil && !d.validationResult.Valid

	if hasIssues || hasValidationErrors {
		d.tabs = append(d.tabs, d.tr.TabActionNeeded)
	}

	// Reset tab index if current tab no longer exists
	if d.tabIndex >= len(d.tabs) {
		d.tabIndex = 0
	}
}

// saveCurrentTabState saves the current scroll position
func (d *DetailsPanel) saveCurrentTabState() {
	if d.tabIndex >= len(d.tabs) {
		return
	}
	tabName := d.tabs[d.tabIndex]
	d.tabOriginY[tabName] = d.originY
}

// restoreTabState restores the scroll position for the current tab
func (d *DetailsPanel) restoreTabState() {
	if d.tabIndex >= len(d.tabs) {
		return
	}
	tabName := d.tabs[d.tabIndex]
	if prevOriginY, exists := d.tabOriginY[tabName]; exists {
		d.originY = prevOriginY
	} else {
		d.originY = 0
	}
}

// NextTab switches to the next tab
func (d *DetailsPanel) NextTab() {
	if len(d.tabs) == 0 {
		return
	}
	// Save current tab state before switching
	d.saveCurrentTabState()

	d.tabIndex = (d.tabIndex + 1) % len(d.tabs)

	// Restore scroll position for new tab
	d.restoreTabState()
}

// PrevTab switches to the previous tab
func (d *DetailsPanel) PrevTab() {
	if len(d.tabs) == 0 {
		return
	}
	// Save current tab state before switching
	d.saveCurrentTabState()

	d.tabIndex = (d.tabIndex - 1 + len(d.tabs)) % len(d.tabs)

	// Restore scroll position for new tab
	d.restoreTabState()
}

// SetApp sets the app reference for modal checking
func (d *DetailsPanel) SetApp(app *App) {
	d.app = app
}

// handleTabClick handles mouse click on tab bar
func (d *DetailsPanel) handleTabClick(tabIndex int) error {
	// Ignore if modal is active
	if d.app != nil && d.app.HasActiveModal() {
		return nil
	}

	// First, switch focus to this panel if not already focused
	if d.app != nil {
		if err := d.app.handlePanelClick(ViewDetails); err != nil {
			return err
		}
	}

	// Ignore if same tab is clicked
	if tabIndex == d.tabIndex {
		return nil
	}

	// Ignore if tab index is out of bounds
	if tabIndex < 0 || tabIndex >= len(d.tabs) {
		return nil
	}

	// Save current tab state
	d.saveCurrentTabState()

	// Switch to clicked tab
	d.tabIndex = tabIndex

	// Restore scroll position for new tab
	d.restoreTabState()

	return nil
}
