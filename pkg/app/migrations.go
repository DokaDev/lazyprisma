package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/database"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

type MigrationsPanel struct {
	BasePanel
	category    prisma.MigrationCategory // Categorized migrations
	items       []string                 // Current tab's migration names
	tabs        []string                 // Tab names (conditional)
	tabIndex    int                      // Current tab index
	selected    int                      // Selected item in current tab
	originY     int                      // Scroll position
	dbClient    *database.Client         // Database connection
	dbConnected bool                     // True if connected to database
	tableExists bool                     // True if _prisma_migrations table exists

	// Tab state preservation
	tabSelected map[string]int // Last selected index per tab
	tabOriginY  map[string]int // Last scroll position per tab

	// Details panel reference
	detailsPanel *DetailsPanel

	// App reference for modal check
	app *App
}

func NewMigrationsPanel(g *gocui.Gui) *MigrationsPanel {
	panel := &MigrationsPanel{
		BasePanel:   NewBasePanel(ViewMigrations, g),
		items:       []string{},
		tabs:        []string{},
		tabIndex:    0,
		selected:    0,
		tabSelected: make(map[string]int),
		tabOriginY:  make(map[string]int),
	}
	panel.loadMigrations()
	return panel
}

func (m *MigrationsPanel) loadMigrations() {
	cwd, err := os.Getwd()
	if err != nil {
		m.items = []string{"Error: Failed to get working directory"}
		m.tabs = []string{"Local"}
		return
	}

	// Get local migrations
	localMigrations, err := prisma.GetLocalMigrations(cwd)
	if err != nil {
		m.items = []string{fmt.Sprintf("Error loading local migrations: %v", err)}
		m.tabs = []string{"Local"}
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
				if isMissingTableError(err) {
					// Table doesn't exist - treat as empty DB (all migrations are pending)
					m.dbConnected = true
					tableExists = false
					dbMigrations = []prisma.DBMigration{} // Empty list
				}
				// Other errors: keep m.dbConnected = false
			}
		}
	}

	// If DB is connected (or table doesn't exist), categorize migrations
	if m.dbConnected {
		m.category = prisma.CompareMigrations(localMigrations, dbMigrations)

		// Build tabs based on available data
		m.tabs = []string{"Local"}
		if len(m.category.Pending) > 0 {
			m.tabs = append(m.tabs, "Pending")
		}
		if len(m.category.DBOnly) > 0 {
			m.tabs = append(m.tabs, "DB-Only")
		}

		// Store table existence info for display
		m.tableExists = tableExists
	} else {
		// DB connection failed completely
		m.category = prisma.MigrationCategory{
			Local:   localMigrations,
			Pending: []prisma.Migration{},
			DBOnly:  []prisma.Migration{},
		}
		m.tabs = []string{"Local"}
		m.tableExists = false
	}

	// Load items for current tab (default: Local)
	m.tabIndex = 0
	m.loadItemsForCurrentTab()
}

func (m *MigrationsPanel) loadItemsForCurrentTab() {
	if m.tabIndex >= len(m.tabs) {
		m.items = []string{}
		return
	}

	tabName := m.tabs[m.tabIndex]
	var migrations []prisma.Migration

	switch tabName {
	case "Local":
		migrations = m.category.Local
	case "Pending":
		migrations = m.category.Pending
	case "DB-Only":
		migrations = m.category.DBOnly
	}

	if len(migrations) == 0 {
		m.items = []string{"No migrations found"}
		return
	}

	m.items = make([]string, len(migrations))
	for i, mig := range migrations {
		// Parse migration name to show only description (without timestamp)
		displayName := mig.Name
		if len(mig.Name) > 15 && mig.Name[14] == '_' {
			displayName = mig.Name[15:] // Skip YYYYMMDDHHMMSS_ prefix
		}

		// Add index number with color based on migration status
		var indexPrefix string
		if mig.IsEmpty {
			indexPrefix = fmt.Sprintf("\033[31m%4d │\033[0m ", i+1) // Red for empty migrations (no migration.sql)
		} else if mig.HasDownSQL {
			indexPrefix = fmt.Sprintf("\033[32m%4d │\033[0m ", i+1) // Green for migrations with down.sql
		} else {
			indexPrefix = fmt.Sprintf("\033[90m%4d │\033[0m ", i+1) // Gray for normal migrations
		}

		// Color priority: Failed > Checksum Mismatch > Empty > Pending > Normal
		if mig.IsFailed {
			// In-Transaction migrations (finished_at IS NULL AND rolled_back_at IS NULL) are shown in cyan
			m.items[i] = indexPrefix + Cyan(displayName)
		} else if mig.ChecksumMismatch {
			// Checksum mismatch migrations are shown in orange
			m.items[i] = indexPrefix + Orange(displayName)
		} else if mig.IsEmpty {
			// Empty migrations (no migration.sql) are shown in red
			m.items[i] = indexPrefix + Red(displayName)
		} else if m.dbConnected && mig.AppliedAt == nil {
			// Pending migrations (not applied to DB) are shown in yellow
			// Only when DB is connected (otherwise we can't determine pending status)
			m.items[i] = indexPrefix + Yellow(displayName)
		} else {
			m.items[i] = indexPrefix + displayName
		}
	}

	// Restore previous selection and scroll position for this tab
	if prevSelected, exists := m.tabSelected[tabName]; exists {
		m.selected = prevSelected
		// Ensure selection is within bounds
		if m.selected >= len(m.items) {
			m.selected = len(m.items) - 1
		}
		if m.selected < 0 {
			m.selected = 0
		}
	} else {
		m.selected = 0
	}

	if prevOriginY, exists := m.tabOriginY[tabName]; exists {
		m.originY = prevOriginY
	} else {
		m.originY = 0
	}

	// Update details panel
	m.updateDetails()
}

func (m *MigrationsPanel) Draw(dim boxlayout.Dimensions) error {
	v, err := m.g.SetView(m.id, dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	// Setup view WITHOUT title (tabs replace title)
	m.v = v
	v.Clear()
	v.Frame = true
	v.FrameRunes = m.frameRunes

	// Set tabs
	v.Tabs = m.tabs
	v.TabIndex = m.tabIndex

	// Set footer based on current tab (moved from subtitle)
	if m.tabIndex < len(m.tabs) {
		// Footer (n of n) for all tabs
		footer := m.buildFooter()
		v.Footer = footer
	} else {
		v.Footer = ""
	}

	// No subtitle
	v.Subtitle = ""

	// Set frame and tab colors based on focus
	if m.focused {
		v.FrameColor = FocusedFrameColor
		v.TitleColor = FocusedTitleColor
		if len(m.tabs) == 1 {
			v.SelFgColor = FocusedTitleColor // Single tab: treat like title
		} else {
			v.SelFgColor = FocusedActiveTabColor // Multiple tabs: use active tab color
		}
	} else {
		v.FrameColor = PrimaryFrameColor
		v.TitleColor = PrimaryTitleColor
		if len(m.tabs) == 1 {
			v.SelFgColor = PrimaryTitleColor // Single tab: treat like title
		} else {
			v.SelFgColor = PrimaryActiveTabColor // Multiple tabs: use active tab color
		}
	}

	// Enable highlight for selection
	v.Highlight = true
	v.SelBgColor = SelectionBgColor

	// Render items
	for _, item := range m.items {
		fmt.Fprintln(v, item)
	}

	// Adjust origin to ensure it's within valid bounds
	AdjustOrigin(v, &m.originY)

	// Set cursor position to selected item
	v.SetCursor(0, m.selected-m.originY)
	v.SetOrigin(0, m.originY)

	return nil
}

// buildFooter builds the footer text (selection info in "n of n" format)
func (m *MigrationsPanel) buildFooter() string {
	// Don't show footer if no valid items
	if len(m.items) == 0 || (len(m.items) == 1 && m.items[0] == "No migrations found") {
		return ""
	}

	// Show selection info: "2 of 5"
	return fmt.Sprintf("%d of %d", m.selected+1, len(m.items))
}

func (m *MigrationsPanel) SelectNext() {
	if len(m.items) == 0 {
		return
	}

	if m.selected < len(m.items)-1 {
		m.selected++

		// Auto-scroll if needed
		if m.v != nil {
			_, h := m.v.Size()
			innerHeight := h - 2 // Subtract frame borders
			if m.selected-m.originY >= innerHeight {
				m.originY++
			}
		}

		// Update details panel
		m.updateDetails()
	}
}

func (m *MigrationsPanel) SelectPrev() {
	if len(m.items) == 0 {
		return
	}

	if m.selected > 0 {
		m.selected--

		// Auto-scroll if needed
		if m.selected < m.originY {
			m.originY--
		}

		// Update details panel
		m.updateDetails()
	}
}

// ScrollUpByWheel scrolls the migrations list up by 2 lines (mouse wheel)
func (m *MigrationsPanel) ScrollUpByWheel() {
	if m.originY > 0 {
		m.originY -= 2
		if m.originY < 0 {
			m.originY = 0
		}
	}
}

// ScrollDownByWheel scrolls the migrations list down by 2 lines (mouse wheel)
func (m *MigrationsPanel) ScrollDownByWheel() {
	if m.v == nil || len(m.items) == 0 {
		return
	}

	// Get actual content lines
	contentLines := len(m.items)
	_, viewHeight := m.v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	// Only scroll if we haven't reached the bottom
	if m.originY < maxOrigin {
		m.originY += 2
		if m.originY > maxOrigin {
			m.originY = maxOrigin
		}
	}
}

// ScrollToTop scrolls to the top of the migrations list
func (m *MigrationsPanel) ScrollToTop() {
	if len(m.items) == 0 {
		return
	}

	m.selected = 0
	m.originY = 0
	m.updateDetails()
}

// ScrollToBottom scrolls to the bottom of the migrations list
func (m *MigrationsPanel) ScrollToBottom() {
	if len(m.items) == 0 {
		return
	}

	maxIndex := len(m.items) - 1
	m.selected = maxIndex

	// Adjust origin to show the last item
	if m.v != nil {
		_, h := m.v.Size()
		innerHeight := h - 2 // Subtract frame borders
		m.originY = maxIndex - innerHeight + 1
		if m.originY < 0 {
			m.originY = 0
		}
	}

	m.updateDetails()
}

// NextTab switches to the next tab
func (m *MigrationsPanel) NextTab() {
	if len(m.tabs) == 0 {
		return
	}

	// Save current tab state before switching
	m.saveCurrentTabState()

	m.tabIndex = (m.tabIndex + 1) % len(m.tabs)
	m.loadItemsForCurrentTab()
}

// PrevTab switches to the previous tab
func (m *MigrationsPanel) PrevTab() {
	if len(m.tabs) == 0 {
		return
	}

	// Save current tab state before switching
	m.saveCurrentTabState()

	m.tabIndex = (m.tabIndex - 1 + len(m.tabs)) % len(m.tabs)
	m.loadItemsForCurrentTab()
}

// saveCurrentTabState saves the current selection and scroll position
func (m *MigrationsPanel) saveCurrentTabState() {
	if m.tabIndex >= len(m.tabs) {
		return
	}

	tabName := m.tabs[m.tabIndex]
	m.tabSelected[tabName] = m.selected
	m.tabOriginY[tabName] = m.originY
}

// GetSelectedMigration returns the currently selected migration
func (m *MigrationsPanel) GetSelectedMigration() *prisma.Migration {
	if m.tabIndex >= len(m.tabs) {
		return nil
	}

	tabName := m.tabs[m.tabIndex]
	var migrations []prisma.Migration

	switch tabName {
	case "Local":
		migrations = m.category.Local
	case "Pending":
		migrations = m.category.Pending
	case "DB-Only":
		migrations = m.category.DBOnly
	}

	if m.selected >= 0 && m.selected < len(migrations) {
		return &migrations[m.selected]
	}

	return nil
}

// GetCurrentTab returns the name of the current tab
func (m *MigrationsPanel) GetCurrentTab() string {
	if m.tabIndex >= len(m.tabs) {
		return ""
	}
	return m.tabs[m.tabIndex]
}

// SetDetailsPanel sets the details panel reference and performs initial update
func (m *MigrationsPanel) SetDetailsPanel(details *DetailsPanel) {
	m.detailsPanel = details
	// Set bidirectional reference
	details.migrationsPanel = m
	// Load Action-Needed data
	details.LoadActionNeededData()
	// Update details with initial selection (index 0)
	m.updateDetails()
}

// updateDetails updates the details panel with current selection
func (m *MigrationsPanel) updateDetails() {
	if m.detailsPanel == nil {
		return
	}

	migration := m.GetSelectedMigration()
	tabName := m.GetCurrentTab()
	m.detailsPanel.UpdateFromMigration(migration, tabName)
}

// isMissingTableError checks if error is due to missing _prisma_migrations table
func isMissingTableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Common error patterns across different databases
	missingTablePatterns := []string{
		"does not exist",           // PostgreSQL
		"doesn't exist",            // MySQL
		"no such table",            // SQLite
		"invalid object name",      // SQL Server
		"table or view does not exist", // Oracle
	}

	for _, pattern := range missingTablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// SetApp sets the app reference for modal checking
func (m *MigrationsPanel) SetApp(app *App) {
	m.app = app
}

// handleTabClick handles mouse click on tab bar
func (m *MigrationsPanel) handleTabClick(tabIndex int) error {
	// Ignore if modal is active
	if m.app != nil && m.app.HasActiveModal() {
		return nil
	}

	// First, switch focus to this panel if not already focused
	if m.app != nil {
		if err := m.app.handlePanelClick(ViewMigrations); err != nil {
			return err
		}
	}

	// Ignore if same tab is clicked
	if tabIndex == m.tabIndex {
		return nil
	}

	// Ignore if tab index is out of bounds
	if tabIndex < 0 || tabIndex >= len(m.tabs) {
		return nil
	}

	// Save current tab state
	m.saveCurrentTabState()

	// Switch to clicked tab
	m.tabIndex = tabIndex
	m.loadItemsForCurrentTab()

	return nil
}

// handleListClick handles mouse click on list item
func (m *MigrationsPanel) handleListClick(y int) error {
	// Ignore if modal is active
	if m.app != nil && m.app.HasActiveModal() {
		return nil
	}

	// Ignore if no items
	if len(m.items) == 0 {
		return nil
	}

	// opts.Y is already content-relative index (including origin)
	clickedIndex := y

	// Validate index
	if clickedIndex < 0 || clickedIndex >= len(m.items) {
		return nil
	}

	// Update selected index
	m.selected = clickedIndex

	// Update details panel
	m.updateDetails()

	// Switch focus to this panel if not already focused
	if m.app != nil {
		if err := m.app.handlePanelClick(ViewMigrations); err != nil {
			return err
		}
	}

	return nil
}

// Refresh reloads all migration data
func (m *MigrationsPanel) Refresh() {
	// Save current state to restore after refresh
	currentTabIndex := m.tabIndex
	currentSelected := m.selected
	currentOriginY := m.originY

	// Save current tab state before refresh (to prevent loadItemsForCurrentTab from resetting selection)
	if currentTabIndex < len(m.tabs) {
		currentTabName := m.tabs[currentTabIndex]
		m.tabSelected[currentTabName] = currentSelected
		m.tabOriginY[currentTabName] = currentOriginY
	}

	// Reload migrations
	m.loadMigrations()

	// Restore tab index if still valid
	if currentTabIndex < len(m.tabs) {
		m.tabIndex = currentTabIndex
	} else {
		// Reset to first tab if current tab no longer exists
		m.tabIndex = 0
	}

	// Reload items for current tab
	m.loadItemsForCurrentTab()

	// Restore selection if still valid
	if currentSelected < len(m.items) {
		m.selected = currentSelected
		m.originY = currentOriginY
		// Only update details if selection changed
		m.updateDetails()
	} else if len(m.items) > 0 {
		// If old selection is invalid, select last valid item
		m.selected = len(m.items) - 1
		m.originY = 0
		m.updateDetails()
	} else {
		// No items, reset
		m.selected = 0
		m.originY = 0
	}
}
