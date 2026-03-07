package context

import (
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/jesseduffield/gocui"
)

type BaseContext struct {
	key       types.ContextKey
	kind      types.ContextKind
	viewName  string
	view      *gocui.View
	focusable bool
	title     string

	// Lifecycle hooks (multiple can attach)
	onFocusFns     []func(types.OnFocusOpts)
	onFocusLostFns []func(types.OnFocusLostOpts)

	// Keybinding attachment
	keybindingsFns []types.KeybindingsFn
}

var _ types.IBaseContext = &BaseContext{}

type BaseContextOpts struct {
	Key       types.ContextKey
	Kind      types.ContextKind
	ViewName  string
	View      *gocui.View
	Focusable bool
	Title     string
}

func NewBaseContext(opts BaseContextOpts) *BaseContext {
	return &BaseContext{
		key:       opts.Key,
		kind:      opts.Kind,
		viewName:  opts.ViewName,
		view:      opts.View,
		focusable: opts.Focusable,
		title:     opts.Title,
	}
}

func (self *BaseContext) GetKey() types.ContextKey {
	return self.key
}

func (self *BaseContext) GetKind() types.ContextKind {
	return self.kind
}

func (self *BaseContext) GetViewName() string {
	if self.view != nil {
		return self.view.Name()
	}
	return self.viewName
}

func (self *BaseContext) GetView() *gocui.View {
	return self.view
}

func (self *BaseContext) SetView(v *gocui.View) {
	self.view = v
}

func (self *BaseContext) IsFocusable() bool {
	return self.focusable
}

func (self *BaseContext) Title() string {
	return self.title
}

// AddKeybindingsFn registers a function that provides keybindings for this context.
// Controllers call this to attach their bindings.
func (self *BaseContext) AddKeybindingsFn(fn types.KeybindingsFn) {
	self.keybindingsFns = append(self.keybindingsFns, fn)
}

// GetKeybindings collects all registered keybindings.
// Later-registered functions take precedence (appended in reverse order).
func (self *BaseContext) GetKeybindings() []*types.Binding {
	bindings := []*types.Binding{}
	for i := range self.keybindingsFns {
		bindings = append(bindings, self.keybindingsFns[len(self.keybindingsFns)-1-i]()...)
	}
	return bindings
}

// AddOnFocusFn registers a lifecycle hook called when this context gains focus.
func (self *BaseContext) AddOnFocusFn(fn func(types.OnFocusOpts)) {
	if fn != nil {
		self.onFocusFns = append(self.onFocusFns, fn)
	}
}

// AddOnFocusLostFn registers a lifecycle hook called when this context loses focus.
func (self *BaseContext) AddOnFocusLostFn(fn func(types.OnFocusLostOpts)) {
	if fn != nil {
		self.onFocusLostFns = append(self.onFocusLostFns, fn)
	}
}
