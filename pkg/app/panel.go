package app

import (
	"github.com/dokadev/lazyprisma/pkg/gui/style"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

type Panel interface {
	ID() string
	Draw(dim boxlayout.Dimensions) error
	OnFocus()
	OnBlur()
}

type BasePanel struct {
	id         string
	g          *gocui.Gui
	v          *gocui.View
	focused    bool
	frameRunes []rune
}

func NewBasePanel(id string, g *gocui.Gui) BasePanel {
	return BasePanel{
		id:         id,
		g:          g,
		frameRunes: style.DefaultFrameRunes,
	}
}

func (bp *BasePanel) ID() string {
	return bp.id
}

func (bp *BasePanel) OnFocus() {
	bp.focused = true
	if bp.v != nil {
		bp.v.FrameColor = style.FocusedFrameColor
		bp.v.TitleColor = style.FocusedTitleColor
	}
}

func (bp *BasePanel) OnBlur() {
	bp.focused = false
	if bp.v != nil {
		bp.v.FrameColor = style.PrimaryFrameColor
		bp.v.TitleColor = style.PrimaryTitleColor
	}
}

// SetupView handles common view setup
func (bp *BasePanel) SetupView(v *gocui.View, title string) {
	bp.v = v
	v.Clear()
	v.Frame = true
	v.Title = title
	v.FrameRunes = bp.frameRunes

	if bp.focused {
		v.FrameColor = style.FocusedFrameColor
		v.TitleColor = style.FocusedTitleColor
	} else {
		v.FrameColor = style.PrimaryFrameColor
		v.TitleColor = style.PrimaryTitleColor
	}
}

// AdjustOrigin adjusts the origin to ensure it's within valid bounds
// Call this after content is rendered but before SetOrigin
func AdjustOrigin(v *gocui.View, originY *int) {
	if v == nil || originY == nil {
		return
	}

	// Get actual content lines from the rendered view buffer
	contentLines := len(v.ViewBufferLines())
	_, viewHeight := v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	// Adjust origin if it exceeds maxOrigin (e.g., after terminal resize)
	if *originY > maxOrigin {
		*originY = maxOrigin
	}
}
