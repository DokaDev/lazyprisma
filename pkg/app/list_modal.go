package app

import (
	"fmt"

	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// ListModalItem represents a selectable item in the list modal
type ListModalItem struct {
	Label       string       // Display text in the list
	Description string       // Description shown in the bottom view
	OnSelect    func() error // Callback when item is selected with Enter
}

// ListModal displays a list of items with descriptions
type ListModal struct {
	*BaseModal
	title       string
	items       []ListModalItem
	selectedIdx int
	originY     int // Scroll position for list view
	width       int
	height      int
	onCancel    func()
}

// NewListModal creates a new list modal
func NewListModal(g *gocui.Gui, tr *i18n.TranslationSet, title string, items []ListModalItem, onCancel func()) *ListModal {
	return &ListModal{
		BaseModal:   NewBaseModal("list_modal", g, tr),
		title:       title,
		items:       items,
		selectedIdx: 0,
		onCancel:    onCancel,
	}
}

// WithStyle sets the modal style
func (m *ListModal) WithStyle(style MessageModalStyle) *ListModal {
	m.SetStyle(style)
	return m
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
	// Calculate width (5/7 of screen, min 80)
	m.width = m.CalculateDimensions(5.0/7.0, 80)

	// Calculate description height dynamically based on selected item's description
	availableWidth := m.width - 4 // Minus frame and padding
	var descContentLines int
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		desc := m.items[m.selectedIdx].Description
		wrappedLines := WrapText(desc, availableWidth, "")
		descContentLines = len(wrappedLines)
	}

	// Calculate heights
	// List view: number of items + 2 (borders)
	// Description view: actual content lines + 2 (borders)
	listHeight := len(m.items) + 2
	descHeight := descContentLines + 2

	m.height = listHeight + descHeight + 1 // +1 for gap

	// Don't exceed screen height
	screenWidth, screenHeight := m.g.Size()
	_ = screenWidth
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
	x0, y0, _, _ := m.CenterBox(m.width, m.height)

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
	v, _, err := m.SetupView(m.listViewID(), x0, y0, x1, y1, 0, " "+m.title+" ", "")
	if err != nil {
		return err
	}

	v.Clear()
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
	v, _, err := m.SetupView(m.descViewID(), x0, y0, x1, y1, 0, "", m.tr.ModalFooterListNavigate)
	if err != nil {
		return err
	}

	v.Clear()
	v.Wrap = true

	// Render description for selected item
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		desc := m.items[m.selectedIdx].Description

		// Word wrap description
		availableWidth := (x1 - x0) - 4 // Minus frame and padding
		wrappedLines := WrapText(desc, availableWidth, "")

		for _, line := range wrappedLines {
			fmt.Fprintln(v, "  "+line)
		}
	}

	return nil
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
