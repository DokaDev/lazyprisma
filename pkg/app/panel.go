package app

import (
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

// Frame and title styling
var (
	defaultFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	PrimaryFrameColor = gocui.ColorWhite
	FocusedFrameColor = gocui.ColorGreen

	PrimaryTitleColor = gocui.ColorWhite | gocui.AttrNone
	FocusedTitleColor = gocui.ColorGreen | gocui.AttrBold

	// Tab styling
	FocusedActiveTabColor = gocui.ColorGreen | gocui.AttrBold // Active tab when panel is focused
	PrimaryActiveTabColor = gocui.ColorGreen | gocui.AttrNone // Active tab when panel is not focused

	// List selection color
	SelectionBgColor = gocui.ColorBlue
)

func NewBasePanel(id string, g *gocui.Gui) BasePanel {
	return BasePanel{
		id:         id,
		g:          g,
		frameRunes: defaultFrameRunes,
	}
}

func (bp *BasePanel) ID() string {
	return bp.id
}

func (bp *BasePanel) OnFocus() {
	bp.focused = true
	if bp.v != nil {
		bp.v.FrameColor = FocusedFrameColor
		bp.v.TitleColor = FocusedTitleColor
	}
}

func (bp *BasePanel) OnBlur() {
	bp.focused = false
	if bp.v != nil {
		bp.v.FrameColor = PrimaryFrameColor
		bp.v.TitleColor = PrimaryTitleColor
	}
}

// SetupView는 공통 뷰 설정을 처리합니다
func (bp *BasePanel) SetupView(v *gocui.View, title string) {
	bp.v = v
	v.Clear()
	v.Frame = true
	v.Title = title
	v.FrameRunes = bp.frameRunes

	if bp.focused {
		v.FrameColor = FocusedFrameColor
		v.TitleColor = FocusedTitleColor
	} else {
		v.FrameColor = PrimaryFrameColor
		v.TitleColor = PrimaryTitleColor
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
