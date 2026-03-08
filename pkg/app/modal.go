package app

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// Modal represents a modal dialog
type Modal interface {
	ID() string
	Draw(dim boxlayout.Dimensions) error
	HandleKey(key any, mod gocui.Modifier) error
	OnClose()
	// AcceptsTextInput reports whether the modal uses keyboard input for text entry
	// (e.g. InputModal). When true, single-character keys like 'q' are not treated
	// as close commands.
	AcceptsTextInput() bool
	// ClosesOnEnter reports whether pressing Enter should close the modal
	// (e.g. MessageModal). When false, Enter is forwarded to HandleKey instead.
	ClosesOnEnter() bool
}
