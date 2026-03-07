package context

import (
	"fmt"
	"time"

	"github.com/dokadev/lazyprisma/pkg/gui/style"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// Frame and title styling constants (matching app.panel.go values)
var (
	outputDefaultFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	outputPrimaryFrameColor = gocui.ColorWhite
	outputFocusedFrameColor = gocui.ColorGreen

	outputPrimaryTitleColor = gocui.ColorWhite | gocui.AttrNone
	outputFocusedTitleColor = gocui.ColorGreen | gocui.AttrBold
)

type OutputContext struct {
	*SimpleContext
	*ScrollableTrait

	g       *gocui.Gui
	tr      *i18n.TranslationSet
	content string
	subtitle string
	focused  bool
	autoScrollToBottom bool
}

var _ types.Context = &OutputContext{}
var _ types.IScrollableContext = &OutputContext{}

type OutputContextOpts struct {
	Gui          *gocui.Gui
	Tr           *i18n.TranslationSet
	ViewName     string
}

func NewOutputContext(opts OutputContextOpts) *OutputContext {
	baseCtx := NewBaseContext(BaseContextOpts{
		Key:       types.ContextKey(opts.ViewName),
		Kind:      types.MAIN_CONTEXT,
		ViewName:  opts.ViewName,
		Focusable: true,
		Title:     opts.Tr.PanelTitleOutput,
	})

	simpleCtx := NewSimpleContext(baseCtx)

	oc := &OutputContext{
		SimpleContext:  simpleCtx,
		ScrollableTrait: &ScrollableTrait{},
		g:              opts.Gui,
		tr:             opts.Tr,
		content:        "",
	}

	return oc
}

// ID returns the view identifier (implements Panel interface from app package)
func (o *OutputContext) ID() string {
	return o.GetViewName()
}

// Draw renders the output panel (implements Panel interface from app package)
func (o *OutputContext) Draw(dim boxlayout.Dimensions) error {
	v, err := o.g.SetView(o.GetViewName(), dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	// Setup view (replicates BasePanel.SetupView)
	o.setupView(v)
	o.SetView(v)           // BaseContext
	o.ScrollableTrait.SetView(v) // ScrollableTrait

	v.Subtitle = o.subtitle
	v.Wrap = true
	fmt.Fprint(v, o.content)

	// Auto-scroll to bottom if flagged
	if o.autoScrollToBottom {
		contentLines := len(v.ViewBufferLines())
		_, viewHeight := v.Size()
		innerHeight := viewHeight - 2
		maxOrigin := contentLines - innerHeight
		if maxOrigin < 0 {
			maxOrigin = 0
		}
		o.ScrollableTrait.SetOriginY(maxOrigin)
		o.autoScrollToBottom = false
	}

	// Adjust scroll and apply origin
	o.ScrollableTrait.AdjustScroll()

	return nil
}

// setupView configures the view with common settings (replaces BasePanel.SetupView)
func (o *OutputContext) setupView(v *gocui.View) {
	v.Clear()
	v.Frame = true
	v.Title = o.tr.PanelTitleOutput
	v.FrameRunes = outputDefaultFrameRunes

	if o.focused {
		v.FrameColor = outputFocusedFrameColor
		v.TitleColor = outputFocusedTitleColor
	} else {
		v.FrameColor = outputPrimaryFrameColor
		v.TitleColor = outputPrimaryTitleColor
	}
}

// OnFocus handles focus gain (implements Panel interface from app package)
func (o *OutputContext) OnFocus() {
	o.focused = true
	if v := o.GetView(); v != nil {
		v.FrameColor = outputFocusedFrameColor
		v.TitleColor = outputFocusedTitleColor
	}
}

// OnBlur handles focus loss (implements Panel interface from app package)
func (o *OutputContext) OnBlur() {
	o.focused = false
	if v := o.GetView(); v != nil {
		v.FrameColor = outputPrimaryFrameColor
		v.TitleColor = outputPrimaryTitleColor
	}
}

// AppendOutput appends text to the output buffer and flags auto-scroll
func (o *OutputContext) AppendOutput(text string) {
	o.content += text + "\n"
	o.autoScrollToBottom = true
}

// LogAction logs an action with timestamp and optional details
func (o *OutputContext) LogAction(action string, details ...string) {
	timestamp := time.Now().Format("15:04:05")

	if o.content != "" {
		o.content += "\n"
	}

	header := fmt.Sprintf("%s %s", style.Gray(timestamp), style.CyanBold(action))
	o.content += header + "\n"

	for _, detail := range details {
		o.content += "  " + detail + "\n"
	}

	o.autoScrollToBottom = true
}

// LogActionRed logs an action in red (for errors/warnings)
func (o *OutputContext) LogActionRed(action string, details ...string) {
	timestamp := time.Now().Format("15:04:05")

	if o.content != "" {
		o.content += "\n"
	}

	header := fmt.Sprintf("%s %s", style.Gray(timestamp), style.RedBold(action))
	o.content += header + "\n"

	for _, detail := range details {
		o.content += "  " + style.Red(detail) + "\n"
	}

	o.autoScrollToBottom = true
}

// SetSubtitle sets the custom subtitle for the panel
func (o *OutputContext) SetSubtitle(subtitle string) {
	o.subtitle = subtitle
}

// ScrollUpByWheel scrolls up by wheel increment (delegates to ScrollableTrait)
// This method is provided for backward compatibility with existing callers
// that pass no arguments (the old OutputPanel signature).
func (o *OutputContext) ScrollUpByWheel() {
	o.ScrollableTrait.ScrollUpByWheel()
}

// ScrollDownByWheel scrolls down by wheel increment (delegates to ScrollableTrait)
func (o *OutputContext) ScrollDownByWheel() {
	o.ScrollableTrait.ScrollDownByWheel()
}
