package app

import (
	"fmt"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

type StatusBar struct {
	BasePanel
	app *App // Reference to App for accessing command state
}

func NewStatusBar(g *gocui.Gui, app *App) *StatusBar {
	return &StatusBar{
		BasePanel: NewBasePanel(ViewStatusbar, g),
		app:       app,
	}
}

func (s *StatusBar) Draw(dim boxlayout.Dimensions) error {
	// StatusBar has no frame, so adjust dimensions
	frameOffset := 1
	x0 := dim.X0 - frameOffset
	y0 := dim.Y0 - frameOffset
	x1 := dim.X1 + frameOffset
	y1 := dim.Y1 + frameOffset

	v, err := s.g.SetView(s.id, x0, y0, x1, y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	s.v = v
	v.Clear()
	v.Frame = false

	// Build status bar content
	var leftContent string
	var visibleLen int

	// Show spinner if command is running
	if s.app.commandRunning.Load() {
		frameIndex := s.app.spinnerFrame.Load()
		spinner := string(spinnerFrames[frameIndex])

		// Get running task name
		taskName := ""
		if val := s.app.runningCommandName.Load(); val != nil {
			taskName = val.(string)
		}

		leftContent = fmt.Sprintf(" %s %s ", Cyan(spinner), Gray(taskName))
		visibleLen += 1 + 1 + 1 + len(taskName) + 1 // " " + spinner + " " + taskName + " "
	} else {
		leftContent = " " // Single space when not running
		visibleLen += 1
	}

	// Show Studio status if running
	if s.app.studioRunning {
		studioMsg := "[Studio: ON]"
		leftContent += fmt.Sprintf("%s ", Green(studioMsg))
		visibleLen += len(studioMsg) + 1
	}

	// Helper to format key binding: [k]ey -> [Cyan(k)]Gray(ey)
	// Returns styled string and its visible length
	appendKey := func(key, desc string) {
		// Style: [key]desc
		styled := fmt.Sprintf("[%s]%s", Cyan(key), Gray(desc))
		// Visible: [key]desc
		vLen := 1 + len(key) + 1 + len(desc)
		
		leftContent += styled + " "
		visibleLen += vLen + 1
	}

	appendKey("r", "efresh")
	appendKey("d", "ev")
	appendKey("D", "eploy")
	appendKey("g", "enerate")
	appendKey("s", "resolve")
	appendKey("S", "tudio")
	appendKey("c", "opy")

	// Right content (Metadata)
	// Style right content (e.g., in blue or default)
	styledRight := fmt.Sprintf("%s %s", Blue(s.app.config.Developer), Gray(s.app.config.Version))
	rightLen := len(s.app.config.Developer) + 1 + len(s.app.config.Version)

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

	fmt.Fprint(v, leftContent + padding + styledRight)

	return nil
}

// 상태바는 포커스를 받지 않음
func (s *StatusBar) OnFocus() {}
func (s *StatusBar) OnBlur()  {}
