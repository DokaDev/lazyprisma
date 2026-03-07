package types

import (
	"github.com/jesseduffield/gocui"
)

// ContextKey uniquely identifies a context.
type ContextKey string

// ContextKind categorises contexts by their role in the layout.
type ContextKind int

const (
	// SIDE_CONTEXT is a panel on the left-hand side (workspace, migrations).
	SIDE_CONTEXT ContextKind = iota
	// MAIN_CONTEXT is the main content area (details, output).
	MAIN_CONTEXT
	// TEMPORARY_POPUP is a transient popup (confirm, prompt, menu, message).
	TEMPORARY_POPUP
)

// OnFocusOpts carries information when a context gains focus.
type OnFocusOpts struct {
	ClickedViewLineIdx int
}

// OnFocusLostOpts carries information when a context loses focus.
type OnFocusLostOpts struct {
	NewContextKey ContextKey
}

// IBaseContext defines the minimal identity and metadata for a context.
type IBaseContext interface {
	GetKey() ContextKey
	GetKind() ContextKind
	GetViewName() string
	GetView() *gocui.View
	IsFocusable() bool
	Title() string
}

// Context extends IBaseContext with lifecycle hooks.
type Context interface {
	IBaseContext

	HandleFocus(opts OnFocusOpts)
	HandleFocusLost(opts OnFocusLostOpts)
	HandleRender()
}

// IListContext is a context that presents a selectable list of items.
type IListContext interface {
	Context

	GetSelectedIdx() int
	GetItemCount() int
	SelectNext()
	SelectPrev()
}

// ITabbedContext is a context that supports tabbed sub-views.
type ITabbedContext interface {
	Context

	NextTab()
	PrevTab()
	GetCurrentTab() int
}

// IScrollableContext is a context that supports vertical scrolling.
type IScrollableContext interface {
	Context

	ScrollUp()
	ScrollDown()
	ScrollToTop()
	ScrollToBottom()
}
