package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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
