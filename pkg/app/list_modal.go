package app

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// ListModalItem represents a selectable item in the list modal
type ListModalItem struct {
	Label       string      // Display text in the list
	Description string      // Description shown in the bottom view
	OnSelect    func() error // Callback when item is selected with Enter
}

// ListModal displays a list of items with descriptions
type ListModal struct {
	g            *gocui.Gui
	title        string
	items        []ListModalItem
	selectedIdx  int
	originY      int // Scroll position for list view
	width        int
	height       int
	style        MessageModalStyle
	onCancel     func()
}

// NewListModal creates a new list modal
func NewListModal(g *gocui.Gui, title string, items []ListModalItem, onCancel func()) *ListModal {
	return &ListModal{
		g:           g,
		title:       title,
		items:       items,
		selectedIdx: 0,
		style:       MessageModalStyle{}, // Default style
		onCancel:    onCancel,
	}
}

// WithStyle sets the modal style
func (m *ListModal) WithStyle(style MessageModalStyle) *ListModal {
	m.style = style
	return m
}

// ID returns the modal's view ID
func (m *ListModal) ID() string {
	return "list_modal"
}

// listViewID returns the list view ID
func (m *ListModal) listViewID() string {
	return "list_modal_list"
}

// descViewID returns the description view ID
func (m *ListModal) descViewID() string {
	return "list_modal_desc"
}

// Draw renders the list modal with two views (list on top, description on bottom)
func (m *ListModal) Draw(dim boxlayout.Dimensions) error {
	// Get screen size
	screenWidth, screenHeight := m.g.Size()

	// Calculate width (5/7 of screen, min 80)
	m.width = 5 * screenWidth / 7
	minWidth := 80
	if m.width < minWidth {
		if screenWidth-2 < minWidth {
			m.width = screenWidth - 2
		} else {
			m.width = minWidth
		}
	}

	// Calculate description height dynamically based on selected item's description
	availableWidth := m.width - 4 // Minus frame and padding
	var descContentLines int
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		desc := m.items[m.selectedIdx].Description
		wrappedLines := m.wrapText(desc, availableWidth)
		descContentLines = len(wrappedLines)
	}

	// Calculate heights
	// List view: number of items + 2 (borders)
	// Description view: actual content lines + 2 (borders)
	listHeight := len(m.items) + 2
	descHeight := descContentLines + 2

	m.height = listHeight + descHeight + 1 // +1 for gap

	// Don't exceed screen height
	maxHeight := screenHeight - 4
	if m.height > maxHeight {
		m.height = maxHeight
		// If total exceeds screen, adjust desc height (list height stays for items)
		descHeight = m.height - listHeight
		if descHeight < 4 {
			// Minimum desc height is 4 (2 borders + 2 content lines)
			descHeight = 4
			listHeight = m.height - descHeight
		}
	}

	// Center the modal
	x0 := (screenWidth - m.width) / 2
	y0 := (screenHeight - m.height) / 2

	// Draw list view (top)
	listX0 := x0
	listY0 := y0
	listX1 := x0 + m.width
	listY1 := y0 + listHeight

	if err := m.drawListView(listX0, listY0, listX1, listY1); err != nil {
		return err
	}

	// Draw description view (bottom)
	descX0 := x0
	descY0 := listY1 + 1 // One line gap
	descX1 := x0 + m.width
	descY1 := y0 + m.height

	if err := m.drawDescView(descX0, descY0, descX1, descY1); err != nil {
		return err
	}

	return nil
}

// drawListView renders the list view (top)
func (m *ListModal) drawListView(x0, y0, x1, y1 int) error {
	v, err := m.g.SetView(m.listViewID(), x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Clear()
	v.Frame = true
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
	v.Title = " " + m.title + " "
	v.Footer = ""

	// Apply frame color (border) if set
	if m.style.BorderColor != ColorDefault {
		v.FrameColor = gocui.Attribute(colorToAnsiCode(m.style.BorderColor))
	}

	// Apply title color if set
	if m.style.TitleColor != ColorDefault {
		v.TitleColor = gocui.Attribute(colorToAnsiCode(m.style.TitleColor))
	}

	v.Wrap = false

	// Enable highlight for selection (like MigrationsPanel)
	v.Highlight = true
	v.SelBgColor = SelectionBgColor

	// Render list items
	for _, item := range m.items {
		fmt.Fprintln(v, item.Label)
	}

	// Adjust origin to ensure it's within valid bounds
	AdjustOrigin(v, &m.originY)

	// Set cursor position to selected item (like MigrationsPanel)
	v.SetCursor(0, m.selectedIdx-m.originY)
	v.SetOrigin(0, m.originY)

	return nil
}

// drawDescView renders the description view (bottom)
func (m *ListModal) drawDescView(x0, y0, x1, y1 int) error {
	v, err := m.g.SetView(m.descViewID(), x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	v.Clear()
	v.Frame = true
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
	v.Title = ""
	v.Footer = " [↑/↓] Navigate [Enter] Select [ESC] Cancel "

	// Apply frame color (border) if set
	if m.style.BorderColor != ColorDefault {
		v.FrameColor = gocui.Attribute(colorToAnsiCode(m.style.BorderColor))
	}

	v.Wrap = true

	// Render description for selected item
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		desc := m.items[m.selectedIdx].Description

		// Word wrap description
		availableWidth := (x1 - x0) - 4 // Minus frame and padding
		wrappedLines := m.wrapText(desc, availableWidth)

		for _, line := range wrappedLines {
			fmt.Fprintln(v, "  "+line)
		}
	}

	return nil
}

// wrapText wraps text to fit within the specified width
func (m *ListModal) wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if len(para) == 0 {
			lines = append(lines, "")
			continue
		}

		if len(para) <= width {
			lines = append(lines, para)
		} else {
			// Simple word wrapping
			words := strings.Fields(para)
			currentLine := ""

			for _, word := range words {
				if len(currentLine)+len(word)+1 <= width {
					if currentLine == "" {
						currentLine = word
					} else {
						currentLine += " " + word
					}
				} else {
					// Current line is full, start new line
					lines = append(lines, currentLine)
					currentLine = word
				}
			}

			// Add remaining line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
		}
	}

	return lines
}

// HandleKey handles keyboard input
func (m *ListModal) HandleKey(key any, mod gocui.Modifier) error {
	switch key {
	case gocui.KeyArrowUp:
		m.selectPrev()
	case gocui.KeyArrowDown:
		m.selectNext()
	case gocui.KeyEnter:
		return m.onEnter()
	case gocui.KeyEsc, 'q':
		if m.onCancel != nil {
			m.onCancel()
		}
		return nil
	}

	return nil
}

// selectNext selects the next item (circular)
func (m *ListModal) selectNext() {
	if len(m.items) == 0 {
		return
	}
	m.selectedIdx++
	if m.selectedIdx >= len(m.items) {
		m.selectedIdx = 0 // Wrap to first item
		m.originY = 0     // Reset scroll position when wrapping
	} else {
		// Auto-scroll if needed (like MigrationsPanel)
		v, err := m.g.View(m.listViewID())
		if err == nil {
			_, h := v.Size()
			innerHeight := h - 2 // Subtract frame borders
			if m.selectedIdx-m.originY >= innerHeight {
				m.originY++
			}
		}
	}
	// Redraw to update selection
	m.g.Update(func(g *gocui.Gui) error {
		return nil
	})
}

// selectPrev selects the previous item (circular)
func (m *ListModal) selectPrev() {
	if len(m.items) == 0 {
		return
	}
	m.selectedIdx--
	if m.selectedIdx < 0 {
		m.selectedIdx = len(m.items) - 1 // Wrap to last item
		// Scroll to bottom when wrapping
		v, err := m.g.View(m.listViewID())
		if err == nil {
			_, h := v.Size()
			innerHeight := h - 2
			m.originY = len(m.items) - innerHeight
			if m.originY < 0 {
				m.originY = 0
			}
		}
	} else {
		// Auto-scroll if needed (like MigrationsPanel)
		if m.selectedIdx < m.originY {
			m.originY--
		}
	}
	// Redraw to update selection
	m.g.Update(func(g *gocui.Gui) error {
		return nil
	})
}

// onEnter executes the callback for the selected item
func (m *ListModal) onEnter() error {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		if m.items[m.selectedIdx].OnSelect != nil {
			return m.items[m.selectedIdx].OnSelect()
		}
	}
	return nil
}

// OnClose is called when the modal is closed
func (m *ListModal) OnClose() {
	// Delete all views
	m.g.DeleteView(m.listViewID())
	m.g.DeleteView(m.descViewID())
}
