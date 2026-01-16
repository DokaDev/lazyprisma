package app

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// ConfirmModal displays a confirmation dialog with Yes/No options
type ConfirmModal struct {
	g       *gocui.Gui
	title   string
	message string
	onYes   func()
	onNo    func()
	width   int
	height  int
	style   MessageModalStyle
}

// NewConfirmModal creates a new confirmation modal
func NewConfirmModal(g *gocui.Gui, title string, message string, onYes func(), onNo func()) *ConfirmModal {
	return &ConfirmModal{
		g:       g,
		title:   title,
		message: message,
		onYes:   onYes,
		onNo:    onNo,
		style:   MessageModalStyle{}, // Default style
	}
}

// WithStyle sets the modal style
func (m *ConfirmModal) WithStyle(style MessageModalStyle) *ConfirmModal {
	m.style = style
	return m
}

// ID returns the modal's view ID
func (m *ConfirmModal) ID() string {
	return "confirm_modal"
}

// Draw renders the modal
func (m *ConfirmModal) Draw(dim boxlayout.Dimensions) error {
	// Get screen size
	screenWidth, screenHeight := m.g.Size()

	// Calculate width (4/7 of screen, min 80)
	m.width = 4 * screenWidth / 7
	minWidth := 80
	if m.width < minWidth {
		if screenWidth-2 < minWidth {
			m.width = screenWidth - 2
		} else {
			m.width = minWidth
		}
	}

	// Parse message into lines
	availableWidth := m.width - 4
	lines := m.wrapText(m.message, availableWidth)

	// Calculate height based on content
	contentHeight := len(lines)
	m.height = contentHeight + 2 // +2 for borders

	// Don't exceed screen height
	maxHeight := screenHeight - 4
	if m.height > maxHeight {
		m.height = maxHeight
	}

	// Center the modal
	x0 := (screenWidth - m.width) / 2
	y0 := (screenHeight - m.height) / 2
	x1 := x0 + m.width
	y1 := y0 + m.height

	// Create modal view
	v, err := m.g.SetView(m.ID(), x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Clear()
	v.Frame = true
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
	v.Title = " " + m.title + " "
	v.Footer = " [Y] Yes [N] No [ESC] Cancel "

	// Apply frame color (border) if set
	if m.style.BorderColor != ColorDefault {
		v.FrameColor = gocui.Attribute(colorToAnsiCode(m.style.BorderColor))
	}

	// Apply title color if set
	if m.style.TitleColor != ColorDefault {
		v.TitleColor = gocui.Attribute(colorToAnsiCode(m.style.TitleColor))
	}

	v.Wrap = false

	// Render content
	for _, line := range lines {
		fmt.Fprintln(v, line)
	}

	return nil
}

// wrapText wraps text to fit within the specified width
func (m *ConfirmModal) wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string

	if len(text) == 0 {
		lines = append(lines, "")
		return lines
	}

	// Word wrap
	if len(text) <= width {
		lines = append(lines, "  "+text)
	} else {
		// Simple word wrapping
		words := strings.Fields(text)
		currentLine := "  "

		for _, word := range words {
			if len(currentLine)+len(word)+1 <= width+2 { // +2 for initial "  "
				if currentLine == "  " {
					currentLine += word
				} else {
					currentLine += " " + word
				}
			} else {
				// Current line is full, start new line
				lines = append(lines, currentLine)
				currentLine = "  " + word
			}
		}

		// Add remaining line
		if currentLine != "  " {
			lines = append(lines, currentLine)
		}
	}

	return lines
}

// HandleKey handles keyboard input
func (m *ConfirmModal) HandleKey(key any, mod gocui.Modifier) error {
	switch key {
	case 'y', 'Y':
		// Yes - execute onYes callback
		if m.onYes != nil {
			m.onYes()
		}
		return nil
	case 'n', 'N', gocui.KeyEsc:
		// No - execute onNo callback
		if m.onNo != nil {
			m.onNo()
		}
		return nil
	}

	return nil
}

// OnClose is called when the modal is closed
func (m *ConfirmModal) OnClose() {
	// Delete the modal view
	m.g.DeleteView(m.ID())
}
