package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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
	versionText := prismaStatus + " " + a.status.Version
	if a.status.IsGlobal {
		versionText += " (Global)"
	}
	prismaLine.extra = append(prismaLine.extra, struct {
		x     int
		text  string
		style tcell.Style
	}{x: 9, text: versionText, style: prismaStyle})
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
