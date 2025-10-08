package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

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
