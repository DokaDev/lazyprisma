package app

import (
	"strings"

	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// InputModal displays an input field for user text entry
type InputModal struct {
	*BaseModal
	title            string // Used as placeholder
	subtitle         string // Optional subtitle
	footer           string // Key bindings description
	width            int
	height           int
	onSubmit         func(string)
	onCancel         func()
	required         bool
	onValidationFail func(string)
}

// NewInputModal creates a new input modal
func NewInputModal(g *gocui.Gui, tr *i18n.TranslationSet, title string, onSubmit func(string), onCancel func()) *InputModal {
	return &InputModal{
		BaseModal: NewBaseModal("input_modal", g, tr),
		title:     title,
		footer:    tr.ModalFooterInputSubmitCancel,
		onSubmit:  onSubmit,
		onCancel:  onCancel,
	}
}

// WithStyle sets the modal style
func (m *InputModal) WithStyle(style MessageModalStyle) *InputModal {
	m.SetStyle(style)
	return m
}

// WithSubtitle sets the modal subtitle
func (m *InputModal) WithSubtitle(subtitle string) *InputModal {
	m.subtitle = subtitle
	return m
}

// WithRequired sets whether the input is required (non-empty)
func (m *InputModal) WithRequired(required bool) *InputModal {
	m.required = required
	return m
}

// OnValidationFail sets the callback for validation failures
func (m *InputModal) OnValidationFail(callback func(string)) *InputModal {
	m.onValidationFail = callback
	return m
}

// AcceptsTextInput returns true because InputModal uses keyboard for text entry.
func (m *InputModal) AcceptsTextInput() bool { return true }

// Draw renders the input modal
func (m *InputModal) Draw(dim boxlayout.Dimensions) error {
	// Calculate width
	m.width = m.CalculateDimensions(4.0/7.0, 80)

	// Height for input modal: minimal single line
	m.height = 2

	// Center the modal
	x0, y0, x1, y1 := m.CenterBox(m.width, m.height)

	// Create input view
	v, isNew, err := m.SetupView(m.ID(), x0, y0, x1, y1, 0, " "+m.title+" ", m.footer)
	if err != nil {
		return err
	}

	// Only clear on first creation (TextArea manages content)
	if isNew {
		v.Clear()
		// Initial render to make footer visible
		v.RenderTextArea()
	}

	if m.subtitle != "" {
		v.Subtitle = " " + m.subtitle + " "
	}

	// Input field settings (CRITICAL - DO NOT CHANGE)
	v.Editable = true
	v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
		// Single-line input: block Enter key (pass to HandleKey instead)
		if key == gocui.KeyEnter {
			return false
		}

		// Use DefaultEditor.Edit() for other keys
		// Note: DefaultEditor already calls v.RenderTextArea() internally
		return gocui.DefaultEditor.Edit(v, key, ch, mod)
	})
	v.Wrap = false
	v.Autoscroll = false

	// Enable cursor at Gui level
	m.g.Cursor = true

	return nil
}

// HandleKey handles keyboard input
func (m *InputModal) HandleKey(key any, mod gocui.Modifier) error {
	// Submit on Enter
	if key == gocui.KeyEnter {
		v, err := m.g.View(m.ID())
		if err != nil {
			return err
		}

		// Get input value from TextArea and trim whitespace
		input := strings.TrimSpace(v.TextArea.GetContent())

		// Validate if required
		if m.required && input == "" {
			if m.onValidationFail != nil {
				m.onValidationFail(m.tr.ModalMsgInputRequired)
			}
			return nil // Don't submit
		}

		// Submit valid input
		if m.onSubmit != nil {
			m.onSubmit(input)
		}
		return nil
	}

	// Cancel on ESC only (not 'q', which is used for input)
	if key == gocui.KeyEsc {
		if m.onCancel != nil {
			m.onCancel()
		}
		return nil
	}

	// Let other keys pass through to editor for input
	return nil
}

// OnClose is called when the modal is closed
func (m *InputModal) OnClose() {
	// Disable cursor at Gui level
	m.g.Cursor = false
	// Delete the input modal view
	m.BaseModal.OnClose()
}
