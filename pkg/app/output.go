package app

import (
	"fmt"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

type OutputPanel struct {
	BasePanel
	content              string
	subtitle             string // Custom subtitle
	originY              int    // Scroll position
	autoScrollToBottom   bool   // Auto-scroll to bottom on next draw
}

func NewOutputPanel(g *gocui.Gui) *OutputPanel {
	return &OutputPanel{
		BasePanel: NewBasePanel(ViewOutputs, g),
		content:   "", // Start with empty output
	}
}

func (o *OutputPanel) Draw(dim boxlayout.Dimensions) error {
	v, err := o.g.SetView(o.id, dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	o.SetupView(v, "Output")
	o.v = v
	v.Subtitle = o.subtitle // Set subtitle
	v.Wrap = true           // Enable word wrap
	fmt.Fprint(v, o.content)

	// Auto-scroll to bottom if flagged
	if o.autoScrollToBottom {
		// Calculate maxOrigin
		contentLines := len(v.ViewBufferLines())
		_, viewHeight := v.Size()
		innerHeight := viewHeight - 2 // Exclude frame
		maxOrigin := contentLines - innerHeight
		if maxOrigin < 0 {
			maxOrigin = 0
		}
		o.originY = maxOrigin
		o.autoScrollToBottom = false // Reset flag
	}

	// Adjust origin to ensure it's within valid bounds
	AdjustOrigin(v, &o.originY)
	v.SetOrigin(0, o.originY)

	return nil
}

func (o *OutputPanel) AppendOutput(text string) {
	o.content += text + "\n"
	// Flag to auto-scroll on next draw
	o.autoScrollToBottom = true
}

// LogAction logs an action with timestamp and optional details
func (o *OutputPanel) LogAction(action string, details ...string) {
	// Get current timestamp
	timestamp := time.Now().Format("15:04:05")

	// Add separator if there's already content
	if o.content != "" {
		o.content += "\n"
	}

	// Format: [Timestamp] Action (in cyan bold)
	header := fmt.Sprintf("%s %s", Gray(timestamp), Stylize(action, Style{FgColor: ColorCyan, Bold: true}))
	o.content += header + "\n"

	// Add details with indentation
	for _, detail := range details {
		o.content += "  " + detail + "\n"
	}

	// Flag to auto-scroll on next draw
	o.autoScrollToBottom = true
}

// SetSubtitle sets the custom subtitle for the panel
func (o *OutputPanel) SetSubtitle(subtitle string) {
	o.subtitle = subtitle
}

// LogActionRed logs an action in red (for errors/warnings)
func (o *OutputPanel) LogActionRed(action string, details ...string) {
	// Get current timestamp
	timestamp := time.Now().Format("15:04:05")

	// Add separator if there's already content
	if o.content != "" {
		o.content += "\n"
	}

	// Format: [Timestamp] Action in RED
	header := fmt.Sprintf("%s %s", Gray(timestamp),
		Stylize(action, Style{FgColor: ColorRed, Bold: true}))
	o.content += header + "\n"

	// Add details with indentation in red
	for _, detail := range details {
		o.content += "  " + Red(detail) + "\n"
	}

	// Flag to auto-scroll on next draw
	o.autoScrollToBottom = true
}

// ScrollUp scrolls the output panel up
func (o *OutputPanel) ScrollUp() {
	if o.originY > 0 {
		o.originY--
	}
}

// ScrollDown scrolls the output panel down
func (o *OutputPanel) ScrollDown() {
	o.originY++
	// AdjustOrigin will be called in Draw() to ensure bounds
}

// ScrollUpByWheel scrolls the output panel up by 2 lines (mouse wheel)
func (o *OutputPanel) ScrollUpByWheel() {
	if o.originY > 0 {
		o.originY -= 2
		if o.originY < 0 {
			o.originY = 0
		}
	}
}

// ScrollDownByWheel scrolls the output panel down by 2 lines (mouse wheel)
func (o *OutputPanel) ScrollDownByWheel() {
	if o.v == nil {
		return
	}

	// Get actual content lines from the rendered view buffer
	contentLines := len(o.v.ViewBufferLines())
	_, viewHeight := o.v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	// Only scroll if we haven't reached the bottom
	if o.originY < maxOrigin {
		o.originY += 2
		if o.originY > maxOrigin {
			o.originY = maxOrigin
		}
	}
}

// ScrollToTop scrolls to the top of the output panel
func (o *OutputPanel) ScrollToTop() {
	o.originY = 0
}

// ScrollToBottom scrolls to the bottom of the output panel
func (o *OutputPanel) ScrollToBottom() {
	if o.v == nil {
		return
	}

	// Get actual content lines from the rendered view buffer
	contentLines := len(o.v.ViewBufferLines())
	_, viewHeight := o.v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	o.originY = maxOrigin
}
