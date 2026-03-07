package app

import (
	"fmt"

	"github.com/dokadev/lazyprisma/pkg/i18n"
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
	*BaseModal
	title        string
	contentLines []string // Original content lines
	lines        []string // Wrapped content lines
	width        int
	height       int
}

// NewMessageModal creates a new message modal
func NewMessageModal(g *gocui.Gui, tr *i18n.TranslationSet, title string, lines ...string) *MessageModal {
	return &MessageModal{
		BaseModal:    NewBaseModal("modal", g, tr),
		title:        title,
		contentLines: lines,
	}
}

// WithStyle sets the modal style
func (m *MessageModal) WithStyle(style MessageModalStyle) *MessageModal {
	m.SetStyle(style)
	return m
}

// Draw renders the modal
func (m *MessageModal) Draw(dim boxlayout.Dimensions) error {
	// Calculate width
	m.width = m.CalculateDimensions(4.0/7.0, 80)

	// Parse content into lines and calculate required height
	m.parseContent()

	// Calculate height based on content
	m.height = len(m.lines) + 1

	// Center the modal
	x0, y0, x1, y1 := m.CenterBox(m.width, m.height)

	// Create modal view
	v, _, err := m.SetupView(m.ID(), x0, y0, x1, y1, 0, " "+m.title+" ", m.tr.ModalFooterMessageClose)
	if err != nil {
		return err
	}

	v.Clear()
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

		wrapped := WrapText(line, availableWidth, "  ")
		m.lines = append(m.lines, wrapped...)
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
	m.BaseModal.OnClose()
}

// colorToAnsiCode converts Color to gocui color attribute
// Deprecated: use ColorToGocuiAttr instead. Kept for backward compatibility.
func colorToAnsiCode(c Color) int {
	return ColorToGocuiAttr(c)
}
