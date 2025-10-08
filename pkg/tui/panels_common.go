package tui

import "github.com/gdamore/tcell/v2"

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
