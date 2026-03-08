package context

import (
	"github.com/dokadev/lazyprisma/pkg/gui/style"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/jesseduffield/gocui"
)

type BaseContext struct {
	key       types.ContextKey
	kind      types.ContextKind
	viewName  string
	view      *gocui.View
	focusable bool
	focused   bool
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

// IsFocused returns whether this context currently has focus.
func (self *BaseContext) IsFocused() bool {
	return self.focused
}

// SetFocused sets the focus state directly (without applying styles).
func (self *BaseContext) SetFocused(f bool) {
	self.focused = f
}

// ApplyFocusStyle sets the view's frame and title colours based on the
// current focus state. Safe to call when the view is nil.
func (self *BaseContext) ApplyFocusStyle() {
	if v := self.view; v != nil {
		if self.focused {
			v.FrameColor = style.FocusedFrameColor
			v.TitleColor = style.FocusedTitleColor
		} else {
			v.FrameColor = style.PrimaryFrameColor
			v.TitleColor = style.PrimaryTitleColor
		}
	}
}

// OnFocus marks this context as focused and applies the focused style.
func (self *BaseContext) OnFocus() {
	self.focused = true
	self.ApplyFocusStyle()
}

// OnBlur marks this context as unfocused and applies the primary style.
func (self *BaseContext) OnBlur() {
	self.focused = false
	self.ApplyFocusStyle()
}
