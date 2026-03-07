package context

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/database"
	"github.com/dokadev/lazyprisma/pkg/gui/style"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// Frame and title styling constants (matching app.panel.go values)
var (
	migDefaultFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	migPrimaryFrameColor = gocui.ColorWhite
	migFocusedFrameColor = gocui.ColorGreen

	migPrimaryTitleColor = gocui.ColorWhite | gocui.AttrNone
	migFocusedTitleColor = gocui.ColorGreen | gocui.AttrBold

	// Tab styling
	migFocusedActiveTabColor = gocui.ColorGreen | gocui.AttrBold
	migPrimaryActiveTabColor = gocui.ColorGreen | gocui.AttrNone

	// List selection colour
	migSelectionBgColor = gocui.ColorBlue
)

// MigrationsContext manages the migrations list with tabs (Local, Pending, DB-Only).
type MigrationsContext struct {
	*SimpleContext
	*ScrollableTrait
	*TabbedTrait

	g       *gocui.Gui
	tr      *i18n.TranslationSet
	focused bool

	// Data
	category    prisma.MigrationCategory // Categorised migrations
	items       []string                 // Current tab's rendered display strings
	selected    int                      // Selected item index in current tab
	dbClient    *database.Client         // Database connection
	dbConnected bool                     // True if connected to database
	tableExists bool                     // True if _prisma_migrations table exists

	// Per-tab state preservation
	tabSelectedMap map[string]int // Last selected index per tab (keyed by tab name)
	tabOriginYMap  map[string]int // Last scroll position per tab (keyed by tab name)

	// Callbacks (replace direct panel/app references)
	onSelectionChanged func(migration *prisma.Migration, tabName string)
	hasActiveModal     func() bool
	onPanelClick       func(viewID string)
}

var _ types.Context = &MigrationsContext{}

type MigrationsContextOpts struct {
	Gui      *gocui.Gui
	Tr       *i18n.TranslationSet
	ViewName string
}

func NewMigrationsContext(opts MigrationsContextOpts) *MigrationsContext {
	baseCtx := NewBaseContext(BaseContextOpts{
		Key:       types.ContextKey(opts.ViewName),
		Kind:      types.SIDE_CONTEXT,
		ViewName:  opts.ViewName,
		Focusable: true,
	})

	simpleCtx := NewSimpleContext(baseCtx)

	mc := &MigrationsContext{
		SimpleContext:  simpleCtx,
		ScrollableTrait: &ScrollableTrait{},
		g:              opts.Gui,
		tr:             opts.Tr,
		items:          []string{},
		selected:       0,
		tabSelectedMap: make(map[string]int),
		tabOriginYMap:  make(map[string]int),
	}

	// Initialise TabbedTrait with empty tabs (loadMigrations will populate)
	tt := NewTabbedTrait([]string{})
	mc.TabbedTrait = &tt

	mc.loadMigrations()
	return mc
}

// ---------------------------------------------------------------------------
// Callback setters
// ---------------------------------------------------------------------------

// SetOnSelectionChanged registers a callback invoked whenever the selected
// migration changes (replaces the old SetDetailsPanel coupling).
func (m *MigrationsContext) SetOnSelectionChanged(cb func(*prisma.Migration, string)) {
	m.onSelectionChanged = cb
}

// SetModalCallbacks registers callbacks for modal and panel-click checks
// (replaces the old SetApp coupling).
func (m *MigrationsContext) SetModalCallbacks(hasActiveModal func() bool, onPanelClick func(string)) {
	m.hasActiveModal = hasActiveModal
	m.onPanelClick = onPanelClick
}

// ---------------------------------------------------------------------------
// Public accessors
// ---------------------------------------------------------------------------

// ID returns the view identifier (Panel interface compatibility).
func (m *MigrationsContext) ID() string {
	return m.GetViewName()
}

// GetSelectedMigration returns the currently selected migration, or nil.
func (m *MigrationsContext) GetSelectedMigration() *prisma.Migration {
	tabName := m.TabbedTrait.GetCurrentTab()
	if tabName == "" {
		return nil
	}

	migrations := m.migrationsForTab(tabName)

	if m.selected >= 0 && m.selected < len(migrations) {
		return &migrations[m.selected]
	}
	return nil
}

// GetCurrentTabName returns the name of the active tab.
func (m *MigrationsContext) GetCurrentTabName() string {
	return m.TabbedTrait.GetCurrentTab()
}

// GetCategory exposes the full migration category for external use.
func (m *MigrationsContext) GetCategory() prisma.MigrationCategory {
	return m.category
}

// IsDBConnected returns whether the database connection is active.
func (m *MigrationsContext) IsDBConnected() bool {
	return m.dbConnected
}

// ---------------------------------------------------------------------------
// Focus / Blur
// ---------------------------------------------------------------------------

// OnFocus handles focus gain (Panel interface compatibility).
func (m *MigrationsContext) OnFocus() {
	m.focused = true
	if v := m.BaseContext.GetView(); v != nil {
		v.FrameColor = migFocusedFrameColor
		v.TitleColor = migFocusedTitleColor
	}
}

// OnBlur handles focus loss (Panel interface compatibility).
func (m *MigrationsContext) OnBlur() {
	m.focused = false
	if v := m.BaseContext.GetView(); v != nil {
		v.FrameColor = migPrimaryFrameColor
		v.TitleColor = migPrimaryTitleColor
	}
}

// ---------------------------------------------------------------------------
// Draw
// ---------------------------------------------------------------------------

// Draw renders the migrations panel (Panel interface compatibility).
func (m *MigrationsContext) Draw(dim boxlayout.Dimensions) error {
	v, err := m.g.SetView(m.GetViewName(), dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	// Store view references
	m.BaseContext.SetView(v)
	m.ScrollableTrait.SetView(v)

	v.Clear()
	v.Frame = true
	v.FrameRunes = migDefaultFrameRunes

	// Set tabs
	tabs := m.TabbedTrait.GetTabs()
	v.Tabs = tabs
	v.TabIndex = m.TabbedTrait.GetCurrentTabIdx()

	// Footer
	footer := m.buildFooter()
	v.Footer = footer
	v.Subtitle = ""

	// Frame and tab colours based on focus
	if m.focused {
		v.FrameColor = migFocusedFrameColor
		v.TitleColor = migFocusedTitleColor
		if len(tabs) == 1 {
			v.SelFgColor = migFocusedTitleColor
		} else {
			v.SelFgColor = migFocusedActiveTabColor
		}
	} else {
		v.FrameColor = migPrimaryFrameColor
		v.TitleColor = migPrimaryTitleColor
		if len(tabs) == 1 {
			v.SelFgColor = migPrimaryTitleColor
		} else {
			v.SelFgColor = migPrimaryActiveTabColor
		}
	}

	// Enable highlight for selection
	v.Highlight = true
	v.SelBgColor = migSelectionBgColor

	// Render items
	for _, item := range m.items {
		fmt.Fprintln(v, item)
	}

	// Adjust origin to ensure it's within valid bounds
	m.adjustOrigin(v)

	// Set cursor position to selected item
	v.SetCursor(0, m.selected-m.ScrollableTrait.GetOriginY())
	v.SetOrigin(0, m.ScrollableTrait.GetOriginY())

	return nil
}

// adjustOrigin clamps the scroll origin within valid bounds.
func (m *MigrationsContext) adjustOrigin(v *gocui.View) {
	if v == nil {
		return
	}

	contentLines := len(v.ViewBufferLines())
	_, viewHeight := v.Size()
	innerHeight := viewHeight - 2

	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	originY := m.ScrollableTrait.GetOriginY()
	if originY > maxOrigin {
		m.ScrollableTrait.SetOriginY(maxOrigin)
	}
}

// buildFooter builds the footer text (selection info in "n of n" format).
func (m *MigrationsContext) buildFooter() string {
	if len(m.items) == 0 || (len(m.items) == 1 && m.items[0] == m.tr.ErrorNoMigrationsFound) {
		return ""
	}
	return fmt.Sprintf(m.tr.MigrationsFooterFormat, m.selected+1, len(m.items))
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

// SelectNext moves the selection down by one.
func (m *MigrationsContext) SelectNext() {
	if len(m.items) == 0 {
		return
	}

	if m.selected < len(m.items)-1 {
		m.selected++

		// Auto-scroll if needed
		if v := m.BaseContext.GetView(); v != nil {
			_, h := v.Size()
			innerHeight := h - 2
			originY := m.ScrollableTrait.GetOriginY()
			if m.selected-originY >= innerHeight {
				m.ScrollableTrait.SetOriginY(originY + 1)
			}
		}

		m.notifySelectionChanged()
	}
}

// SelectPrev moves the selection up by one.
func (m *MigrationsContext) SelectPrev() {
	if len(m.items) == 0 {
		return
	}

	if m.selected > 0 {
		m.selected--

		// Auto-scroll if needed
		originY := m.ScrollableTrait.GetOriginY()
		if m.selected < originY {
			m.ScrollableTrait.SetOriginY(originY - 1)
		}

		m.notifySelectionChanged()
	}
}

// ---------------------------------------------------------------------------
// Scroll overrides (list-aware: also update selection)
// ---------------------------------------------------------------------------

// ScrollToTop scrolls to the top of the list and selects the first item.
func (m *MigrationsContext) ScrollToTop() {
	if len(m.items) == 0 {
		return
	}
	m.selected = 0
	m.ScrollableTrait.SetOriginY(0)
	m.notifySelectionChanged()
}

// ScrollToBottom scrolls to the bottom of the list and selects the last item.
func (m *MigrationsContext) ScrollToBottom() {
	if len(m.items) == 0 {
		return
	}

	maxIndex := len(m.items) - 1
	m.selected = maxIndex

	if v := m.BaseContext.GetView(); v != nil {
		_, h := v.Size()
		innerHeight := h - 2
		newOriginY := maxIndex - innerHeight + 1
		if newOriginY < 0 {
			newOriginY = 0
		}
		m.ScrollableTrait.SetOriginY(newOriginY)
	}

	m.notifySelectionChanged()
}

// ScrollUpByWheel scrolls the view up by 2 lines (mouse wheel).
func (m *MigrationsContext) ScrollUpByWheel() {
	m.ScrollableTrait.ScrollUpByWheel()
}

// ScrollDownByWheel scrolls the view down by 2 lines (mouse wheel).
func (m *MigrationsContext) ScrollDownByWheel() {
	if m.BaseContext.GetView() == nil || len(m.items) == 0 {
		return
	}

	contentLines := len(m.items)
	v := m.BaseContext.GetView()
	_, viewHeight := v.Size()
	innerHeight := viewHeight - 2

	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	originY := m.ScrollableTrait.GetOriginY()
	if originY < maxOrigin {
		newOriginY := originY + 2
		if newOriginY > maxOrigin {
			newOriginY = maxOrigin
		}
		m.ScrollableTrait.SetOriginY(newOriginY)
	}
}

// ---------------------------------------------------------------------------
// Tab switching (custom logic wrapping TabbedTrait)
// ---------------------------------------------------------------------------

// NextTab switches to the next tab, saving and restoring per-tab state.
func (m *MigrationsContext) NextTab() {
	tabs := m.TabbedTrait.GetTabs()
	if len(tabs) == 0 {
		return
	}

	m.saveCurrentTabState()
	m.TabbedTrait.NextTab()
	m.loadItemsForCurrentTab()
}

// PrevTab switches to the previous tab, saving and restoring per-tab state.
func (m *MigrationsContext) PrevTab() {
	tabs := m.TabbedTrait.GetTabs()
	if len(tabs) == 0 {
		return
	}

	m.saveCurrentTabState()
	m.TabbedTrait.PrevTab()
	m.loadItemsForCurrentTab()
}

// ---------------------------------------------------------------------------
// Mouse handlers
// ---------------------------------------------------------------------------

// HandleTabClick handles mouse click on a tab.
func (m *MigrationsContext) HandleTabClick(tabIndex int) error {
	if m.hasActiveModal != nil && m.hasActiveModal() {
		return nil
	}

	// Switch focus to this panel if not already focused
	if m.onPanelClick != nil {
		m.onPanelClick(m.GetViewName())
	}

	// Ignore if same tab or out of bounds
	if tabIndex == m.TabbedTrait.GetCurrentTabIdx() {
		return nil
	}
	tabs := m.TabbedTrait.GetTabs()
	if tabIndex < 0 || tabIndex >= len(tabs) {
		return nil
	}

	m.saveCurrentTabState()
	m.TabbedTrait.SetCurrentTabIdx(tabIndex)
	m.loadItemsForCurrentTab()

	return nil
}

// HandleListClick handles mouse click on a list item.
func (m *MigrationsContext) HandleListClick(y int) error {
	if m.hasActiveModal != nil && m.hasActiveModal() {
		return nil
	}

	if len(m.items) == 0 {
		return nil
	}

	clickedIndex := y
	if clickedIndex < 0 || clickedIndex >= len(m.items) {
		return nil
	}

	m.selected = clickedIndex
	m.notifySelectionChanged()

	// Switch focus to this panel if not already focused
	if m.onPanelClick != nil {
		m.onPanelClick(m.GetViewName())
	}

	return nil
}

// ---------------------------------------------------------------------------
// Refresh
// ---------------------------------------------------------------------------

// Refresh reloads all migration data, preserving current tab and selection
// where possible.
func (m *MigrationsContext) Refresh() {
	currentTabIdx := m.TabbedTrait.GetCurrentTabIdx()
	currentSelected := m.selected
	currentOriginY := m.ScrollableTrait.GetOriginY()

	// Save current tab state before refresh
	tabs := m.TabbedTrait.GetTabs()
	if currentTabIdx < len(tabs) {
		tabName := tabs[currentTabIdx]
		m.tabSelectedMap[tabName] = currentSelected
		m.tabOriginYMap[tabName] = currentOriginY
	}

	// Reload migrations
	m.loadMigrations()

	// Restore tab index if still valid
	newTabs := m.TabbedTrait.GetTabs()
	if currentTabIdx < len(newTabs) {
		m.TabbedTrait.SetCurrentTabIdx(currentTabIdx)
	} else {
		m.TabbedTrait.SetCurrentTabIdx(0)
	}

	// Reload items for current tab
	m.loadItemsForCurrentTab()

	// Restore selection if still valid
	if currentSelected < len(m.items) {
		m.selected = currentSelected
		m.ScrollableTrait.SetOriginY(currentOriginY)
		m.notifySelectionChanged()
	} else if len(m.items) > 0 {
		m.selected = len(m.items) - 1
		m.ScrollableTrait.SetOriginY(0)
		m.notifySelectionChanged()
	} else {
		m.selected = 0
		m.ScrollableTrait.SetOriginY(0)
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// notifySelectionChanged invokes the onSelectionChanged callback if set.
func (m *MigrationsContext) notifySelectionChanged() {
	if m.onSelectionChanged == nil {
		return
	}
	migration := m.GetSelectedMigration()
	tabName := m.GetCurrentTabName()
	m.onSelectionChanged(migration, tabName)
}

// migrationsForTab returns the migration slice for the given tab name.
func (m *MigrationsContext) migrationsForTab(tabName string) []prisma.Migration {
	switch tabName {
	case m.tr.TabLocal:
		return m.category.Local
	case m.tr.TabPending:
		return m.category.Pending
	case m.tr.TabDBOnly:
		return m.category.DBOnly
	}
	return nil
}

// saveCurrentTabState saves the current selection and scroll position for the
// active tab (keyed by tab name).
func (m *MigrationsContext) saveCurrentTabState() {
	tabName := m.TabbedTrait.GetCurrentTab()
	if tabName == "" {
		return
	}
	m.tabSelectedMap[tabName] = m.selected
	m.tabOriginYMap[tabName] = m.ScrollableTrait.GetOriginY()
}

// loadMigrations loads local and (optionally) DB migrations and sets up tabs.
func (m *MigrationsContext) loadMigrations() {
	cwd, err := os.Getwd()
	if err != nil {
		m.items = []string{m.tr.ErrorFailedGetWorkingDirectory}
		m.TabbedTrait.SetTabs([]string{m.tr.TabLocal})
		return
	}

	// Get local migrations
	localMigrations, err := prisma.GetLocalMigrations(cwd)
	if err != nil {
		m.items = []string{fmt.Sprintf(m.tr.ErrorLoadingLocalMigrations, err)}
		m.TabbedTrait.SetTabs([]string{m.tr.TabLocal})
		return
	}

	// Try to connect to database
	ds, err := prisma.GetDatasource(cwd)
	var dbMigrations []prisma.DBMigration
	m.dbConnected = false
	tableExists := false

	if err == nil && ds.URL != "" {
		client, err := database.NewClientFromDSN(ds.Provider, ds.URL)
		if err == nil {
			m.dbClient = client
			dbMigrations, err = prisma.GetDBMigrations(client.DB())
			if err == nil {
				m.dbConnected = true
				tableExists = true
			} else {
				// Check if error is due to missing table
				if isMigMissingTableError(err) {
					m.dbConnected = true
					tableExists = false
					dbMigrations = []prisma.DBMigration{}
				}
			}
		}
	}

	if m.dbConnected {
		m.category = prisma.CompareMigrations(localMigrations, dbMigrations)

		tabs := []string{m.tr.TabLocal}
		if len(m.category.Pending) > 0 {
			tabs = append(tabs, m.tr.TabPending)
		}
		if len(m.category.DBOnly) > 0 {
			tabs = append(tabs, m.tr.TabDBOnly)
		}
		m.TabbedTrait.SetTabs(tabs)

		m.tableExists = tableExists
	} else {
		m.category = prisma.MigrationCategory{
			Local:   localMigrations,
			Pending: []prisma.Migration{},
			DBOnly:  []prisma.Migration{},
		}
		m.TabbedTrait.SetTabs([]string{m.tr.TabLocal})
		m.tableExists = false
	}

	// Default to first tab
	m.TabbedTrait.SetCurrentTabIdx(0)
	m.loadItemsForCurrentTab()
}

// loadItemsForCurrentTab rebuilds the display items for the active tab and
// restores any previously saved selection / scroll position.
func (m *MigrationsContext) loadItemsForCurrentTab() {
	tabName := m.TabbedTrait.GetCurrentTab()
	if tabName == "" {
		m.items = []string{}
		return
	}

	migrations := m.migrationsForTab(tabName)

	if len(migrations) == 0 {
		m.items = []string{m.tr.ErrorNoMigrationsFound}
		return
	}

	m.items = make([]string, len(migrations))
	for i, mig := range migrations {
		// Parse migration name to show only description (without timestamp)
		displayName := mig.Name
		if len(mig.Name) > 15 && mig.Name[14] == '_' {
			displayName = mig.Name[15:] // Skip YYYYMMDDHHMMSS_ prefix
		}

		// Add index number with colour based on migration status
		var indexPrefix string
		if mig.IsEmpty {
			indexPrefix = style.Red(fmt.Sprintf("%4d │", i+1)) + " " // Red for empty
		} else if mig.HasDownSQL {
			indexPrefix = style.Green(fmt.Sprintf("%4d │", i+1)) + " " // Green for down.sql
		} else {
			indexPrefix = style.Stylize(fmt.Sprintf("%4d │", i+1), "90", false) + " " // Gray for normal
		}

		// Colour priority: Failed > Checksum Mismatch > Empty > Pending > Normal
		if mig.IsFailed {
			m.items[i] = indexPrefix + style.Cyan(displayName)
		} else if mig.ChecksumMismatch {
			m.items[i] = indexPrefix + style.Orange(displayName)
		} else if mig.IsEmpty {
			m.items[i] = indexPrefix + style.Red(displayName)
		} else if m.dbConnected && mig.AppliedAt == nil {
			m.items[i] = indexPrefix + style.Yellow(displayName)
		} else {
			m.items[i] = indexPrefix + displayName
		}
	}

	// Restore previous selection and scroll position for this tab
	if prevSelected, exists := m.tabSelectedMap[tabName]; exists {
		m.selected = prevSelected
		if m.selected >= len(m.items) {
			m.selected = len(m.items) - 1
		}
		if m.selected < 0 {
			m.selected = 0
		}
	} else {
		m.selected = 0
	}

	if prevOriginY, exists := m.tabOriginYMap[tabName]; exists {
		m.ScrollableTrait.SetOriginY(prevOriginY)
	} else {
		m.ScrollableTrait.SetOriginY(0)
	}

	// Notify selection changed
	m.notifySelectionChanged()
}

// isMigMissingTableError checks if an error is due to a missing
// _prisma_migrations table.
func isMigMissingTableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	missingTablePatterns := []string{
		"does not exist",               // PostgreSQL
		"doesn't exist",                // MySQL
		"no such table",                // SQLite
		"invalid object name",          // SQL Server
		"table or view does not exist", // Oracle
	}

	for _, pattern := range missingTablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}
