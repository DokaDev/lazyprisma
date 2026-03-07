package context

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// ============================================================================
// Self-contained ANSI styling helpers (avoid importing pkg/app for colours)
// ============================================================================

func detailsRed(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", text)
}

func detailsGreen(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", text)
}

func detailsYellow(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[33m%s\x1b[0m", text)
}

func detailsMagenta(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[35m%s\x1b[0m", text)
}

func detailsCyan(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[36m%s\x1b[0m", text)
}

func detailsOrange(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[38;5;208m%s\x1b[0m", text)
}

// detailsStylize applies combined ANSI styling (fg colour + bold).
func detailsStylize(text string, fgCode string, bold bool) string {
	if text == "" {
		return text
	}
	codes := make([]string, 0, 2)
	if fgCode != "" {
		codes = append(codes, fgCode)
	}
	if bold {
		codes = append(codes, "1")
	}
	if len(codes) == 0 {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", strings.Join(codes, ";"), text)
}

// Frame and title styling constants (matching app.panel.go values)
var (
	detailsDefaultFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	detailsPrimaryFrameColor = gocui.ColorWhite
	detailsFocusedFrameColor = gocui.ColorGreen

	detailsPrimaryTitleColor = gocui.ColorWhite | gocui.AttrNone
	detailsFocusedTitleColor = gocui.ColorGreen | gocui.AttrBold

	detailsFocusedActiveTabColor = gocui.ColorGreen | gocui.AttrBold
	detailsPrimaryActiveTabColor = gocui.ColorGreen | gocui.AttrNone
)

// DetailsContext is the context-based replacement for DetailsPanel.
// It displays migration details and action-needed information with tabbed navigation.
type DetailsContext struct {
	*SimpleContext
	*ScrollableTrait
	*TabbedTrait

	g  *gocui.Gui
	tr *i18n.TranslationSet

	// Content fields
	content              string
	currentMigrationName string

	// Action-needed data
	actionNeededMigrations []prisma.Migration
	validationResult       *prisma.ValidateResult

	// UI state
	focused bool

	// Callback-based decoupling (replaces direct App reference)
	hasActiveModal func() bool
	onPanelClick   func(viewID string)
}

var _ types.Context = &DetailsContext{}
var _ types.IScrollableContext = &DetailsContext{}

// DetailsContextOpts holds the options for creating a DetailsContext.
type DetailsContextOpts struct {
	Gui      *gocui.Gui
	Tr       *i18n.TranslationSet
	ViewName string
}

// NewDetailsContext creates a new DetailsContext.
func NewDetailsContext(opts DetailsContextOpts) *DetailsContext {
	baseCtx := NewBaseContext(BaseContextOpts{
		Key:       types.ContextKey(opts.ViewName),
		Kind:      types.MAIN_CONTEXT,
		ViewName:  opts.ViewName,
		Focusable: true,
		Title:     opts.Tr.PanelTitleDetails,
	})

	simpleCtx := NewSimpleContext(baseCtx)

	tabbedTrait := NewTabbedTrait([]string{opts.Tr.TabDetails})

	dc := &DetailsContext{
		SimpleContext:          simpleCtx,
		ScrollableTrait:        &ScrollableTrait{},
		TabbedTrait:            &tabbedTrait,
		g:                      opts.Gui,
		tr:                     opts.Tr,
		content:                opts.Tr.DetailsPanelInitialPlaceholder,
		actionNeededMigrations: []prisma.Migration{},
	}

	return dc
}

// ID returns the view identifier (implements Panel interface from app package).
func (d *DetailsContext) ID() string {
	return d.GetViewName()
}

// SetModalCallbacks sets callbacks that replace the direct App reference.
func (d *DetailsContext) SetModalCallbacks(hasActiveModal func() bool, onPanelClick func(string)) {
	d.hasActiveModal = hasActiveModal
	d.onPanelClick = onPanelClick
}

// Draw renders the details panel (implements Panel interface from app package).
func (d *DetailsContext) Draw(dim boxlayout.Dimensions) error {
	v, err := d.g.SetView(d.GetViewName(), dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	// Setup view WITHOUT title (tabs replace title)
	d.SetView(v)                  // BaseContext
	d.ScrollableTrait.SetView(v)  // ScrollableTrait

	v.Clear()
	v.Frame = true
	v.FrameRunes = detailsDefaultFrameRunes
	v.Wrap = true // Enable word wrap for long lines

	// Set tabs from TabbedTrait
	v.Tabs = d.TabbedTrait.GetTabs()
	v.TabIndex = d.TabbedTrait.GetCurrentTabIdx()

	// Set frame and tab colors based on focus
	tabs := d.TabbedTrait.GetTabs()
	if d.focused {
		v.FrameColor = detailsFocusedFrameColor
		v.TitleColor = detailsFocusedTitleColor
		if len(tabs) == 1 {
			v.SelFgColor = detailsFocusedTitleColor // Single tab: treat like title
		} else {
			v.SelFgColor = detailsFocusedActiveTabColor // Multiple tabs: use active tab color
		}
	} else {
		v.FrameColor = detailsPrimaryFrameColor
		v.TitleColor = detailsPrimaryTitleColor
		if len(tabs) == 1 {
			v.SelFgColor = detailsPrimaryTitleColor // Single tab: treat like title
		} else {
			v.SelFgColor = detailsPrimaryActiveTabColor // Multiple tabs: use active tab color
		}
	}

	// Render content based on current tab
	currentTab := d.TabbedTrait.GetCurrentTab()
	if currentTab == d.tr.TabActionNeeded {
		fmt.Fprint(v, d.buildActionNeededContent())
	} else {
		fmt.Fprint(v, d.content)
	}

	// Adjust scroll and apply origin
	d.ScrollableTrait.AdjustScroll()

	return nil
}

// OnFocus handles focus gain (implements Panel interface from app package).
func (d *DetailsContext) OnFocus() {
	d.focused = true
	if v := d.GetView(); v != nil {
		v.FrameColor = detailsFocusedFrameColor
		v.TitleColor = detailsFocusedTitleColor
	}
}

// OnBlur handles focus loss (implements Panel interface from app package).
func (d *DetailsContext) OnBlur() {
	d.focused = false
	if v := d.GetView(); v != nil {
		v.FrameColor = detailsPrimaryFrameColor
		v.TitleColor = detailsPrimaryTitleColor
	}
}

// SetContent sets the content text directly.
func (d *DetailsContext) SetContent(content string) {
	d.content = content
}

// UpdateFromMigration updates the details panel with migration information.
func (d *DetailsContext) UpdateFromMigration(migration *prisma.Migration, tabName string) {
	// Only reset scroll position for Details tab if viewing a different migration
	if migration != nil && d.currentMigrationName != migration.Name {
		// Reset Details tab scroll position only
		d.TabbedTrait.tabOriginY[d.tabIdxByName(d.tr.TabDetails)] = 0
		// If currently on Details tab, also update originY
		if d.TabbedTrait.GetCurrentTab() == d.tr.TabDetails {
			d.ScrollableTrait.SetOriginY(0)
		}
		d.currentMigrationName = migration.Name
	} else if migration == nil {
		// Reset Details tab scroll position only
		d.TabbedTrait.tabOriginY[d.tabIdxByName(d.tr.TabDetails)] = 0
		// If currently on Details tab, also update originY
		if d.TabbedTrait.GetCurrentTab() == d.tr.TabDetails {
			d.ScrollableTrait.SetOriginY(0)
		}
		d.currentMigrationName = ""
	}

	if migration == nil {
		d.content = d.tr.DetailsPanelInitialPlaceholder
		return
	}

	d.content = d.buildMigrationDetailContent(migration, tabName)
}

// buildMigrationDetailContent builds the detail content for a given migration.
func (d *DetailsContext) buildMigrationDetailContent(migration *prisma.Migration, tabName string) string {
	// Handle different cases (priority: Failed > DB-Only > Checksum Mismatch > Empty)

	// In-Transaction migrations (highest priority)
	if migration.IsFailed {
		return d.buildFailedMigrationContent(migration)
	}

	if tabName == "DB-Only" {
		return d.buildDBOnlyContent(migration)
	}

	// Checksum mismatch
	if migration.ChecksumMismatch {
		return d.buildChecksumMismatchContent(migration)
	}

	if migration.IsEmpty {
		return d.buildEmptyMigrationContent(migration)
	}

	// Normal migration
	return d.buildNormalMigrationContent(migration)
}

// buildFailedMigrationContent builds content for failed/in-transaction migrations.
func (d *DetailsContext) buildFailedMigrationContent(migration *prisma.Migration) string {
	timestamp, name := detailsParseMigrationName(migration.Name)
	header := fmt.Sprintf(d.tr.DetailsNameLabel, detailsCyan(name))
	header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
	if migration.Path != "" {
		header += fmt.Sprintf(d.tr.DetailsPathLabel, detailsGetRelativePath(migration.Path))
	}
	header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n", detailsCyan(d.tr.MigrationStatusInTransaction))

	// Show down migration availability
	if migration.HasDownSQL {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsGreen(d.tr.DetailsDownMigrationAvailable))
	} else {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsRed(d.tr.DetailsDownMigrationNotAvailable))
	}

	// Show started_at if available
	if migration.StartedAt != nil {
		header += fmt.Sprintf(d.tr.DetailsStartedAtLabel+"%s\n", migration.StartedAt.Format("2006-01-02 15:04:05"))
	}

	header += "\n" + detailsYellow(d.tr.DetailsInTransactionWarning)
	header += "\n" + detailsYellow(d.tr.DetailsNoAdditionalMigrationsWarning)
	header += "\n\n" + d.tr.DetailsResolveManuallyInstruction

	// Show logs if available
	if migration.Logs != nil && *migration.Logs != "" {
		header += "\n" + d.tr.DetailsErrorLogsLabel + "\n" + detailsRed(*migration.Logs)
	}

	// Read and show migration.sql content (if Path is available - not DB-Only)
	if migration.Path != "" {
		sqlPath := filepath.Join(migration.Path, "migration.sql")
		content, err := os.ReadFile(sqlPath)
		if err == nil {
			highlightedSQL := detailsHighlightSQL(string(content))
			result := header + "\n\n" + highlightedSQL

			// Show down.sql if available
			if migration.HasDownSQL {
				downSQLPath := filepath.Join(migration.Path, "down.sql")
				downContent, err := os.ReadFile(downSQLPath)
				if err == nil {
					highlightedDownSQL := detailsHighlightSQL(string(downContent))
					result += "\n\n" + detailsYellow(d.tr.DetailsDownMigrationSQLLabel) + "\n\n" + highlightedDownSQL
				}
			}
			return result
		}
	}

	return header
}

// buildDBOnlyContent builds content for DB-only migrations.
func (d *DetailsContext) buildDBOnlyContent(migration *prisma.Migration) string {
	timestamp, name := detailsParseMigrationName(migration.Name)
	header := fmt.Sprintf(d.tr.DetailsNameLabel, detailsYellow(name))
	header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
	header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n\n", detailsRed(d.tr.MigrationStatusDBOnly))
	header += d.tr.DetailsDBOnlyDescription
	return header
}

// buildChecksumMismatchContent builds content for checksum mismatch migrations.
func (d *DetailsContext) buildChecksumMismatchContent(migration *prisma.Migration) string {
	timestamp, name := detailsParseMigrationName(migration.Name)
	header := fmt.Sprintf(d.tr.DetailsNameLabel, detailsOrange(name))
	header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
	if migration.Path != "" {
		header += fmt.Sprintf(d.tr.DetailsPathLabel, detailsGetRelativePath(migration.Path))
	}
	// Show Applied status with Checksum Mismatch warning
	statusLine := fmt.Sprintf(d.tr.DetailsStatusLabel+"%s", detailsGreen(d.tr.MigrationStatusApplied))
	if migration.AppliedAt != nil {
		statusLine += fmt.Sprintf(" (%s)", fmt.Sprintf(d.tr.DetailsAppliedAtLabel, migration.AppliedAt.Format("2006-01-02 15:04:05")))
	}
	statusLine += fmt.Sprintf(" - %s\n", detailsOrange(d.tr.MigrationStatusChecksumMismatch))
	header += statusLine

	// Show down migration availability
	if migration.HasDownSQL {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsGreen(d.tr.DetailsDownMigrationAvailable))
	} else {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsRed(d.tr.DetailsDownMigrationNotAvailable))
	}

	header += "\n" + d.tr.DetailsChecksumModifiedDescription
	header += d.tr.DetailsChecksumIssuesWarning

	// Show checksum values (in orange for emphasis)
	header += fmt.Sprintf(d.tr.DetailsLocalChecksumLabel+"%s\n", detailsOrange(migration.Checksum))
	header += fmt.Sprintf(d.tr.DetailsHistoryChecksumLabel+"%s\n", detailsOrange(migration.DBChecksum))

	// Read and show migration.sql content
	sqlPath := filepath.Join(migration.Path, "migration.sql")
	content, err := os.ReadFile(sqlPath)
	if err == nil {
		highlightedSQL := detailsHighlightSQL(string(content))
		result := header + "\n" + highlightedSQL

		// Show down.sql if available
		if migration.HasDownSQL {
			downSQLPath := filepath.Join(migration.Path, "down.sql")
			downContent, err := os.ReadFile(downSQLPath)
			if err == nil {
				highlightedDownSQL := detailsHighlightSQL(string(downContent))
				result += "\n\n" + detailsYellow(d.tr.DetailsDownMigrationSQLLabel) + "\n\n" + highlightedDownSQL
			}
		}
		return result
	}

	return header
}

// buildEmptyMigrationContent builds content for empty migrations.
func (d *DetailsContext) buildEmptyMigrationContent(migration *prisma.Migration) string {
	timestamp, name := detailsParseMigrationName(migration.Name)
	header := fmt.Sprintf(d.tr.DetailsNameLabel, detailsMagenta(name))
	header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
	if migration.Path != "" {
		header += fmt.Sprintf(d.tr.DetailsPathLabel, detailsGetRelativePath(migration.Path))
	}
	header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n", detailsRed(d.tr.MigrationStatusEmptyMigration))

	// Show down migration availability (even for empty migrations)
	if migration.HasDownSQL {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsGreen(d.tr.DetailsDownMigrationAvailable))
	} else {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsRed(d.tr.DetailsDownMigrationNotAvailable))
	}

	header += "\n" + d.tr.DetailsEmptyMigrationDescription
	header += d.tr.DetailsEmptyMigrationWarning
	return header
}

// buildNormalMigrationContent builds content for normal (applied/pending) migrations.
func (d *DetailsContext) buildNormalMigrationContent(migration *prisma.Migration) string {
	// Read migration.sql content
	sqlPath := filepath.Join(migration.Path, "migration.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		timestamp, name := detailsParseMigrationName(migration.Name)
		return fmt.Sprintf(d.tr.DetailsNameLabel, name) +
			fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp) +
			"\n" + fmt.Sprintf(d.tr.ErrorReadingMigrationSQL, err)
	}

	// Build header with status
	timestamp, name := detailsParseMigrationName(migration.Name)
	var header string
	if migration.AppliedAt != nil {
		header = fmt.Sprintf(d.tr.DetailsNameLabel, detailsGreen(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		if migration.Path != "" {
			header += fmt.Sprintf(d.tr.DetailsPathLabel, detailsGetRelativePath(migration.Path))
		}
		header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s (%s)\n",
			detailsGreen(d.tr.MigrationStatusApplied),
			fmt.Sprintf(d.tr.DetailsAppliedAtLabel, migration.AppliedAt.Format("2006-01-02 15:04:05")))
	} else {
		header = fmt.Sprintf(d.tr.DetailsNameLabel, detailsYellow(name))
		header += fmt.Sprintf(d.tr.DetailsTimestampLabel, timestamp)
		if migration.Path != "" {
			header += fmt.Sprintf(d.tr.DetailsPathLabel, detailsGetRelativePath(migration.Path))
		}
		header += fmt.Sprintf(d.tr.DetailsStatusLabel+"%s\n", detailsYellow(d.tr.MigrationStatusPending))
	}

	// Show down migration availability
	if migration.HasDownSQL {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsGreen(d.tr.DetailsDownMigrationAvailable))
	} else {
		header += fmt.Sprintf(d.tr.DetailsDownMigrationLabel+"%s\n", detailsRed(d.tr.DetailsDownMigrationNotAvailable))
	}

	// Apply syntax highlighting to SQL content
	highlightedSQL := detailsHighlightSQL(string(content))

	result := header + "\n" + highlightedSQL

	// Show down.sql if available
	if migration.HasDownSQL {
		downSQLPath := filepath.Join(migration.Path, "down.sql")
		downContent, err := os.ReadFile(downSQLPath)
		if err == nil {
			highlightedDownSQL := detailsHighlightSQL(string(downContent))
			result += "\n\n" + detailsYellow(d.tr.DetailsDownMigrationSQLLabel) + "\n\n" + highlightedDownSQL
		}
	}

	return result
}

// SetActionNeededMigrations receives migration data from outside (App will call this).
// This replaces the old pattern of directly accessing MigrationsPanel.
func (d *DetailsContext) SetActionNeededMigrations(migrations []prisma.Migration) {
	d.actionNeededMigrations = migrations
}

// LoadActionNeededData loads action-needed data using the internal migrations list and validates schema.
func (d *DetailsContext) LoadActionNeededData() {
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

// updateTabs rebuilds the tabs list based on available data.
func (d *DetailsContext) updateTabs() {
	// Always have Details tab
	newTabs := []string{d.tr.TabDetails}

	// Add Action-Needed tab if there are migration issues or validation errors
	hasIssues := len(d.actionNeededMigrations) > 0
	hasValidationErrors := d.validationResult != nil && !d.validationResult.Valid

	if hasIssues || hasValidationErrors {
		newTabs = append(newTabs, d.tr.TabActionNeeded)
	}

	d.TabbedTrait.SetTabs(newTabs)
}

// buildActionNeededContent builds the content for the Action-Needed tab.
func (d *DetailsContext) buildActionNeededContent() string {
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
	content.WriteString(fmt.Sprintf("%s (%d%s", detailsYellow(d.tr.ActionNeededHeader), totalCount, d.tr.ActionNeededIssueSingular))
	if totalCount > 1 {
		content.WriteString(d.tr.ActionNeededIssuePlural)
	}
	content.WriteString(")\n\n")

	// Empty Migrations Section
	if emptyCount > 0 {
		content.WriteString(strings.Repeat("━", 40) + "\n")
		content.WriteString(fmt.Sprintf("%s (%d)\n", detailsRed(d.tr.ActionNeededEmptyMigrationsHeader), emptyCount))
		content.WriteString(strings.Repeat("━", 40) + "\n\n")

		content.WriteString(d.tr.ActionNeededEmptyDescription)

		content.WriteString(d.tr.ActionNeededAffectedLabel)
		for _, mig := range emptyMigrations {
			_, name := detailsParseMigrationName(mig.Name)
			content.WriteString(fmt.Sprintf("  • %s\n", detailsRed(name)))
		}

		content.WriteString("\n" + d.tr.ActionNeededRecommendedLabel)
		content.WriteString(d.tr.ActionNeededAddMigrationSQL)
		content.WriteString(d.tr.ActionNeededDeleteEmptyFolders)
		content.WriteString(d.tr.ActionNeededMarkAsBaseline)
	}

	// Checksum Mismatch Section
	if mismatchCount > 0 {
		content.WriteString(strings.Repeat("━", 40) + "\n")
		content.WriteString(fmt.Sprintf("%s (%d)\n", detailsOrange(d.tr.ActionNeededChecksumMismatchHeader), mismatchCount))
		content.WriteString(strings.Repeat("━", 40) + "\n\n")

		content.WriteString(d.tr.ActionNeededChecksumModifiedDescription)

		content.WriteString(detailsYellow(d.tr.ActionNeededWarningPrefix))
		content.WriteString(d.tr.ActionNeededEditingInconsistenciesWarning)

		content.WriteString(d.tr.ActionNeededAffectedLabel)
		for _, mig := range mismatchMigrations {
			_, name := detailsParseMigrationName(mig.Name)
			content.WriteString(fmt.Sprintf("  • %s\n", detailsOrange(name)))
		}

		content.WriteString("\n" + d.tr.ActionNeededRecommendedLabel)
		content.WriteString(d.tr.ActionNeededRevertLocalChanges)
		content.WriteString(d.tr.ActionNeededCreateNewInstead)
		content.WriteString(d.tr.ActionNeededContactTeamIfNeeded)
	}

	// Schema Validation Section
	if validationErrorCount > 0 {
		content.WriteString(strings.Repeat("━", 40) + "\n")
		content.WriteString(fmt.Sprintf("%s (%d)\n", detailsRed(d.tr.ActionNeededSchemaValidationErrorsHeader), validationErrorCount))
		content.WriteString(strings.Repeat("━", 40) + "\n\n")

		content.WriteString(d.tr.ActionNeededSchemaValidationFailedDesc)
		content.WriteString(d.tr.ActionNeededFixBeforeMigration)

		// Show full validation output (contains detailed error info)
		if d.validationResult.Output != "" {
			content.WriteString(detailsStylize(d.tr.ActionNeededValidationOutputLabel, "33", true) + "\n")
			// Display the full output with proper formatting (preserve all line breaks)
			outputLines := strings.Split(d.validationResult.Output, "\n")
			for _, line := range outputLines {
				// Highlight error lines
				if strings.Contains(line, "Error:") || strings.Contains(line, "error:") {
					content.WriteString(detailsRed(line) + "\n")
				} else if strings.Contains(line, "-->") {
					content.WriteString(detailsYellow(line) + "\n")
				} else {
					// Preserve empty lines and all other content as-is
					content.WriteString(line + "\n")
				}
			}
			content.WriteString("\n")
		}

		content.WriteString(detailsStylize(d.tr.ActionNeededRecommendedActionsLabel, "33", true) + "\n")
		content.WriteString(d.tr.ActionNeededFixSchemaErrors)
		content.WriteString(d.tr.ActionNeededCheckLineNumbers)
		content.WriteString(d.tr.ActionNeededReferPrismaDocumentation)
	}

	return content.String()
}

// NextTab switches to the next tab with scroll state save/restore.
func (d *DetailsContext) NextTab() {
	if len(d.TabbedTrait.GetTabs()) == 0 {
		return
	}
	// Save current scroll position before switching
	d.TabbedTrait.SaveTabOriginY(d.ScrollableTrait.GetOriginY())
	d.TabbedTrait.NextTab()
	// Restore scroll position for new tab
	d.ScrollableTrait.SetOriginY(d.TabbedTrait.RestoreTabOriginY())
}

// PrevTab switches to the previous tab with scroll state save/restore.
func (d *DetailsContext) PrevTab() {
	if len(d.TabbedTrait.GetTabs()) == 0 {
		return
	}
	// Save current scroll position before switching
	d.TabbedTrait.SaveTabOriginY(d.ScrollableTrait.GetOriginY())
	d.TabbedTrait.PrevTab()
	// Restore scroll position for new tab
	d.ScrollableTrait.SetOriginY(d.TabbedTrait.RestoreTabOriginY())
}

// handleTabClick handles mouse click on tab bar.
func (d *DetailsContext) HandleTabClick(tabIndex int) error {
	// Ignore if modal is active
	if d.hasActiveModal != nil && d.hasActiveModal() {
		return nil
	}

	// First, switch focus to this panel if not already focused
	if d.onPanelClick != nil {
		d.onPanelClick(d.GetViewName())
	}

	// Ignore if same tab is clicked
	if tabIndex == d.TabbedTrait.GetCurrentTabIdx() {
		return nil
	}

	// Ignore if tab index is out of bounds
	tabs := d.TabbedTrait.GetTabs()
	if tabIndex < 0 || tabIndex >= len(tabs) {
		return nil
	}

	// Save current tab state
	d.TabbedTrait.SaveTabOriginY(d.ScrollableTrait.GetOriginY())

	// Switch to clicked tab
	d.TabbedTrait.SetCurrentTabIdx(tabIndex)

	// Restore scroll position for new tab
	d.ScrollableTrait.SetOriginY(d.TabbedTrait.RestoreTabOriginY())

	return nil
}

// ScrollUpByWheel scrolls up by wheel increment (delegates to ScrollableTrait).
func (d *DetailsContext) ScrollUpByWheel() {
	d.ScrollableTrait.ScrollUpByWheel()
}

// ScrollDownByWheel scrolls down by wheel increment (delegates to ScrollableTrait).
func (d *DetailsContext) ScrollDownByWheel() {
	d.ScrollableTrait.ScrollDownByWheel()
}

// ============================================================================
// Private helpers
// ============================================================================

// tabIdxByName returns the index of the tab with the given name, or 0 if not found.
func (d *DetailsContext) tabIdxByName(name string) int {
	for i, t := range d.TabbedTrait.GetTabs() {
		if t == name {
			return i
		}
	}
	return 0
}

// detailsHighlightSQL applies syntax highlighting to SQL code with line numbers.
func detailsHighlightSQL(code string) string {
	// Get SQL lexer
	lexer := lexers.Get("sql")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Get style (dracula is a popular dark theme)
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

// detailsParseMigrationName parses a Prisma migration name into timestamp and description.
// Expected format: YYYYMMDDHHMMSS_description
// Example: 20231123052950_create_career_table -> "2023-11-23 05:29:50", "create_career_table"
func detailsParseMigrationName(fullName string) (timestamp, name string) {
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

// detailsGetRelativePath converts absolute path to relative path from current working directory.
func detailsGetRelativePath(absPath string) string {
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
