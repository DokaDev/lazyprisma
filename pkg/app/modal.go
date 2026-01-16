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
}
