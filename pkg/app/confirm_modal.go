package app

import (
	"fmt"

	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// ConfirmModal displays a confirmation dialog with Yes/No options
type ConfirmModal struct {
	*BaseModal
	title   string
	message string
	onYes   func()
	onNo    func()
	width   int
	height  int
}

// NewConfirmModal creates a new confirmation modal
func NewConfirmModal(g *gocui.Gui, tr *i18n.TranslationSet, title string, message string, onYes func(), onNo func()) *ConfirmModal {
	return &ConfirmModal{
		BaseModal: NewBaseModal("confirm_modal", g, tr),
		title:     title,
		message:   message,
		onYes:     onYes,
		onNo:      onNo,
	}
}

// WithStyle sets the modal style
func (m *ConfirmModal) WithStyle(style MessageModalStyle) *ConfirmModal {
	m.SetStyle(style)
	return m
}

// Draw renders the modal
func (m *ConfirmModal) Draw(dim boxlayout.Dimensions) error {
	// Calculate width
	m.width = m.CalculateDimensions(4.0/7.0, 80)

	// Parse message into lines
	availableWidth := m.width - 4
	lines := WrapText(m.message, availableWidth, "  ")

	// Calculate height based on content
	m.height = len(lines) + 2 // +2 for borders

	// Center the modal
	x0, y0, x1, y1 := m.CenterBox(m.width, m.height)

	// Create modal view
	v, _, err := m.SetupView(m.ID(), x0, y0, x1, y1, 0, " "+m.title+" ", m.tr.ModalFooterConfirmYesNo)
	if err != nil {
		return err
	}

	v.Clear()
	v.Wrap = false

	// Render content
	for _, line := range lines {
		fmt.Fprintln(v, line)
	}

	return nil
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
	m.BaseModal.OnClose()
}
