package app

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// MessageModalStyle holds styling options for the modal
type MessageModalStyle struct {
	TitleColor  Color // Color for title (default: ColorDefault)
	BorderColor Color // Color for border (default: ColorDefault)
}

// MessageModal displays a message with title and content
type MessageModal struct {
	g            *gocui.Gui
	title        string
	contentLines []string // Original content lines
	lines        []string // Wrapped content lines
	width        int
	height       int
	style        MessageModalStyle
}

// NewMessageModal creates a new message modal
func NewMessageModal(g *gocui.Gui, title string, lines ...string) *MessageModal {
	return &MessageModal{
		g:            g,
		title:        title,
		contentLines: lines,
		style:        MessageModalStyle{}, // Default style
	}
}

// WithStyle sets the modal style
func (m *MessageModal) WithStyle(style MessageModalStyle) *MessageModal {
	m.style = style
	return m
}

// ID returns the modal's view ID
func (m *MessageModal) ID() string {
	return "modal"
}

// Draw renders the modal
func (m *MessageModal) Draw(dim boxlayout.Dimensions) error {
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

	// Parse content into lines and calculate required height
	m.parseContent()

	// Calculate height based on content
	// Content + 2 (top and bottom borders with title/footer)
	contentHeight := len(m.lines)
	// m.height = contentHeight + 2
	m.height = contentHeight + 1

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
	v.Footer = " [Enter/q/ESC] Close "

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
	for _, line := range m.lines {
		fmt.Fprintln(v, line)
	}

	return nil
}

// parseContent wraps long lines for display
func (m *MessageModal) parseContent() {
	m.lines = []string{}

	// Available width for content (minus frame and padding)
	availableWidth := m.width - 4

	for _, line := range m.contentLines {
		if len(line) == 0 {
			m.lines = append(m.lines, "")
			continue
		}

		// Word wrap long lines
		if len(line) <= availableWidth {
			m.lines = append(m.lines, "  "+line)
		} else {
			// Simple word wrapping
			words := strings.Fields(line)
			currentLine := "  "

			for _, word := range words {
				if len(currentLine)+len(word)+1 <= availableWidth+2 { // +2 for initial "  "
					if currentLine == "  " {
						currentLine += word
					} else {
						currentLine += " " + word
					}
				} else {
					// Current line is full, start new line
					m.lines = append(m.lines, currentLine)
					currentLine = "  " + word
				}
			}

			// Add remaining line
			if currentLine != "  " {
				m.lines = append(m.lines, currentLine)
			}
		}
	}
}

// HandleKey handles keyboard input
func (m *MessageModal) HandleKey(key any, mod gocui.Modifier) error {
	// Close on 'q', ESC, or Enter
	if key == 'q' || key == gocui.KeyEsc || key == gocui.KeyEnter {
		// Modal will be closed by App
		return nil
	}

	return nil
}

// OnClose is called when the modal is closed
func (m *MessageModal) OnClose() {
	// Delete the modal view
	m.g.DeleteView(m.ID())
}

// colorToAnsiCode converts Color to gocui color attribute
func colorToAnsiCode(c Color) int {
	switch c {
	case ColorBlack:
		return int(gocui.ColorBlack)
	case ColorRed:
		return int(gocui.ColorRed)
	case ColorGreen:
		return int(gocui.ColorGreen)
	case ColorYellow:
		return int(gocui.ColorYellow)
	case ColorBlue:
		return int(gocui.ColorBlue)
	case ColorMagenta:
		return int(gocui.ColorMagenta)
	case ColorCyan:
		return int(gocui.ColorCyan)
	case ColorWhite:
		return int(gocui.ColorWhite)
	default:
		return int(gocui.ColorDefault)
	}
}
