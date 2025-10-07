package tui

import (
	"fmt"
	"strings"

	"lazyprisma/pkg/version"

	"github.com/gdamore/tcell/v2"
)

// drawBox draws a box with optional title
func (a *App) drawBox(x1, y1, x2, y2 int, title string, active bool) {
	if x2 <= x1 || y2 <= y1 {
		return
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	if active {
		style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	}

	// Top border (rounded corners)
	a.screen.SetContent(x1, y1, '╭', nil, style)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y1, '─', nil, style)
	}
	a.screen.SetContent(x2, y1, '╮', nil, style)

	// Left and right borders
	for y := y1 + 1; y < y2; y++ {
		a.screen.SetContent(x1, y, '│', nil, style)
		a.screen.SetContent(x2, y, '│', nil, style)
	}

	// Bottom border (rounded corners)
	a.screen.SetContent(x1, y2, '╰', nil, style)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y2, '─', nil, style)
	}
	a.screen.SetContent(x2, y2, '╯', nil, style)

	// Title
	if title != "" {
		titleStr := " " + title + " "
		titleStyle := style
		if active {
			titleStyle = style.Bold(true)
		}
		a.drawText(x1+2, y1, x1+2+len(titleStr), y1, titleStr, titleStyle)
	}
}

// drawText draws text at the given position
func (a *App) drawText(x1, y, x2, maxY int, text string, style tcell.Style) {
	x := x1
	for _, r := range text {
		if x >= x2 {
			break
		}
		a.screen.SetContent(x, y, r, nil, style)
		x++
	}
}

func (a *App) drawInfoPanel(x1, y1, x2, y2 int) {
	active := a.activePanelIdx == 0
	a.drawBox(x1, y1, x2, y2, "Info", active)

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	// Structure to store line information
	type Line struct {
		text  string
		style tcell.Style
		extra []struct {
			x     int
			text  string
			style tcell.Style
		}
	}

	// Construct all lines first
	lines := []Line{}

	// Node.js
	lines = append(lines, Line{text: fmt.Sprintf("Node.js: %s", a.status.NodeVersion), style: style})

	// npm
	lines = append(lines, Line{text: fmt.Sprintf("npm:     %s", a.status.NPMVersion), style: style})

	// Prisma
	prismaStatus := "✓"
	prismaColor := tcell.ColorGreen
	if !a.status.CLIAvailable {
		prismaStatus = "✗"
		prismaColor = tcell.ColorRed
	}
	prismaStyle := tcell.StyleDefault.Foreground(prismaColor)
	prismaLine := Line{text: "Prisma:  ", style: style}
	prismaLine.extra = append(prismaLine.extra, struct {
		x     int
		text  string
		style tcell.Style
	}{x: 9, text: prismaStatus + " " + a.status.Version, style: prismaStyle})
	lines = append(lines, prismaLine)

	// Empty line
	lines = append(lines, Line{text: "", style: style})

	// Add only when Schema exists
	if a.status.SchemaExists {
		// Provider
		if a.status.SchemaInfo.DatasourceProvider != "" {
			lines = append(lines, Line{text: fmt.Sprintf("Provider: %s", a.status.SchemaInfo.DatasourceProvider), style: style})
		}

		// Client
		var clientPath string
		var clientColor tcell.Color
		if a.status.SchemaInfo.GeneratorOutputIsSet {
			// Explicitly set - yellow
			clientPath = a.status.SchemaInfo.GeneratorOutput
			clientColor = tcell.ColorYellow
		} else {
			// Using default value - green
			clientPath = "node_modules/.prisma/client (Default)"
			clientColor = tcell.ColorGreen
		}

		clientLine := Line{text: "Client:   ", style: style}
		clientLine.extra = append(clientLine.extra, struct {
			x     int
			text  string
			style tcell.Style
		}{x: 10, text: clientPath, style: tcell.StyleDefault.Foreground(clientColor)})
		lines = append(lines, clientLine)

		// Database Status
		if a.status.DatabaseURL == "Not configured" {
			lines = append(lines, Line{text: "Database: Not configured", style: style})
		} else {
			connStatus := "✓ Connected"
			connColor := tcell.ColorGreen
			if !a.dbConnected {
				connStatus = "✗ Disconnected"
				connColor = tcell.ColorRed
			}
			connStyle := tcell.StyleDefault.Foreground(connColor)
			dbLine := Line{text: "Database: ", style: style}
			dbLine.extra = append(dbLine.extra, struct {
				x     int
				text  string
				style tcell.Style
			}{x: 10, text: connStatus, style: connStyle})

			// Display if there are schema changes
			if a.schemaDiff != "" {
				warningStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
				dbLine.extra = append(dbLine.extra, struct {
					x     int
					text  string
					style tcell.Style
				}{x: 10 + len(connStatus) + 1, text: "(Schema changed)", style: warningStyle})
			}

			lines = append(lines, dbLine)

			// Database URL (line break at last / before DB name)
			urlStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua)
			url := a.status.DatabaseURL

			// Find last / position (after port number)
			// Format: postgresql://user:pass@host:port/dbname?params
			lastSlashIdx := strings.LastIndex(url, "/")
			if lastSlashIdx > 0 && lastSlashIdx < len(url)-1 {
				// Check if it's after the protocol's // (exclude cases like postgresql://)
				protocolEnd := strings.Index(url, "://")
				if protocolEnd >= 0 && lastSlashIdx > protocolEnd+3 {
					// Host:port part
					hostPart := url[:lastSlashIdx]
					// DB name part
					dbPart := url[lastSlashIdx:]

					lines = append(lines, Line{text: "URL:      " + hostPart, style: urlStyle})
					lines = append(lines, Line{text: "          " + dbPart, style: urlStyle})
				} else {
					lines = append(lines, Line{text: "URL:      " + url, style: urlStyle})
				}
			} else {
				lines = append(lines, Line{text: "URL:      " + url, style: urlStyle})
			}
		}
	}

	// Draw lines with scroll applied
	visibleStart := a.infoScroll
	visibleEnd := len(lines)
	y := y1 + 2

	for i := visibleStart; i < visibleEnd; i++ {
		if y >= y2 {
			break
		}

		line := lines[i]
		a.drawText(x1+2, y, x2-1, y2-1, line.text, line.style)

		// Draw additional text if present
		for _, extra := range line.extra {
			a.drawText(x1+2+extra.x, y, x2-1, y2-1, extra.text, extra.style)
		}

		y++
	}

	// Draw scrollbar
	totalLines := len(lines)
	panelHeight := y2 - y1 - 2 // Actual height excluding borders
	if totalLines > panelHeight {
		scrollbarX := x2 - 1
		scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if active {
			scrollbarStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		}

		// Calculate scrollbar height
		scrollbarHeight := (panelHeight * panelHeight) / totalLines
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}

		// Calculate scrollbar position
		scrollbarPos := (a.infoScroll * panelHeight) / totalLines
		scrollbarY := y1 + 1 + scrollbarPos

		// Draw scrollbar
		for i := 0; i < scrollbarHeight; i++ {
			if scrollbarY+i < y2 {
				a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
			}
		}
	}
}

func (a *App) drawMigrationsPanel(x1, y1, x2, y2 int) {
	active := a.activePanelIdx == 1
	a.drawBox(x1, y1, x2, y2, "Migrations", active)

	if !a.status.SchemaExists {
		return
	}

	y := y1 + 2

	if len(a.status.Migrations) == 0 {
		style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		a.drawText(x1+2, y, x2-1, y2-1, "No migrations found", style)
		return
	}

	// Apply scroll (local migrations only)
	localMigrations := len(a.status.Migrations)
	totalItems := localMigrations

	visibleStart := a.migrationsScroll
	visibleEnd := totalItems

	for i := visibleStart; i < visibleEnd; i++ {
		if y >= y2-1 {
			break
		}

		var text string
		var textStyle tcell.Style
		var isPending bool

		migration := a.status.Migrations[i]
		text = fmt.Sprintf(" %d. %s", i+1, migration.Name)
		migrationKey := migration.Timestamp + "_" + migration.Name

		// Check pending
		if a.pendingMigrations[migrationKey] || a.pendingMigrations[migration.Name] {
			isPending = true
			textStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
		} else {
			textStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite)
		}

		// Display selected item (bold regardless of panel active state)
		if i == a.selectedItemIdx {
			// Selected item is always bold
			if isPending {
				textStyle = tcell.StyleDefault.
					Foreground(tcell.ColorYellow).
					Bold(true)
			} else {
				textStyle = tcell.StyleDefault.
					Foreground(tcell.ColorWhite).
					Bold(true)
			}

			// Add background color when it's the active panel
			if active {
				bgColor := tcell.ColorBlue
				if isPending {
					textStyle = tcell.StyleDefault.
						Background(bgColor).
						Foreground(tcell.ColorYellow).
						Bold(true)
				} else {
					textStyle = tcell.StyleDefault.
						Background(bgColor).
						Foreground(tcell.ColorWhite).
						Bold(true)
				}
				// Fill entire line
				for x := x1 + 1; x < x2; x++ {
					a.screen.SetContent(x, y, ' ', nil, textStyle)
				}
			}
		}

		a.drawText(x1+1, y, x2-1, y2-1, text, textStyle)

		// Display icon
		indicatorX := x2 - 3
		if isPending {
			iconStyle := textStyle
			if !active || i != a.selectedItemIdx {
				iconStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
			}
			a.screen.SetContent(indicatorX, y, '●', nil, iconStyle)
		}

		y++
	}

	// Draw scrollbar
	totalCount := totalItems
	panelHeight := y2 - y1 - 3 // Actual height excluding borders
	if totalCount > panelHeight {
		scrollbarX := x2 - 1
		scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if active {
			scrollbarStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		}

		// Calculate scrollbar height
		scrollbarHeight := (panelHeight * panelHeight) / totalCount
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}

		// Calculate scrollbar position
		scrollbarPos := (a.migrationsScroll * panelHeight) / totalCount
		scrollbarY := y1 + 1 + scrollbarPos

		// Draw scrollbar
		for i := 0; i < scrollbarHeight; i++ {
			if scrollbarY+i < y2 {
				a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
			}
		}
	}

	// Display "n of n" at bottom right
	currentIdx := a.selectedItemIdx + 1 // 1-based index
	counterText := fmt.Sprintf("%d of %d", currentIdx, totalCount)
	counterStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	if active {
		counterStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	}
	counterX := x2 - len(counterText) - 1
	a.drawText(counterX, y2, x2-1, y2, counterText, counterStyle)
}

func (a *App) drawDBOnlyPanel(x1, y1, x2, y2 int) {
	active := a.activePanelIdx == 2
	a.drawBox(x1, y1, x2, y2, "DB Only Migrations", active)

	if len(a.missingMigrationList) == 0 {
		return
	}

	y := y1 + 2

	// Apply scroll
	totalItems := len(a.missingMigrationList)
	visibleStart := a.dbOnlyScroll
	visibleEnd := totalItems

	for i := visibleStart; i < visibleEnd; i++ {
		if y >= y2-1 {
			break
		}

		migrationName := a.missingMigrationList[i]
		text := fmt.Sprintf(" %d. %s", i+1, migrationName)
		textStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

		// Display selected item
		if i == a.selectedDBOnlyIdx {
			textStyle = tcell.StyleDefault.
				Foreground(tcell.ColorRed).
				Bold(true)

			// Add background color when panel is active
			if active {
				textStyle = tcell.StyleDefault.
					Background(tcell.ColorBlue).
					Foreground(tcell.ColorRed).
					Bold(true)
				// Fill entire line
				for x := x1 + 1; x < x2; x++ {
					a.screen.SetContent(x, y, ' ', nil, textStyle)
				}
			}
		}

		a.drawText(x1+1, y, x2-1, y2-1, text, textStyle)

		// Display icon
		indicatorX := x2 - 3
		iconStyle := textStyle
		if !active || i != a.selectedDBOnlyIdx {
			iconStyle = tcell.StyleDefault.Foreground(tcell.ColorRed)
		}
		a.screen.SetContent(indicatorX, y, '●', nil, iconStyle)

		y++
	}

	// Draw scrollbar
	panelHeight := y2 - y1 - 3
	if totalItems > panelHeight {
		scrollbarX := x2 - 1
		scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if active {
			scrollbarStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		}

		scrollbarHeight := (panelHeight * panelHeight) / totalItems
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}

		scrollbarPos := (a.dbOnlyScroll * panelHeight) / totalItems
		scrollbarY := y1 + 1 + scrollbarPos

		for i := 0; i < scrollbarHeight; i++ {
			if scrollbarY+i < y2 {
				a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
			}
		}
	}

	// Display "n of n" at bottom right
	currentIdx := a.selectedDBOnlyIdx + 1
	counterText := fmt.Sprintf("%d of %d", currentIdx, totalItems)
	counterStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	if active {
		counterStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	}
	counterX := x2 - len(counterText) - 1
	a.drawText(counterX, y2, x2-1, y2, counterText, counterStyle)
}

func (a *App) drawSchemaDiffContent(x1, y1, x2, y2, startY int) {
	active := a.activePanelIdx == 3
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	// Display Schema Diff content
	lines := strings.Split(a.schemaDiff, "\n")

	// Draw lines with scroll applied
	visibleStart := a.detailScroll
	visibleEnd := len(lines)
	panelWidth := x2 - x1 - 3 // Excluding left and right margins
	y := startY

	for i := visibleStart; i < visibleEnd; i++ {
		if y > y2-1 {
			break
		}

		line := lines[i]

		// Truncate long text
		if len(line) > panelWidth {
			line = line[:panelWidth]
		}

		// SQL keyword highlighting
		sqlStyle := style
		upperLine := strings.ToUpper(line)
		keywords := []string{"CREATE", "TABLE", "ALTER", "DROP", "SELECT", "INSERT", "UPDATE", "DELETE", "FROM", "WHERE", "PRIMARY", "KEY", "FOREIGN", "REFERENCES", "ADD", "COLUMN", "INDEX"}
		for _, keyword := range keywords {
			if strings.Contains(upperLine, keyword) {
				sqlStyle = tcell.StyleDefault.Foreground(tcell.ColorAqua)
				break
			}
		}

		a.drawText(x1+2, y, x2-1, y2-1, line, sqlStyle)
		y++
	}

	// Draw scrollbar
	totalLines := len(lines)
	panelHeight := y2 - y1 - 2 // Actual height excluding borders
	if totalLines > panelHeight {
		scrollbarX := x2 - 1
		scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if active {
			scrollbarStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		}

		// Calculate scrollbar height
		scrollbarHeight := (panelHeight * panelHeight) / totalLines
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}

		// Calculate scrollbar position
		scrollbarPos := (a.detailScroll * panelHeight) / totalLines
		scrollbarY := y1 + 1 + scrollbarPos

		// Draw scrollbar
		for i := 0; i < scrollbarHeight; i++ {
			if scrollbarY+i < y2 {
				a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
			}
		}
	}
}

func (a *App) drawMigrationDetailPanel(x1, y1, x2, y2 int) {
	active := a.activePanelIdx == 3

	// Draw basic box without title
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	if active {
		style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	}

	// Top border
	a.screen.SetContent(x1, y1, '╭', nil, style)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y1, '─', nil, style)
	}
	a.screen.SetContent(x2, y1, '╮', nil, style)

	// Left and right borders
	for y := y1 + 1; y < y2; y++ {
		a.screen.SetContent(x1, y, '│', nil, style)
		a.screen.SetContent(x2, y, '│', nil, style)
	}

	// Bottom border
	a.screen.SetContent(x1, y2, '╰', nil, style)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y2, '─', nil, style)
	}
	a.screen.SetContent(x2, y2, '╯', nil, style)

	// Draw tab menu
	tabY := y1
	currentX := x1 + 2

	// Tab 1: Migration Detail
	tab1Text := " Migration Detail "
	tab1Style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	if a.detailViewMode == "migration" {
		tab1Style = tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	}
	for i, r := range tab1Text {
		a.screen.SetContent(currentX+i, tabY, r, nil, tab1Style)
	}
	a.migrationDetailBounds = PanelBounds{
		x1: currentX,
		y1: tabY,
		x2: currentX + len(tab1Text) - 1,
		y2: tabY,
	}
	currentX += len(tab1Text)

	// Tab 2: Schema Diff (only when schema changed)
	if a.schemaDiff != "" {
		tab2Text := " Schema Diff "
		tab2Style := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		if a.detailViewMode == "schema_diff" {
			tab2Style = tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
		}
		for i, r := range tab2Text {
			a.screen.SetContent(currentX+i, tabY, r, nil, tab2Style)
		}
		a.schemaDiffBounds = PanelBounds{
			x1: currentX,
			y1: tabY,
			x2: currentX + len(tab2Text) - 1,
			y2: tabY,
		}
	}

	contentStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	y := y1 + 2

	// Display diff content when in Schema Diff mode
	if a.detailViewMode == "schema_diff" && a.schemaDiff != "" {
		a.drawSchemaDiffContent(x1, y1, x2, y2, y)
		return
	}

	// Display header information (Timestamp, Path, etc.)
	headerLines := a.getSelectedMigrationHeader()
	if len(headerLines) > 0 {
		headerStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		for _, headerLine := range headerLines {
			a.drawText(x1+2, y, x2-1, y2-1, headerLine, headerStyle)
			y++
		}
		// Separator line between header and SQL
		separatorStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		separator := strings.Repeat("─", (x2-x1-3))
		a.drawText(x1+2, y, x2-1, y2-1, separator, separatorStyle)
		y++
	}

	// Get SQL content of selected migration
	sqlContent := a.getSelectedMigrationSQL()

	if sqlContent == "" {
		a.drawText(x1+2, y, x2-1, y2-1, "No migration selected.", contentStyle)
		y += 2
		a.drawText(x1+2, y, x2-1, y2-1, "Select a migration from the list to view its SQL.", contentStyle)
		return
	}

	// Display SQL content line by line
	lines := strings.Split(sqlContent, "\n")

	// Draw lines with scroll applied
	visibleStart := a.detailScroll
	visibleEnd := len(lines)
	panelWidth := x2 - x1 - 3 // Excluding left and right margins

	for i := visibleStart; i < visibleEnd; i++ {
		if y > y2-1 {
			break
		}

		line := lines[i]

		// Truncate long text
		if len(line) > panelWidth {
			line = line[:panelWidth]
		}

		// SQL keyword highlighting
		sqlStyle := contentStyle
		upperLine := strings.ToUpper(line)
		keywords := []string{"CREATE", "TABLE", "ALTER", "DROP", "SELECT", "INSERT", "UPDATE", "DELETE", "FROM", "WHERE", "PRIMARY", "KEY", "FOREIGN", "REFERENCES"}
		for _, keyword := range keywords {
			if strings.Contains(upperLine, keyword) {
				sqlStyle = tcell.StyleDefault.Foreground(tcell.ColorAqua)
				break
			}
		}

		a.drawText(x1+2, y, x2-1, y2-1, line, sqlStyle)
		y++
	}

	// Draw scrollbar
	totalLines := len(lines)
	panelHeight := y2 - y1 - 2 // Actual height excluding borders
	if totalLines > panelHeight {
		scrollbarX := x2 - 1
		scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if active {
			scrollbarStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		}

		// Calculate scrollbar height
		scrollbarHeight := (panelHeight * panelHeight) / totalLines
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}

		// Calculate scrollbar position
		scrollbarPos := (a.detailScroll * panelHeight) / totalLines
		scrollbarY := y1 + 1 + scrollbarPos

		// Draw scrollbar
		for i := 0; i < scrollbarHeight; i++ {
			if scrollbarY+i < y2 {
				a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
			}
		}
	}
}

func (a *App) drawOutputPanel(x1, y1, x2, y2 int) {
	active := a.activePanelIdx == 4
	a.drawBox(x1, y1, x2, y2, "Output", active)

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	// Display message if no command logs exist
	if len(a.commandLogs) == 0 {
		y := y1 + 2
		a.drawText(x1+2, y, x2-1, y2-1, "No commands executed yet.", style)
		y += 2
		a.drawText(x1+2, y, x2-1, y2-1, "Press 'r' to run migrate status or 'g' to generate.", style)
		return
	}

	// Construct all log lines
	type Line struct {
		text  string
		style tcell.Style
	}

	lines := []Line{}
	for _, log := range a.commandLogs {
		// Timestamp + command
		cmdStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
		lines = append(lines, Line{
			text:  fmt.Sprintf("[%s] $ %s", log.Time, log.Command),
			style: cmdStyle,
		})

		// Output result
		outputLines := strings.Split(log.Output, "\n")
		for _, outLine := range outputLines {
			lines = append(lines, Line{
				text:  outLine,
				style: style,
			})
		}

		// Empty line (separator)
		lines = append(lines, Line{text: "", style: style})
	}

	// Draw lines with scroll applied
	visibleStart := a.outputScroll
	visibleEnd := len(lines)
	y := y1 + 2
	panelWidth := x2 - x1 - 3 // Excluding left and right margins

	for i := visibleStart; i < visibleEnd; i++ {
		if y >= y2 {
			break
		}

		line := lines[i]

		// Wrap long text
		if len(line.text) > panelWidth {
			wrapped := line.text[:panelWidth]
			a.drawText(x1+2, y, x2-1, y2-1, wrapped, line.style)
		} else {
			a.drawText(x1+2, y, x2-1, y2-1, line.text, line.style)
		}

		y++
	}

	// Draw scrollbar
	totalLines := len(lines)
	panelHeight := y2 - y1 - 2 // Actual height excluding borders
	if totalLines > panelHeight {
		scrollbarX := x2 - 1
		scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if active {
			scrollbarStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		}

		// Calculate scrollbar height
		scrollbarHeight := (panelHeight * panelHeight) / totalLines
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}

		// Calculate scrollbar position
		scrollbarPos := (a.outputScroll * panelHeight) / totalLines
		scrollbarY := y1 + 1 + scrollbarPos

		// Draw scrollbar
		for i := 0; i < scrollbarHeight; i++ {
			if scrollbarY+i < y2 {
				a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
			}
		}
	}
}

func (a *App) drawHelp(x1, y1, x2, y2 int) {
	// Display different guides based on active panel
	var helpText string
	switch a.activePanelIdx {
	case 0: // Info panel
		helpText = " ←/→: Switch | ↑/↓: Scroll | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 1: // Migrations panel
		helpText = " ←/→: Switch | ↑/↓: Select | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 2: // DB Only panel
		helpText = " ←/→: Switch | ↑/↓: Select | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 3: // Migration Detail panel
		helpText = " ←/→: Switch | ↑/↓: Scroll | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 4: // Output panel
		helpText = " ←/→: Switch | ↑/↓: Scroll | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	default:
		helpText = " ←/→: Switch | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	}

	// Display key commands in purple, descriptions in white
	x := x1
	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	keyStyle := tcell.StyleDefault.Foreground(tcell.ColorPurple)

	for _, ch := range helpText {
		style := normalStyle

		// Detect key parts (alphabet, arrow symbols, slash)
		if ch == '←' || ch == '→' || ch == '↑' || ch == '↓' ||
		   (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '/' {
			style = keyStyle
		}

		if x >= x2 {
			break
		}
		a.screen.SetContent(x, y1, ch, nil, style)
		x++
	}
}

// drawVersionInfo draws the DokaLab branding and version in the bottom right corner
func (a *App) drawVersionInfo(width, height int) {
	// Build version string: "DokaLab v0.1 Beta"
	author := version.Author
	ver := version.Version
	versionText := author + " v" + ver

	// Calculate starting position (bottom right corner, same line as help)
	x := width - len(versionText) - 1
	y := height - 2

	// Style: DokaLab in cyan, version in gray
	authorStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkCyan).Bold(true)
	versionStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	// Draw author name
	currentX := x
	for _, ch := range author {
		a.screen.SetContent(currentX, y, ch, nil, authorStyle)
		currentX++
	}

	// Draw " v" and version
	versionPrefix := " v" + ver
	for _, ch := range versionPrefix {
		a.screen.SetContent(currentX, y, ch, nil, versionStyle)
		currentX++
	}
}
