package tui

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/DokaDev/lazyprisma/pkg/env"
	"github.com/DokaDev/lazyprisma/pkg/prisma"

	"github.com/gdamore/tcell/v2"
)

type PanelBounds struct {
	x1, y1, x2, y2 int
}

type CommandLog struct {
	Command string
	Output  string
	Time    string
}

type App struct {
	screen               tcell.Screen
	status               prisma.Status
	executor             *prisma.Executor
	activePanelIdx       int
	selectedItemIdx      int // Index of selected item in active panel
	selectedDBOnlyIdx    int // Index of selected item in DB only panel
	lastSelectedPanel    int // Last selected panel (1: migrations, 2: db only)
	quit                 bool
	schemaDiff           string // Schema changes (diff result)
	showModal            bool   // Whether modal is displayed
	modalInput           string // Modal input content
	modalTitle           string // Modal title
	modalType            string // Modal type (input, confirm, error)
	modalCallback        func() // Callback to execute on confirmation
	panelBounds          [5]PanelBounds // info, migrations, dbonly, migration detail, output
	infoScroll           int            // info panel scroll offset
	migrationsScroll     int            // migrations panel scroll offset
	dbOnlyScroll         int            // db only panel scroll offset
	detailScroll         int            // migration detail panel scroll offset
	outputScroll         int            // output panel scroll offset
	isLoading            bool
	loadingMessage       string
	spinnerFrame         int
	dbConnected          bool
	commandLogs          []CommandLog    // Command execution logs
	pendingMigrations    map[string]bool // Pending migration name map (local only)
	missingMigrations    map[string]bool // Missing migration name map (DB only)
	missingMigrationList    []string        // Sorted list of missing migrations
	migrationDetailBounds   PanelBounds     // Migration Detail tab click area
	schemaDiffBounds        PanelBounds     // Schema Diff tab click area
	detailViewMode          string          // "migration" or "schema_diff"
	modalSelectedButton     int             // Modal button selection index (0: left, 1: right)
	pendingMigrationName    string          // Migration name to execute after reset
	helpScroll              int             // help modal scroll offset
}

func NewApp(status prisma.Status) (*App, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
	screen.EnableMouse()
	screen.Clear()

	app := &App{
		screen:               screen,
		status:               status,
		executor:             prisma.NewExecutor(),
		activePanelIdx:       0,
		selectedItemIdx:      0,
		selectedDBOnlyIdx:    0,
		lastSelectedPanel:    1, // Default: migrations
		quit:                 false,
		schemaDiff:           "",
		showModal:            false,
		modalInput:           "",
		modalTitle:           "",
		modalType:            "",
		modalCallback:        nil,
		infoScroll:           0,
		migrationsScroll:     0,
		dbOnlyScroll:         0,
		detailScroll:         0,
		outputScroll:         0,
		isLoading:            false,
		loadingMessage:       "",
		spinnerFrame:         0,
		dbConnected:          false,
		commandLogs:          []CommandLog{},
		pendingMigrations:    make(map[string]bool),
		missingMigrations:    make(map[string]bool),
		missingMigrationList: []string{},
		detailViewMode:       "migration",
		modalSelectedButton:  0,
		pendingMigrationName: "",
		helpScroll:           0,
	}

	// Check DB connection and migration status (async)
	go app.checkDBConnection()
	go app.initialMigrationStatus()
	go app.checkSchemaDiff()

	return app, nil
}

func (a *App) initialMigrationStatus() {
	// Run migrate status on initial load to identify pending migrations
	output, _ := a.executor.MigrateStatus()
	a.parsePendingMigrations(output)
	a.draw()
}

func (a *App) checkSchemaDiff() {
	output, err := a.executor.MigrateDiff()

	// Check for DB connection errors
	connectionErrors := []string{
		"Can't reach database server",
		"P1001",
		"Connection refused",
		"ECONNREFUSED",
		"database connection failed",
	}

	isConnectionError := false
	if err != nil {
		outputLower := strings.ToLower(output)
		for _, errMsg := range connectionErrors {
			if strings.Contains(outputLower, strings.ToLower(errMsg)) {
				isConnectionError = true
				break
			}
		}
	}

	// Using --exit-code option:
	// No diff: exit code 0 (err == nil), Diff exists: exit code != 0 (err != nil)
	// However, exclude DB connection errors
	if err != nil && output != "" && !isConnectionError {
		// Error with output = changes detected
		a.schemaDiff = output
	} else {
		a.schemaDiff = ""
	}
	a.draw()
}

func (a *App) Run() error {
	defer a.screen.Fini()

	// Initial rendering
	a.draw()

	// Set up timer for spinner animation
	go a.spinnerLoop()

	// Event loop
	for !a.quit {
		a.screen.Show()

		ev := a.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			a.screen.Sync()
			a.draw()
		case *tcell.EventKey:
			a.handleKey(ev)
		case *tcell.EventMouse:
			a.handleMouse(ev)
		}

		a.draw()
	}

	return nil
}

func (a *App) spinnerLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for !a.quit {
		<-ticker.C
		if a.isLoading {
			a.spinnerFrame = (a.spinnerFrame + 1) % 4
			a.draw()
		}
	}
}

func (a *App) handleKey(ev *tcell.EventKey) {
	// Handle modal input if modal is displayed
	if a.showModal {
		a.handleModalKey(ev)
		return
	}

	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		a.quit = true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			a.quit = true
		case 'r':
			if !a.isLoading {
				go a.runMigrateStatus()
			}
		case 'g':
			if !a.isLoading {
				go a.runGenerate()
			}
		case 'd':
			// migrate dev - show modal if schema changes exist
			if !a.isLoading {
				go a.runMigrateDev()
			}
		case 'D':
			// migrate deploy
			if !a.isLoading {
				go a.runMigrateDeploy()
			}
		case 'f':
			// format
			if !a.isLoading {
				go a.runFormat()
			}
		case 't':
			// studio
			if !a.isLoading {
				go a.runStudio()
			}
		case 'h':
			// help
			if !a.isLoading {
				go a.showHelp()
			}
		}
	case tcell.KeyLeft:
		a.activePanelIdx--
		if a.activePanelIdx < 0 {
			a.activePanelIdx = 4 // 5 panels (info, migrations, dbonly, detail, output)
		}
		// Skip DB only panel if it doesn't exist
		if a.activePanelIdx == 2 && len(a.missingMigrations) == 0 {
			a.activePanelIdx--
			if a.activePanelIdx < 0 {
				a.activePanelIdx = 4
			}
		}
	case tcell.KeyRight:
		a.activePanelIdx++
		if a.activePanelIdx > 4 {
			a.activePanelIdx = 0
		}
		// Skip DB only panel if it doesn't exist
		if a.activePanelIdx == 2 && len(a.missingMigrations) == 0 {
			a.activePanelIdx++
			if a.activePanelIdx > 4 {
				a.activePanelIdx = 0
			}
		}
	case tcell.KeyUp:
		a.handleUpKey()
	case tcell.KeyDown:
		a.handleDownKey()
	}
}

func (a *App) handleUpKey() {
	if a.activePanelIdx == 0 { // Info panel
		a.infoScroll--
		if a.infoScroll < 0 {
			a.infoScroll = 0
		}
	} else if a.activePanelIdx == 1 { // Migrations panel
		a.lastSelectedPanel = 1 // Selected from Migrations panel
		prevIdx := a.selectedItemIdx
		a.selectedItemIdx--
		if a.selectedItemIdx < 0 {
			a.selectedItemIdx = 0
		}

		// Reset detail scroll when selection changes
		if prevIdx != a.selectedItemIdx {
			a.detailScroll = 0
		}

		// Adjust scroll
		if a.selectedItemIdx < a.migrationsScroll {
			a.migrationsScroll = a.selectedItemIdx
		}
	} else if a.activePanelIdx == 2 { // DB Only panel
		a.lastSelectedPanel = 2 // Selected from DB Only panel
		prevIdx := a.selectedDBOnlyIdx
		a.selectedDBOnlyIdx--
		if a.selectedDBOnlyIdx < 0 {
			a.selectedDBOnlyIdx = 0
		}

		// Reset detail scroll when selection changes
		if prevIdx != a.selectedDBOnlyIdx {
			a.detailScroll = 0
		}

		// Adjust scroll
		if a.selectedDBOnlyIdx < a.dbOnlyScroll {
			a.dbOnlyScroll = a.selectedDBOnlyIdx
		}
	} else if a.activePanelIdx == 3 { // Migration Detail panel
		a.detailScroll--
		if a.detailScroll < 0 {
			a.detailScroll = 0
		}
	} else if a.activePanelIdx == 4 { // Output panel
		a.outputScroll--
		if a.outputScroll < 0 {
			a.outputScroll = 0
		}
	}
}

func (a *App) handleDownKey() {
	if a.activePanelIdx == 0 { // Info panel
		bounds := a.panelBounds[0]
		visibleLines := bounds.y2 - bounds.y1 - 3 // Exclude borders and padding
		totalLines := a.getInfoTotalLines()

		maxScroll := totalLines - visibleLines
		if maxScroll < 0 {
			maxScroll = 0
		}

		a.infoScroll++
		if a.infoScroll > maxScroll {
			a.infoScroll = maxScroll
		}
	} else if a.activePanelIdx == 1 { // Migrations panel
		a.lastSelectedPanel = 1 // Selected from Migrations panel
		maxIdx := a.getMaxItemIndex()
		if maxIdx < 0 {
			return
		}

		prevIdx := a.selectedItemIdx
		a.selectedItemIdx++
		if a.selectedItemIdx > maxIdx {
			a.selectedItemIdx = maxIdx
		}

		// Reset detail scroll when selection changes
		if prevIdx != a.selectedItemIdx {
			a.detailScroll = 0
		}

		// Adjust scroll
		bounds := a.panelBounds[1]
		visibleLines := bounds.y2 - bounds.y1 - 3 // Exclude borders and padding
		if a.selectedItemIdx >= a.migrationsScroll+visibleLines {
			a.migrationsScroll = a.selectedItemIdx - visibleLines + 1
		}
	} else if a.activePanelIdx == 2 { // DB Only panel
		a.lastSelectedPanel = 2 // Selected from DB Only panel
		maxIdx := len(a.missingMigrationList) - 1
		if maxIdx < 0 {
			return
		}

		prevIdx := a.selectedDBOnlyIdx
		a.selectedDBOnlyIdx++
		if a.selectedDBOnlyIdx > maxIdx {
			a.selectedDBOnlyIdx = maxIdx
		}

		// Reset detail scroll when selection changes
		if prevIdx != a.selectedDBOnlyIdx {
			a.detailScroll = 0
		}

		// Adjust scroll
		bounds := a.panelBounds[2]
		visibleLines := bounds.y2 - bounds.y1 - 3 // Exclude borders and padding
		if a.selectedDBOnlyIdx >= a.dbOnlyScroll+visibleLines {
			a.dbOnlyScroll = a.selectedDBOnlyIdx - visibleLines + 1
		}
	} else if a.activePanelIdx == 3 { // Migration Detail panel
		bounds := a.panelBounds[3]
		visibleLines := bounds.y2 - bounds.y1 - 3 // Exclude borders and padding
		totalLines := a.getDetailTotalLines()

		maxScroll := totalLines - visibleLines
		if maxScroll < 0 {
			maxScroll = 0
		}

		a.detailScroll++
		if a.detailScroll > maxScroll {
			a.detailScroll = maxScroll
		}
	} else if a.activePanelIdx == 4 { // Output panel
		bounds := a.panelBounds[4]
		visibleLines := bounds.y2 - bounds.y1 - 3 // Exclude borders and padding
		totalLines := a.getOutputTotalLines()

		maxScroll := totalLines - visibleLines
		if maxScroll < 0 {
			maxScroll = 0
		}

		a.outputScroll++
		if a.outputScroll > maxScroll {
			a.outputScroll = maxScroll
		}
	}
}

func (a *App) getMaxItemIndex() int {
	switch a.activePanelIdx {
	case 1: // Migrations panel
		totalItems := len(a.status.Migrations)
		return totalItems - 1
	case 2: // DB Only panel
		return len(a.missingMigrationList) - 1
	default:
		return -1 // Panel does not support selection
	}
}

func (a *App) getInfoTotalLines() int {
	totalLines := 4 // Node.js + npm + Prisma + blank line

	if a.status.SchemaExists {
		totalLines += 3 // Provider + Client + Database
		if a.status.DatabaseURL != "Not configured" {
			// URL is split into 2 lines by last / (host, DB name)
			url := a.status.DatabaseURL
			lastSlashIdx := strings.LastIndex(url, "/")
			protocolEnd := strings.Index(url, "://")
			if protocolEnd >= 0 && lastSlashIdx > protocolEnd+3 && lastSlashIdx < len(url)-1 {
				totalLines += 2
			} else {
				totalLines += 1
			}
		}
	}

	return totalLines
}

func (a *App) getDetailTotalLines() int {
	content := a.getSelectedMigrationSQL()
	if content == "" {
		return 1 // "No migration selected" message
	}

	// Calculate number of header lines
	headerLines := a.getSelectedMigrationHeader()
	headerCount := len(headerLines)
	if headerCount > 0 {
		headerCount += 1 // Add separator line
	}

	lines := strings.Split(content, "\n")
	return headerCount + len(lines)
}

func (a *App) getSelectedMigrationHeader() []string {
	// Return header based on last selected panel
	if a.lastSelectedPanel == 1 {
		// Selected from Migrations panel
		if a.selectedItemIdx >= 0 && a.selectedItemIdx < len(a.status.Migrations) {
			migration := a.status.Migrations[a.selectedItemIdx]
			folderName := migration.Timestamp + "_" + migration.Name

			// Format timestamp: 20250909064316 -> 2025-09-09 06:43:16
			timestamp := migration.Timestamp
			if len(timestamp) == 14 {
				timestamp = timestamp[0:4] + "-" + timestamp[4:6] + "-" + timestamp[6:8] + " " +
					timestamp[8:10] + ":" + timestamp[10:12] + ":" + timestamp[12:14]
			}

			return []string{
				"Timestamp: " + timestamp,
				"Path:      prisma/migrations/" + folderName + "/migration.sql",
			}
		}
	} else if a.lastSelectedPanel == 2 {
		// Selected from DB Only panel
		if len(a.missingMigrationList) > 0 && a.selectedDBOnlyIdx >= 0 && a.selectedDBOnlyIdx < len(a.missingMigrationList) {
			migrationName := a.missingMigrationList[a.selectedDBOnlyIdx]

			// Extract and format timestamp
			parts := strings.SplitN(migrationName, "_", 2)
			timestamp := parts[0]
			if len(timestamp) == 14 {
				timestamp = timestamp[0:4] + "-" + timestamp[4:6] + "-" + timestamp[6:8] + " " +
					timestamp[8:10] + ":" + timestamp[10:12] + ":" + timestamp[12:14]
			}

			return []string{
				"Timestamp: " + timestamp,
				"Status:    DB only (local file not found)",
			}
		}
	}

	return []string{}
}

func (a *App) getSelectedMigrationSQL() string {
	// Return content based on last selected panel
	if a.lastSelectedPanel == 1 {
		// Selected from Migrations panel
		if a.selectedItemIdx >= 0 && a.selectedItemIdx < len(a.status.Migrations) {
			migration := a.status.Migrations[a.selectedItemIdx]
			// Construct full folder name from migration.Timestamp and migration.Name
			folderName := migration.Timestamp + "_" + migration.Name
			sqlPath := "prisma/migrations/" + folderName + "/migration.sql"

			content, err := os.ReadFile(sqlPath)
			if err != nil {
				return "Error reading migration file: " + err.Error()
			}

			return string(content)
		}
	} else if a.lastSelectedPanel == 2 {
		// Selected from DB Only panel
		if len(a.missingMigrationList) > 0 && a.selectedDBOnlyIdx >= 0 && a.selectedDBOnlyIdx < len(a.missingMigrationList) {
			return "This migration exists only in the database.\nThe local migration file is not found in prisma/migrations."
		}
	}

	return ""
}

func (a *App) getOutputTotalLines() int {
	if len(a.commandLogs) == 0 {
		return 1 // "No commands executed yet" message
	}

	totalLines := 0
	for _, log := range a.commandLogs {
		totalLines += 1 // timestamp + command
		// output lines (split by newline)
		outputLines := strings.Split(log.Output, "\n")
		totalLines += len(outputLines)
		totalLines += 1 // blank line (separator)
	}

	return totalLines
}

func (a *App) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()

	// Handle mouse wheel
	if ev.Buttons()&tcell.WheelUp != 0 {
		// Wheel up in Info panel
		infoBounds := a.panelBounds[0]
		if x >= infoBounds.x1 && x <= infoBounds.x2 && y >= infoBounds.y1 && y <= infoBounds.y2 {
			a.infoScroll--
			if a.infoScroll < 0 {
				a.infoScroll = 0
			}
			return
		}

		// Wheel up in Migrations panel
		migBounds := a.panelBounds[1]
		if x >= migBounds.x1 && x <= migBounds.x2 && y >= migBounds.y1 && y <= migBounds.y2 {
			a.migrationsScroll--
			if a.migrationsScroll < 0 {
				a.migrationsScroll = 0
			}
			return
		}

		// Wheel up in DB Only panel
		if len(a.missingMigrationList) > 0 {
			dbOnlyBounds := a.panelBounds[2]
			if x >= dbOnlyBounds.x1 && x <= dbOnlyBounds.x2 && y >= dbOnlyBounds.y1 && y <= dbOnlyBounds.y2 {
				a.dbOnlyScroll--
				if a.dbOnlyScroll < 0 {
					a.dbOnlyScroll = 0
				}
				return
			}
		}

		// Wheel up in Migration Detail panel
		detailBounds := a.panelBounds[3]
		if x >= detailBounds.x1 && x <= detailBounds.x2 && y >= detailBounds.y1 && y <= detailBounds.y2 {
			a.detailScroll--
			if a.detailScroll < 0 {
				a.detailScroll = 0
			}
			return
		}

		// Wheel up in Output panel
		outBounds := a.panelBounds[4]
		if x >= outBounds.x1 && x <= outBounds.x2 && y >= outBounds.y1 && y <= outBounds.y2 {
			a.outputScroll--
			if a.outputScroll < 0 {
				a.outputScroll = 0
			}
			return
		}
	}
	if ev.Buttons()&tcell.WheelDown != 0 {
		// Wheel down in Info panel
		infoBounds := a.panelBounds[0]
		if x >= infoBounds.x1 && x <= infoBounds.x2 && y >= infoBounds.y1 && y <= infoBounds.y2 {
			totalLines := a.getInfoTotalLines()
			visibleLines := infoBounds.y2 - infoBounds.y1 - 3
			maxScroll := totalLines - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}

			a.infoScroll++
			if a.infoScroll > maxScroll {
				a.infoScroll = maxScroll
			}
			return
		}

		// Wheel down in Migrations panel
		migBounds := a.panelBounds[1]
		if x >= migBounds.x1 && x <= migBounds.x2 && y >= migBounds.y1 && y <= migBounds.y2 {
			totalItems := len(a.status.Migrations)
			visibleLines := migBounds.y2 - migBounds.y1 - 3
			maxScroll := totalItems - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}

			a.migrationsScroll++
			if a.migrationsScroll > maxScroll {
				a.migrationsScroll = maxScroll
			}
			return
		}

		// Wheel down in DB Only panel
		if len(a.missingMigrationList) > 0 {
			dbOnlyBounds := a.panelBounds[2]
			if x >= dbOnlyBounds.x1 && x <= dbOnlyBounds.x2 && y >= dbOnlyBounds.y1 && y <= dbOnlyBounds.y2 {
				totalItems := len(a.missingMigrationList)
				visibleLines := dbOnlyBounds.y2 - dbOnlyBounds.y1 - 3
				maxScroll := totalItems - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}

				a.dbOnlyScroll++
				if a.dbOnlyScroll > maxScroll {
					a.dbOnlyScroll = maxScroll
				}
				return
			}
		}

		// Wheel down in Migration Detail panel
		detailBounds := a.panelBounds[3]
		if x >= detailBounds.x1 && x <= detailBounds.x2 && y >= detailBounds.y1 && y <= detailBounds.y2 {
			totalLines := a.getDetailTotalLines()
			visibleLines := detailBounds.y2 - detailBounds.y1 - 3
			maxScroll := totalLines - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}

			a.detailScroll++
			if a.detailScroll > maxScroll {
				a.detailScroll = maxScroll
			}
			return
		}

		// Wheel down in Output panel
		outBounds := a.panelBounds[4]
		if x >= outBounds.x1 && x <= outBounds.x2 && y >= outBounds.y1 && y <= outBounds.y2 {
			totalLines := a.getOutputTotalLines()
			visibleLines := outBounds.y2 - outBounds.y1 - 3
			maxScroll := totalLines - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}

			a.outputScroll++
			if a.outputScroll > maxScroll {
				a.outputScroll = maxScroll
			}
			return
		}
	}

	// Handle left click
	if ev.Buttons()&tcell.Button1 == 0 {
		return
	}

	// Check Migration Detail tab click
	if x >= a.migrationDetailBounds.x1 && x <= a.migrationDetailBounds.x2 &&
		y >= a.migrationDetailBounds.y1 && y <= a.migrationDetailBounds.y2 {
		a.detailViewMode = "migration"
		a.detailScroll = 0
		return
	}

	// Check Schema Diff tab click (only when schema changed)
	if a.schemaDiff != "" {
		if x >= a.schemaDiffBounds.x1 && x <= a.schemaDiffBounds.x2 &&
			y >= a.schemaDiffBounds.y1 && y <= a.schemaDiffBounds.y2 {
			a.detailViewMode = "schema_diff"
			a.detailScroll = 0
			return
		}
	}

	// Check panel bounds to find clicked panel
	for i, bounds := range a.panelBounds {
		if x >= bounds.x1 && x <= bounds.x2 && y >= bounds.y1 && y <= bounds.y2 {
			a.activePanelIdx = i

			// Handle item click within Migrations panel
			if i == 1 {
				totalItems := len(a.status.Migrations)
				if totalItems > 0 {
					itemY := y - bounds.y1 - 2 // Exclude borders and padding
					if itemY >= 0 {
						clickedIdx := a.migrationsScroll + itemY
						if clickedIdx < totalItems {
							a.selectedItemIdx = clickedIdx
							a.lastSelectedPanel = 1
							a.detailScroll = 0 // Reset detail scroll when selection changes
						}
					}
				}
			}

			// Handle item click within DB Only panel
			if i == 2 && len(a.missingMigrationList) > 0 {
				totalItems := len(a.missingMigrationList)
				if totalItems > 0 {
					itemY := y - bounds.y1 - 2 // Exclude borders and padding
					if itemY >= 0 {
						clickedIdx := a.dbOnlyScroll + itemY
						if clickedIdx < totalItems {
							a.selectedDBOnlyIdx = clickedIdx
							a.lastSelectedPanel = 2
							a.detailScroll = 0 // Reset detail scroll when selection changes
						}
					}
				}
			}
			break
		}
	}
}

func (a *App) drawSpinner(x, y int) {
	spinnerChars := []rune{'|', '/', '-', '\\'}
	char := spinnerChars[a.spinnerFrame]

	style := tcell.StyleDefault.Foreground(tcell.ColorYellow)

	// Spinner character
	a.screen.SetContent(x, y, char, nil, style)

	// Loading message
	if a.loadingMessage != "" {
		msg := " " + a.loadingMessage + "..."
		for i, r := range msg {
			a.screen.SetContent(x+1+i, y, r, nil, style)
		}
	}
}

func (a *App) draw() {
	a.screen.Clear()

	width, height := a.screen.Size()
	leftWidth := width / 3

	// Calculate Info panel height (dynamically adjusted based on terminal height and focus)
	baseInfoHeight := 13 // Node.js(1) + npm(1) + Prisma(1) + blank line(1) + Provider(1) + Client(1) + Database(1) + URL(2) + padding(3)
	if !a.status.SchemaExists {
		baseInfoHeight = 9 // Smaller when schema doesn't exist
	}

	infoHeight := baseInfoHeight
	minInfoHeight := 4 // Minimum height (title and borders only)

	// Adjust Info panel height when terminal is short
	if height < 30 {
		// Minimize Info to give priority to Migration panel when it has focus
		if a.activePanelIdx == 1 {
			infoHeight = minInfoHeight
		} else {
			// Reduce proportionally to terminal height
			reduction := (30 - height) / 2
			infoHeight = baseInfoHeight - reduction
			if infoHeight < minInfoHeight {
				infoHeight = minInfoHeight
			}
		}
	}

	minHeight := infoHeight + 8 // Info + minimum Migrations space
	showBothPanels := height >= minHeight

	// Draw panels and save boundaries
	if showBothPanels || a.activePanelIdx == 0 {
		// Info panel
		a.panelBounds[0] = PanelBounds{0, 0, leftWidth, infoHeight}
		a.drawInfoPanel(0, 0, leftWidth, infoHeight)
	}

	// Migrations and DB Only panel layout
	migrationsStart := 0
	if showBothPanels {
		migrationsStart = infoHeight + 1
	}

	hasDBOnly := len(a.missingMigrationList) > 0
	dbOnlyHeight := 8 // DB Only panel fixed height

	if hasDBOnly {
		// Split space with Migrations when DB Only exists
		migrationsEnd := height - 3 - dbOnlyHeight - 1
		if showBothPanels || a.activePanelIdx == 1 {
			a.panelBounds[1] = PanelBounds{0, migrationsStart, leftWidth, migrationsEnd}
			a.drawMigrationsPanel(0, migrationsStart, leftWidth, migrationsEnd)
		}

		// DB Only panel
		dbOnlyStart := migrationsEnd + 1
		dbOnlyEnd := height - 3
		if showBothPanels || a.activePanelIdx == 2 {
			a.panelBounds[2] = PanelBounds{0, dbOnlyStart, leftWidth, dbOnlyEnd}
			a.drawDBOnlyPanel(0, dbOnlyStart, leftWidth, dbOnlyEnd)
		}
	} else {
		// Migrations uses all space when DB Only doesn't exist
		if showBothPanels || a.activePanelIdx == 1 {
			a.panelBounds[1] = PanelBounds{0, migrationsStart, leftWidth, height - 3}
			a.drawMigrationsPanel(0, migrationsStart, leftWidth, height-3)
		}
	}

	// Calculate right panel layout
	rightX1 := leftWidth + 1
	rightX2 := width - 1

	// Dynamically calculate Migration Detail panel height (high priority)
	minDetailHeight := 15
	maxDetailHeight := height - 3 - 8 // Reserve minimum Output space
	detailHeight := (height - 3) * 2 / 3 // About 2/3 of terminal height
	if detailHeight < minDetailHeight {
		detailHeight = minDetailHeight
	}
	if detailHeight > maxDetailHeight {
		detailHeight = maxDetailHeight
	}

	// Make larger when Detail panel has focus
	if a.activePanelIdx == 3 && height > 40 {
		detailHeight = (height - 3) * 3 / 4 // Use 75%
		if detailHeight > maxDetailHeight {
			detailHeight = maxDetailHeight
		}
	}

	// Migration Detail panel (top)
	detailY1 := 0
	detailY2 := detailHeight
	a.panelBounds[3] = PanelBounds{rightX1, detailY1, rightX2, detailY2}
	a.drawMigrationDetailPanel(rightX1, detailY1, rightX2, detailY2)

	// Output panel (bottom) - use remaining space
	outputY1 := detailY2 + 1
	outputY2 := height - 3
	a.panelBounds[4] = PanelBounds{rightX1, outputY1, rightX2, outputY2}
	a.drawOutputPanel(rightX1, outputY1, rightX2, outputY2)

	// Help (move to right when spinner is present)
	helpX := 0
	if a.isLoading {
		msgLen := len(a.loadingMessage) + 6 // "| xxx..."
		helpX = msgLen
	}
	a.drawHelp(helpX, height-2, width-1, height-1)

	// Spinner (bottom left)
	if a.isLoading {
		a.drawSpinner(0, height-2)
	}

	// Version info (bottom right)
	a.drawVersionInfo(width, height)

	// Display modal
	if a.showModal {
		a.drawModal()
	}

	a.screen.Show()
}

func (a *App) checkDBConnection() {
	// Check DB connection with migrate status
	output, err := a.executor.MigrateStatus()

	// Check if actual connection failure message exists
	connectionErrors := []string{
		"Can't reach database server",
		"P1001",
		"Connection refused",
		"ECONNREFUSED",
		"database connection failed",
	}

	isConnectionError := false
	outputLower := strings.ToLower(output)
	for _, errMsg := range connectionErrors {
		if strings.Contains(outputLower, strings.ToLower(errMsg)) {
			isConnectionError = true
			break
		}
	}

	// Consider connected if no error or if error doesn't contain connection failure message
	// (DB is connected even if there are pending migrations)
	if err == nil || !isConnectionError {
		a.dbConnected = true
	} else {
		a.dbConnected = false
	}

	a.draw()
}

func (a *App) addCommandLog(command, output string) {
	log := CommandLog{
		Command: command,
		Output:  output,
		Time:    time.Now().Format("15:04:05"),
	}
	a.commandLogs = append(a.commandLogs, log)

	// Auto-scroll to bottom
	totalLines := a.getOutputTotalLines()
	if len(a.panelBounds) > 3 {
		bounds := a.panelBounds[3]
		visibleLines := bounds.y2 - bounds.y1 - 3
		maxScroll := totalLines - visibleLines
		if maxScroll < 0 {
			maxScroll = 0
		}
		a.outputScroll = maxScroll
	}
}

func (a *App) stripANSI(str string) string {
	// Remove ANSI escape codes: \x1b[...m format
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(str, "")
}

func (a *App) runMigrateStatus() {
	a.isLoading = true
	a.loadingMessage = "refreshing"

	// 1. Reload entire status (schema, migrations, env, etc.)
	a.status = prisma.GetStatus()

	// 2. Execute migrate status
	output, err := a.executor.MigrateStatus()

	// 3. Parse pending migrations (parse output regardless of error)
	a.parsePendingMigrations(output)

	// 4. Check DB connection status
	a.checkDBConnectionSync(output)

	// 5. Check schema diff
	go a.checkSchemaDiff()

	// 6. Add log (append error to output if error exists)
	logOutput := output
	if err != nil {
		logOutput = output + "\n\nError: " + err.Error()
	}
	a.addCommandLog("npx prisma migrate status", logOutput)

	a.isLoading = false
	a.loadingMessage = ""
	a.draw() // Refresh screen immediately
}

func (a *App) checkDBConnectionSync(output string) {
	// Check if actual connection failure message exists
	connectionErrors := []string{
		"Can't reach database server",
		"P1001",
		"Connection refused",
		"ECONNREFUSED",
		"database connection failed",
	}

	isConnectionError := false
	outputLower := strings.ToLower(output)
	for _, errMsg := range connectionErrors {
		if strings.Contains(outputLower, strings.ToLower(errMsg)) {
			isConnectionError = true
			break
		}
	}

	// Consider connected if no connection failure message
	a.dbConnected = !isConnectionError
}

func (a *App) parsePendingMigrations(output string) {
	// Initialize maps
	a.pendingMigrations = make(map[string]bool)
	a.missingMigrations = make(map[string]bool)
	a.missingMigrationList = []string{}

	lines := strings.Split(output, "\n")
	inPendingSection := false
	inMissingSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// "The migration have not yet been applied:" section start (local only)
		if (strings.Contains(trimmed, "migration have not yet been applied") ||
			strings.Contains(trimmed, "migrations have not yet been applied")) &&
			!strings.Contains(trimmed, "database") {
			inPendingSection = true
			inMissingSection = false
			continue
		}

		// "The migration from the database are not found locally" section start (DB only)
		if strings.Contains(trimmed, "from the database are not found locally") ||
			strings.Contains(trimmed, "from your database are not found") {
			inMissingSection = true
			inPendingSection = false
			continue
		}

		// End with blank line or other section
		if (inPendingSection || inMissingSection) &&
			(trimmed == "" ||
				strings.HasPrefix(trimmed, "To apply") ||
				strings.HasPrefix(trimmed, "The last common")) {
			if trimmed != "" && !strings.HasPrefix(trimmed, "To apply") {
				// "The last common" etc. don't end the section
			} else {
				inPendingSection = false
				inMissingSection = false
			}
			if trimmed == "" || strings.HasPrefix(trimmed, "To apply") {
				continue
			}
		}

		// Extract migration name
		if inPendingSection && trimmed != "" &&
			!strings.HasPrefix(trimmed, "The") &&
			!strings.HasPrefix(trimmed, "Your") {
			a.pendingMigrations[trimmed] = true
		}

		if inMissingSection && trimmed != "" &&
			!strings.HasPrefix(trimmed, "The") &&
			!strings.HasPrefix(trimmed, "Your") {
			a.missingMigrations[trimmed] = true
			a.missingMigrationList = append(a.missingMigrationList, trimmed)
		}
	}
}

func (a *App) runGenerate() {
	a.isLoading = true
	a.loadingMessage = "generate"

	output, err := a.executor.Generate()
	if err != nil {
		output = "Error: " + err.Error()
	}

	a.addCommandLog("npx prisma generate", output)

	a.isLoading = false
	a.loadingMessage = ""
	a.draw() // Refresh screen immediately
}

func (a *App) runMigrateDev() {
	// 1. Check schema changes
	if a.schemaDiff == "" {
		a.showModal = true
		a.modalTitle = "No Schema Changes"
		a.modalType = "error"
		a.modalInput = "No schema changes detected."
		a.draw()
		return
	}

	// 2. Check schema validation
	a.isLoading = true
	a.loadingMessage = "validating schema"
	a.draw()

	output, err := a.executor.Validate()

	a.isLoading = false
	a.loadingMessage = ""

	if err != nil {
		// Validation failed
		a.showModal = true
		a.modalTitle = "Schema Validation Failed"
		a.modalType = "error"
		a.modalInput = a.stripANSI(output)
		a.draw()
		return
	}

	// 3. Show DB URL confirmation modal
	a.showModal = true
	a.modalTitle = "Confirm Migration Dev"
	a.modalType = "confirm_dev"
	a.modalInput = ""
	a.draw()
}

func (a *App) executeMigrateDev(name string) {
	a.isLoading = true
	a.loadingMessage = "migrate dev"
	a.draw()

	output, err := a.executor.MigrateDev(name)

	a.isLoading = false
	a.loadingMessage = ""

	// Check for reset keywords in output
	outputLower := strings.ToLower(output)
	resetKeywords := []string{
		"reset",
		"all data will be lost",
		"do you want to continue",
	}

	isResetRequired := false
	for _, keyword := range resetKeywords {
		if strings.Contains(outputLower, keyword) {
			isResetRequired = true
			break
		}
	}

	if isResetRequired {
		// Reset required - show confirmation modal
		a.showModal = true
		a.modalTitle = "Database Reset Required"
		a.modalType = "confirm_reset"
		a.modalInput = "Prisma wants to reset the database.\nALL DATA WILL BE LOST.\n\nReasons:\n- Migration drift detected\n- Modified migrations\n- Missing migrations\n\nCheck Output panel for details."
		a.modalSelectedButton = 1 // Default: Cancel
		a.pendingMigrationName = name

		// Record full log to Output
		if err != nil {
			output = output + "\n\nError: " + err.Error()
		}
		a.addCommandLog("npx prisma migrate dev --name "+name, output)
		a.draw()
		return
	}

	// Normal execution
	if err != nil {
		output = output + "\n\nError: " + err.Error()
	}

	a.addCommandLog("npx prisma migrate dev --name "+name, output)

	// Reload status
	a.status = prisma.GetStatus()
	go a.checkSchemaDiff()

	a.draw()
}

func (a *App) executeReset() {
	a.isLoading = true
	a.loadingMessage = "resetting database"
	a.draw()

	// 1. Execute reset
	output, err := a.executor.MigrateReset()
	if err != nil {
		output = output + "\n\nError: " + err.Error()
	}

	a.addCommandLog("npx prisma migrate reset --force", output)

	// 2. Execute original migration dev on successful reset
	if err == nil && a.pendingMigrationName != "" {
		a.loadingMessage = "migrate dev"
		a.draw()

		devOutput, devErr := a.executor.MigrateDev(a.pendingMigrationName)
		if devErr != nil {
			devOutput = devOutput + "\n\nError: " + devErr.Error()
		}

		a.addCommandLog("npx prisma migrate dev --name "+a.pendingMigrationName, devOutput)
	}

	// Reload status
	a.status = prisma.GetStatus()
	go a.checkSchemaDiff()

	a.isLoading = false
	a.loadingMessage = ""
	a.pendingMigrationName = ""
	a.draw()
}

func (a *App) runMigrateDeploy() {
	// Show DB URL confirmation modal
	a.showModal = true
	a.modalTitle = "Confirm Migration Deploy"
	a.modalType = "confirm_deploy"
	a.modalInput = ""
	a.draw()
}

func (a *App) executeMigrateDeploy() {
	a.isLoading = true
	a.loadingMessage = "migrate deploy"

	output, err := a.executor.MigrateDeploy()
	if err != nil {
		output = output + "\n\nError: " + err.Error()
	}

	a.addCommandLog("npx prisma migrate deploy", output)

	// Reload status
	a.status = prisma.GetStatus()
	go a.checkSchemaDiff()

	a.isLoading = false
	a.loadingMessage = ""
	a.draw()
}

func (a *App) runFormat() {
	a.isLoading = true
	a.loadingMessage = "format"

	output, err := a.executor.Format()
	if err != nil {
		output = output + "\n\nError: " + err.Error()
	}

	a.addCommandLog("npx prisma format", output)

	// Re-read schema
	a.status = prisma.GetStatus()

	a.isLoading = false
	a.loadingMessage = ""
	a.draw()
}

func (a *App) runStudio() {
	// Studio runs in background, so just add log without spinner
	output, err := a.executor.Studio()
	if err != nil {
		output = output + "\n\nError: " + err.Error()
	}

	a.addCommandLog("npx prisma studio", output)
	a.draw()
}

func (a *App) showHelp() {
	a.isLoading = true
	a.loadingMessage = "loading help"
	a.draw()

	output, err := a.executor.Help()

	a.isLoading = false
	a.loadingMessage = ""

	if err != nil {
		output = output + "\n\nError: " + err.Error()
	}

	// Remove ANSI codes
	output = a.stripANSI(output)

	// Show help modal
	a.showModal = true
	a.modalTitle = "Prisma CLI Help"
	a.modalType = "help"
	a.modalInput = output
	a.helpScroll = 0
	a.draw()
}

func (a *App) handleModalKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		// Close modal
		a.showModal = false
		a.modalInput = ""
		a.modalType = ""
		a.modalCallback = nil
		a.modalSelectedButton = 0
		a.pendingMigrationName = ""
		a.helpScroll = 0
		a.draw()
	case tcell.KeyUp:
		// Scroll in help modal
		if a.modalType == "help" {
			a.helpScroll--
			if a.helpScroll < 0 {
				a.helpScroll = 0
			}
			a.draw()
		}
	case tcell.KeyDown:
		// Scroll in help modal
		if a.modalType == "help" {
			a.helpScroll++
			a.draw()
		}
	case tcell.KeyLeft:
		// Button selection in confirm_reset modal
		if a.modalType == "confirm_reset" {
			a.modalSelectedButton--
			if a.modalSelectedButton < 0 {
				a.modalSelectedButton = 0
			}
			a.draw()
		}
	case tcell.KeyRight:
		// Button selection in confirm_reset modal
		if a.modalType == "confirm_reset" {
			a.modalSelectedButton++
			if a.modalSelectedButton > 1 {
				a.modalSelectedButton = 1
			}
			a.draw()
		}
	case tcell.KeyEnter:
		// Just close if error display mode
		if a.modalType == "error" {
			a.showModal = false
			a.modalInput = ""
			a.modalType = ""
			a.draw()
			return
		}
		// Just close if help modal
		if a.modalType == "help" {
			a.showModal = false
			a.modalType = ""
			a.helpScroll = 0
			a.draw()
			return
		}
		// confirm_reset: Reset/Cancel selection
		if a.modalType == "confirm_reset" {
			if a.modalSelectedButton == 0 {
				// Reset selected
				a.showModal = false
				a.modalType = ""
				a.modalSelectedButton = 0
				go a.executeReset()
			} else {
				// Cancel selected
				a.showModal = false
				a.modalType = ""
				a.modalSelectedButton = 0
				a.pendingMigrationName = ""
			}
			a.draw()
			return
		}
		// confirm_dev: Switch to input modal after confirmation
		if a.modalType == "confirm_dev" {
			a.showModal = true
			a.modalTitle = "Migration Name"
			a.modalType = "input"
			a.modalInput = ""
			a.draw()
			return
		}
		// confirm_deploy: Execute immediately after confirmation
		if a.modalType == "confirm_deploy" {
			a.showModal = false
			a.modalType = ""
			go a.executeMigrateDeploy()
			return
		}
		// input: Input complete
		if a.modalType == "input" && a.modalInput != "" {
			a.showModal = false
			a.modalType = ""
			go a.executeMigrateDev(a.modalInput)
			a.modalInput = ""
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Allow backspace only in input mode
		if a.modalType == "input" && len(a.modalInput) > 0 {
			a.modalInput = a.modalInput[:len(a.modalInput)-1]
			a.draw()
		}
	case tcell.KeyRune:
		// Allow input only in input mode
		if a.modalType == "input" {
			r := ev.Rune()
			if r == ' ' {
				r = '_'
			}
			a.modalInput += string(r)
			a.draw()
		}
	}
}

func (a *App) drawModal() {
	width, height := a.screen.Size()

	// Modal size and position
	modalWidth := 60
	modalHeight := 7

	// Adjust size for error, confirm_reset, help modals
	if a.modalType == "error" || a.modalType == "confirm_reset" {
		lines := strings.Split(a.modalInput, "\n")
		contentHeight := len(lines) + 6 // Include padding
		if a.modalType == "confirm_reset" {
			contentHeight += 2 // Button space
		}
		if contentHeight > modalHeight {
			modalHeight = contentHeight
		}
		if modalHeight > height-6 {
			modalHeight = height - 6
		}
	} else if a.modalType == "help" {
		// Help modal uses most of terminal size
		modalWidth = width - 10
		modalHeight = height - 6
		if modalWidth < 60 {
			modalWidth = 60
		}
		if modalHeight < 20 {
			modalHeight = 20
		}
	}

	if modalWidth > width-4 {
		modalWidth = width - 4
	}

	x1 := (width - modalWidth) / 2
	y1 := (height - modalHeight) / 2
	x2 := x1 + modalWidth
	y2 := y1 + modalHeight

	// Background (gray background instead of translucent effect)
	bgStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			a.screen.SetContent(x, y, ' ', nil, bgStyle)
		}
	}

	// Border
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
	// Top
	a.screen.SetContent(x1, y1, '╭', nil, boxStyle)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y1, '─', nil, boxStyle)
	}
	a.screen.SetContent(x2, y1, '╮', nil, boxStyle)

	// Left and right
	for y := y1 + 1; y < y2; y++ {
		a.screen.SetContent(x1, y, '│', nil, boxStyle)
		a.screen.SetContent(x2, y, '│', nil, boxStyle)
	}

	// Bottom
	a.screen.SetContent(x1, y2, '╰', nil, boxStyle)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y2, '─', nil, boxStyle)
	}
	a.screen.SetContent(x2, y2, '╯', nil, boxStyle)

	// Title
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack).Bold(true)
	titleText := " " + a.modalTitle + " "
	titleX := x1 + (modalWidth-len(titleText))/2
	for i, r := range titleText {
		a.screen.SetContent(titleX+i, y1, r, nil, titleStyle)
	}

	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	if a.modalType == "error" {
		// Display error message (multi-line support)
		lines := strings.Split(a.modalInput, "\n")
		msgY := y1 + 2

		for _, line := range lines {
			if msgY >= y2-2 {
				break // Stop if out of space
			}

			// Truncate long lines
			displayLine := line
			maxLineWidth := modalWidth - 4
			if len(displayLine) > maxLineWidth {
				displayLine = displayLine[:maxLineWidth]
			}

			// Display line (left-aligned)
			lineX := x1 + 2
			for _, r := range displayLine {
				if lineX >= x2-1 {
					break
				}
				a.screen.SetContent(lineX, msgY, r, nil, textStyle)
				lineX++
			}
			msgY++
		}

		// Help message
		hintText := "Press ESC or Enter to close"
		hintY := y2 - 1
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "confirm_reset" {
		// Display reset confirmation message
		lines := strings.Split(a.modalInput, "\n")
		msgY := y1 + 2

		for _, line := range lines {
			if msgY >= y2-4 {
				break
			}

			displayLine := line
			maxLineWidth := modalWidth - 4
			if len(displayLine) > maxLineWidth {
				displayLine = displayLine[:maxLineWidth]
			}

			lineX := x1 + 2
			for _, r := range displayLine {
				if lineX >= x2-1 {
					break
				}
				a.screen.SetContent(lineX, msgY, r, nil, textStyle)
				lineX++
			}
			msgY++
		}

		// Draw buttons
		buttonY := y2 - 2
		buttonSpacing := 4
		resetBtn := " Reset "
		cancelBtn := " Cancel "

		totalWidth := len(resetBtn) + buttonSpacing + len(cancelBtn)
		startX := x1 + (modalWidth-totalWidth)/2

		// Reset button
		resetStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		if a.modalSelectedButton == 0 {
			resetStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorRed).Bold(true)
		}
		for i, r := range resetBtn {
			a.screen.SetContent(startX+i, buttonY, r, nil, resetStyle)
		}

		// Cancel button
		cancelStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		if a.modalSelectedButton == 1 {
			cancelStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen).Bold(true)
		}
		cancelX := startX + len(resetBtn) + buttonSpacing
		for i, r := range cancelBtn {
			a.screen.SetContent(cancelX+i, buttonY, r, nil, cancelStyle)
		}

		// Help message
		hintText := "←/→: Select | Enter: Confirm | ESC: Cancel"
		hintY := y2 - 1
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "help" {
		// Display help content
		lines := strings.Split(a.modalInput, "\n")

		// Apply scroll
		visibleStart := a.helpScroll
		visibleEnd := len(lines)
		contentY := y1 + 2
		maxContentY := y2 - 2

		for i := visibleStart; i < visibleEnd && contentY < maxContentY; i++ {
			line := lines[i]

			// Truncate long lines
			maxLineWidth := modalWidth - 4
			if len(line) > maxLineWidth {
				line = line[:maxLineWidth]
			}

			// Draw line
			lineX := x1 + 2
			for _, r := range line {
				if lineX >= x2-1 {
					break
				}
				a.screen.SetContent(lineX, contentY, r, nil, textStyle)
				lineX++
			}
			contentY++
		}

		// Display scrollbar
		totalLines := len(lines)
		visibleLines := maxContentY - (y1 + 2)
		if totalLines > visibleLines {
			scrollbarX := x2 - 1
			scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)

			scrollbarHeight := (visibleLines * visibleLines) / totalLines
			if scrollbarHeight < 1 {
				scrollbarHeight = 1
			}

			scrollbarPos := (a.helpScroll * visibleLines) / totalLines
			scrollbarY := y1 + 2 + scrollbarPos

			for i := 0; i < scrollbarHeight; i++ {
				if scrollbarY+i < maxContentY {
					a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
				}
			}
		}

		// Help message
		hintText := "↑/↓: Scroll | ESC or Enter: Close"
		hintY := y2 - 1
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "confirm_dev" || a.modalType == "confirm_deploy" {
		// Display confirmation message
		envVarName := a.status.SchemaInfo.DatasourceEnvVar
		envReader := env.NewDotEnvReader(envVarName)
		maskedURL := envReader.MaskDatabaseURL(a.status.DatabaseURL)

		var msg1, msg2 string
		if a.modalType == "confirm_dev" {
			msg1 = "This will create a new migration and apply it to:"
			msg2 = "Press Enter to continue, ESC to cancel"
		} else {
			msg1 = "This will apply pending migrations to:"
			msg2 = "Press Enter to continue, ESC to cancel"
		}

		msgY := y1 + 2
		msgX := x1 + (modalWidth-len(msg1))/2
		for i, r := range msg1 {
			a.screen.SetContent(msgX+i, msgY, r, nil, textStyle)
		}

		// Display DB URL
		urlStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua).Background(tcell.ColorBlack)
		urlY := y1 + 3
		urlX := x1 + 2
		displayURL := maskedURL
		if len(displayURL) > modalWidth-4 {
			displayURL = displayURL[:modalWidth-4]
		}
		for i, r := range displayURL {
			a.screen.SetContent(urlX+i, urlY, r, nil, urlStyle)
		}

		// Help message
		hintY := y1 + 5
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(msg2))/2
		for i, r := range msg2 {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "input" {
		// Input field label
		labelText := "Enter migration name:"
		labelY := y1 + 2
		for i, r := range labelText {
			a.screen.SetContent(x1+2+i, labelY, r, nil, textStyle)
		}

		// Input field
		inputY := y1 + 3
		inputStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
		inputFieldWidth := modalWidth - 4
		for x := 0; x < inputFieldWidth; x++ {
			a.screen.SetContent(x1+2+x, inputY, ' ', nil, inputStyle)
		}

		// Display input content
		for i, r := range a.modalInput {
			if i < inputFieldWidth {
				a.screen.SetContent(x1+2+i, inputY, r, nil, inputStyle)
			}
		}

		// Display cursor
		cursorX := x1 + 2 + len(a.modalInput)
		if cursorX < x2-1 {
			a.screen.SetContent(cursorX, inputY, '▏', nil, inputStyle.Foreground(tcell.ColorYellow))
		}

		// Help message
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintText := "Press Enter to confirm, Esc to cancel"
		hintY := y1 + 5
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	}
}
