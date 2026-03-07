package context

import (
	"fmt"

	"github.com/dokadev/lazyprisma/pkg/gui/style"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// StatusBarState provides callbacks for accessing App state without direct dependency.
type StatusBarState struct {
	IsCommandRunning func() bool
	GetSpinnerFrame  func() uint32
	IsStudioRunning  func() bool
	GetCommandName   func() string
}

// StatusBarConfig holds static configuration for the status bar display.
type StatusBarConfig struct {
	Developer string
	Version   string
}

var spinnerFrames = []rune{'|', '/', '-', '\\'}

// SpinnerFrameCount returns the number of spinner animation frames.
func SpinnerFrameCount() uint32 {
	return uint32(len(spinnerFrames))
}

type StatusBarContext struct {
	*BaseContext

	g      *gocui.Gui
	tr     *i18n.TranslationSet
	state  StatusBarState
	config StatusBarConfig
}

type StatusBarContextOpts struct {
	Gui      *gocui.Gui
	Tr       *i18n.TranslationSet
	ViewName string
	State    StatusBarState
	Config   StatusBarConfig
}

func NewStatusBarContext(opts StatusBarContextOpts) *StatusBarContext {
	baseCtx := NewBaseContext(BaseContextOpts{
		Key:       types.ContextKey(opts.ViewName),
		Kind:      types.MAIN_CONTEXT,
		ViewName:  opts.ViewName,
		Focusable: false,
		Title:     "",
	})

	return &StatusBarContext{
		BaseContext: baseCtx,
		g:           opts.Gui,
		tr:          opts.Tr,
		state:       opts.State,
		config:      opts.Config,
	}
}

// ID returns the view identifier (implements Panel interface from app package)
func (s *StatusBarContext) ID() string {
	return s.GetViewName()
}

// Draw renders the status bar (implements Panel interface from app package)
func (s *StatusBarContext) Draw(dim boxlayout.Dimensions) error {
	// StatusBar has no frame, so adjust dimensions
	frameOffset := 1
	x0 := dim.X0 - frameOffset
	y0 := dim.Y0 - frameOffset
	x1 := dim.X1 + frameOffset
	y1 := dim.Y1 + frameOffset

	v, err := s.g.SetView(s.GetViewName(), x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	s.SetView(v)
	v.Clear()
	v.Frame = false

	// Build status bar content
	var leftContent string
	var visibleLen int

	// Show spinner if command is running
	if s.state.IsCommandRunning() {
		frameIndex := s.state.GetSpinnerFrame() % uint32(len(spinnerFrames))
		spinner := string(spinnerFrames[frameIndex])

		// Get running task name
		taskName := s.state.GetCommandName()

		leftContent = fmt.Sprintf(" %s %s ", style.Cyan(spinner), style.Gray(taskName))
		visibleLen += 1 + 1 + 1 + len(taskName) + 1 // " " + spinner + " " + taskName + " "
	} else {
		leftContent = " " // Single space when not running
		visibleLen += 1
	}

	// Show Studio status if running
	if s.state.IsStudioRunning() {
		studioMsg := s.tr.StatusStudioOn
		leftContent += fmt.Sprintf("%s ", style.Green(studioMsg))
		visibleLen += len(studioMsg) + 1
	}

	// Helper to format key binding: [k]ey -> [Cyan(k)]Gray(ey)
	// Returns styled string and its visible length
	appendKey := func(key, desc string) {
		// Style: [key]desc
		styled := fmt.Sprintf("[%s]%s", style.Cyan(key), style.Gray(desc))
		// Visible: [key]desc
		vLen := 1 + len(key) + 1 + len(desc)

		leftContent += styled + " "
		visibleLen += vLen + 1
	}

	appendKey("r", s.tr.KeyHintRefresh)
	appendKey("d", s.tr.KeyHintDev)
	appendKey("D", s.tr.KeyHintDeploy)
	appendKey("g", s.tr.KeyHintGenerate)
	appendKey("s", s.tr.KeyHintResolve)
	appendKey("S", s.tr.KeyHintStudio)
	appendKey("c", s.tr.KeyHintCopy)

	// Right content (Metadata)
	styledRight := fmt.Sprintf("%s %s", style.Blue(s.config.Developer), style.Gray(s.config.Version))
	rightLen := len(s.config.Developer) + 1 + len(s.config.Version)

	// Calculate padding
	viewWidth, _ := v.Size()
	paddingLen := viewWidth - visibleLen - rightLen - 2 // -2 for extra safety buffer

	if paddingLen < 1 {
		paddingLen = 1
	}

	padding := ""
	for i := 0; i < paddingLen; i++ {
		padding += " "
	}

	fmt.Fprint(v, leftContent+padding+styledRight)

	return nil
}

// OnFocus is a no-op for the status bar (not focusable)
func (s *StatusBarContext) OnFocus() {}

// OnBlur is a no-op for the status bar (not focusable)
func (s *StatusBarContext) OnBlur() {}
