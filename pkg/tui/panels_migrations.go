package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

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
