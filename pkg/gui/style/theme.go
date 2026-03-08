package style

import "github.com/jesseduffield/gocui"

// DefaultFrameRunes defines the standard rounded-corner frame characters
// used by all panels and contexts.
var DefaultFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

// Frame and title colour constants shared across all panels/contexts.
var (
	PrimaryFrameColor = gocui.ColorWhite
	FocusedFrameColor = gocui.ColorGreen

	PrimaryTitleColor = gocui.ColorWhite | gocui.AttrNone
	FocusedTitleColor = gocui.ColorGreen | gocui.AttrBold

	// Tab styling
	FocusedActiveTabColor = gocui.ColorGreen | gocui.AttrBold
	PrimaryActiveTabColor = gocui.ColorGreen | gocui.AttrNone

	// List selection colour
	SelectionBgColor = gocui.ColorBlue
)
