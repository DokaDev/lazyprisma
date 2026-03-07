package types

import (
	"github.com/jesseduffield/gocui"
)

// Key is an alias for any type that can represent a key (gocui.Key or rune).
type Key = any

// Binding maps a key press to a handler within a specific context.
type Binding struct {
	Key         Key
	Modifier    gocui.Modifier
	Handler     func() error
	Description string
	Tag         string // e.g. "navigation", used for grouping in help views
}

// KeybindingsFn is a function that returns a slice of key bindings.
type KeybindingsFn func() []*Binding

// IController is the interface that all controllers must implement.
// Each controller is associated with exactly one context and provides
// the keybindings for that context.
type IController interface {
	GetKeybindings() []*Binding
	Context() Context
}
