package app

import (
	"strings"

	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/jesseduffield/gocui"
)

// ModalStyle holds styling options for modals
// (renamed from MessageModalStyle for shared use across all modal types)
type ModalStyle = MessageModalStyle

// BaseModal provides common infrastructure shared by all modal types
type BaseModal struct {
	id    string
	g     *gocui.Gui
	tr    *i18n.TranslationSet
	style MessageModalStyle
}

// NewBaseModal creates a new BaseModal with default style
func NewBaseModal(id string, g *gocui.Gui, tr *i18n.TranslationSet) *BaseModal {
	return &BaseModal{
		id:    id,
		g:     g,
		tr:    tr,
		style: MessageModalStyle{},
	}
}

// SetStyle sets the modal style (used by embedding structs' WithStyle methods)
func (b *BaseModal) SetStyle(style MessageModalStyle) {
	b.style = style
}

// Style returns the current modal style
func (b *BaseModal) Style() MessageModalStyle {
	return b.style
}

// ID returns the modal's view ID
func (b *BaseModal) ID() string {
	return b.id
}

// CalculateDimensions computes centered coordinates for a modal view.
// widthRatio is the fraction of screen width (e.g., 4.0/7.0).
// heightContent is the number of content lines (borders added by caller).
// minWidth is the minimum width for the modal.
func (b *BaseModal) CalculateDimensions(widthRatio float64, minWidth int) (width int) {
	screenWidth, _ := b.g.Size()

	width = int(widthRatio * float64(screenWidth))
	if width < minWidth {
		if screenWidth-2 < minWidth {
			width = screenWidth - 2
		} else {
			width = minWidth
		}
	}

	return width
}

// CenterBox returns centered screen coordinates for a box of the given width and height.
func (b *BaseModal) CenterBox(width, height int) (x0, y0, x1, y1 int) {
	screenWidth, screenHeight := b.g.Size()

	// Clamp height to screen
	maxHeight := screenHeight - 4
	if height > maxHeight {
		height = maxHeight
	}

	x0 = (screenWidth - width) / 2
	y0 = (screenHeight - height) / 2
	x1 = x0 + width
	y1 = y0 + height
	return
}

// SetupView creates (or retrieves) a gocui view and applies common frame settings.
// It handles the "unknown view" error from SetView, applies frame runes, title, footer,
// and style colours. Returns the view and whether it was newly created.
func (b *BaseModal) SetupView(name string, x0, y0, x1, y1 int, zIndex byte, title, footer string) (*gocui.View, bool, error) {
	v, err := b.g.SetView(name, x0, y0, x1, y1, zIndex)
	isNew := err != nil
	if err != nil && err.Error() != "unknown view" {
		return nil, false, err
	}

	v.Frame = true
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
	v.Title = title
	v.Footer = footer

	// Apply frame color (border) if set
	if b.style.BorderColor != ColorDefault {
		v.FrameColor = gocui.Attribute(ColorToGocuiAttr(b.style.BorderColor))
	}

	// Apply title color if set
	if b.style.TitleColor != ColorDefault {
		v.TitleColor = gocui.Attribute(ColorToGocuiAttr(b.style.TitleColor))
	}

	return v, isNew, nil
}

// OnClose deletes the modal's primary view
func (b *BaseModal) OnClose() {
	b.g.DeleteView(b.id)
}

// ColorToGocuiAttr converts a Color to a gocui color attribute value.
// Exported so it can be used by any code that needs this conversion.
func ColorToGocuiAttr(c Color) int {
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

// WrapText wraps text to fit within the specified width.
// Each resulting line is prefixed with the given padding string.
// Handles multiple paragraphs separated by newlines.
func WrapText(text string, maxWidth int, padding string) []string {
	if maxWidth <= 0 {
		return []string{padding + text}
	}

	var lines []string
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if len(para) == 0 {
			lines = append(lines, "")
			continue
		}

		if len(para) <= maxWidth {
			lines = append(lines, padding+para)
		} else {
			// Word wrapping
			words := strings.Fields(para)
			currentLine := padding

			for _, word := range words {
				if len(currentLine)+len(word)+1 <= maxWidth+len(padding) {
					if currentLine == padding {
						currentLine += word
					} else {
						currentLine += " " + word
					}
				} else {
					lines = append(lines, currentLine)
					currentLine = padding + word
				}
			}

			if currentLine != padding {
				lines = append(lines, currentLine)
			}
		}
	}

	return lines
}
