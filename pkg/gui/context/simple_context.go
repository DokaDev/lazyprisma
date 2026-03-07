package context

import (
	"github.com/dokadev/lazyprisma/pkg/gui/types"
)

type SimpleContext struct {
	*BaseContext
	onRenderFn func()
}

var _ types.Context = &SimpleContext{}

func NewSimpleContext(baseContext *BaseContext) *SimpleContext {
	return &SimpleContext{
		BaseContext: baseContext,
	}
}

// SetOnRenderFn sets the function called during HandleRender.
func (self *SimpleContext) SetOnRenderFn(fn func()) {
	self.onRenderFn = fn
}

// HandleFocus is called when this context gains focus.
// It invokes all registered onFocusFns in order.
func (self *SimpleContext) HandleFocus(opts types.OnFocusOpts) {
	for _, fn := range self.onFocusFns {
		fn(opts)
	}
}

// HandleFocusLost is called when this context loses focus.
// It invokes all registered onFocusLostFns in order.
func (self *SimpleContext) HandleFocusLost(opts types.OnFocusLostOpts) {
	for _, fn := range self.onFocusLostFns {
		fn(opts)
	}
}

// HandleRender is called when the context needs to re-render its content.
func (self *SimpleContext) HandleRender() {
	if self.onRenderFn != nil {
		self.onRenderFn()
	}
}
