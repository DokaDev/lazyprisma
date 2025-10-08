package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

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

func (a *App) drawSchemaErrorContent(x1, y1, x2, y2, startY int) {
	active := a.activePanelIdx == 3
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)

	// Display Schema Error content
	lines := strings.Split(a.schemaValidationError, "\n")

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

		// Error highlighting
		errorStyle := style
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "error") {
			errorStyle = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
		} else if strings.Contains(lowerLine, "warning") {
			errorStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
		}

		a.drawText(x1+2, y, x2-1, y2-1, line, errorStyle)
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
		currentX += len(tab2Text)
	}

	// Tab 3: Schema Error (only when validation error exists)
	if a.schemaValidationError != "" {
		tab3Text := " Schema Error "
		tab3Style := tcell.StyleDefault.Foreground(tcell.ColorRed)
		if a.detailViewMode == "schema_error" {
			tab3Style = tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
		}
		for i, r := range tab3Text {
			a.screen.SetContent(currentX+i, tabY, r, nil, tab3Style)
		}
		a.schemaErrorBounds = PanelBounds{
			x1: currentX,
			y1: tabY,
			x2: currentX + len(tab3Text) - 1,
			y2: tabY,
		}
	}

	contentStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	y := y1 + 2

	// Display schema error content when in Schema Error mode
	if a.detailViewMode == "schema_error" && a.schemaValidationError != "" {
		a.drawSchemaErrorContent(x1, y1, x2, y2, y)
		return
	}

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
