package app

import (
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// InputModal displays an input field for user text entry
type InputModal struct {
	g                *gocui.Gui
	title            string // Used as placeholder
	subtitle         string // Optional subtitle
	footer           string // Key bindings description
	width            int
	height           int
	style            MessageModalStyle
	onSubmit         func(string)
	onCancel         func()
	required         bool
	onValidationFail func(string)
}

// NewInputModal creates a new input modal
func NewInputModal(g *gocui.Gui, title string, onSubmit func(string), onCancel func()) *InputModal {
	return &InputModal{
		g:        g,
		title:    title,
		footer:   " [Enter] Submit [ESC] Cancel ",
		style:    MessageModalStyle{}, // Default style
		onSubmit: onSubmit,
		onCancel: onCancel,
	}
}

// WithStyle sets the modal style
func (m *InputModal) WithStyle(style MessageModalStyle) *InputModal {
	m.style = style
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

// ID returns the modal's view ID
func (m *InputModal) ID() string {
	return "input_modal"
}

// Draw renders the input modal
func (m *InputModal) Draw(dim boxlayout.Dimensions) error {
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

	// Height for input modal: minimal single line
	m.height = 2

	// Center the modal
	x0 := (screenWidth - m.width) / 2
	y0 := (screenHeight - m.height) / 2
	x1 := x0 + m.width
	y1 := y0 + m.height

	// Create input view
	v, err := m.g.SetView(m.ID(), x0, y0, x1, y1, 0)
	isNewView := err != nil
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	// Only clear on first creation (TextArea manages content)
	if isNewView {
		v.Clear()
		// Initial render to make footer visible
		v.RenderTextArea()
	}

	v.Frame = true
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
	v.Title = " " + m.title + " "
	if m.subtitle != "" {
		v.Subtitle = " " + m.subtitle + " "
	}
	v.Footer = m.footer

	// Apply frame color (border) if set
	if m.style.BorderColor != ColorDefault {
		v.FrameColor = gocui.Attribute(colorToAnsiCode(m.style.BorderColor))
	}

	// Apply title color if set
	if m.style.TitleColor != ColorDefault {
		v.TitleColor = gocui.Attribute(colorToAnsiCode(m.style.TitleColor))
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
				m.onValidationFail("Input is required")
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
	m.g.DeleteView(m.ID())
}
