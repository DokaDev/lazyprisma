package context

import (
	"github.com/jesseduffield/gocui"
)

const wheelScrollLines = 2

// ScrollableTrait provides shared vertical scroll logic.
// It tracks originY manually and applies it to the gocui view,
// replicating the exact behaviour used across all existing panels.
type ScrollableTrait struct {
	view    *gocui.View
	originY int
}

// SetView assigns (or reassigns) the underlying gocui view.
func (self *ScrollableTrait) SetView(v *gocui.View) {
	self.view = v
}

// GetOriginY returns the current scroll offset.
func (self *ScrollableTrait) GetOriginY() int {
	return self.originY
}

// SetOriginY sets the scroll offset directly (e.g. when restoring tab state).
func (self *ScrollableTrait) SetOriginY(y int) {
	self.originY = y
}

// ScrollUp scrolls the view up by 1 line.
func (self *ScrollableTrait) ScrollUp() {
	if self.originY > 0 {
		self.originY--
	}
}

// ScrollDown scrolls the view down by 1 line, clamping to the maximum scrollable position.
func (self *ScrollableTrait) ScrollDown() {
	if self.view == nil {
		return
	}

	maxOrigin := self.maxOrigin()
	if self.originY < maxOrigin {
		self.originY++
	}
}

// ScrollUpByWheel scrolls the view up by the wheel increment.
func (self *ScrollableTrait) ScrollUpByWheel() {
	if self.originY > 0 {
		self.originY -= wheelScrollLines
		if self.originY < 0 {
			self.originY = 0
		}
	}
}

// ScrollDownByWheel scrolls the view down by the wheel increment,
// clamping to the maximum scrollable position.
func (self *ScrollableTrait) ScrollDownByWheel() {
	if self.view == nil {
		return
	}

	maxOrigin := self.maxOrigin()
	if self.originY < maxOrigin {
		self.originY += wheelScrollLines
		if self.originY > maxOrigin {
			self.originY = maxOrigin
		}
	}
}

// ScrollToTop scrolls to the very top.
func (self *ScrollableTrait) ScrollToTop() {
	self.originY = 0
}

// ScrollToBottom scrolls to the very bottom.
func (self *ScrollableTrait) ScrollToBottom() {
	if self.view == nil {
		return
	}

	maxOrigin := self.maxOrigin()
	self.originY = maxOrigin
}

// AdjustScroll clamps originY to valid bounds and applies it to the view.
// Call this during render after content has been written to the view.
func (self *ScrollableTrait) AdjustScroll() {
	if self.view == nil {
		return
	}

	maxOrigin := self.maxOrigin()
	if self.originY > maxOrigin {
		self.originY = maxOrigin
	}
	if self.originY < 0 {
		self.originY = 0
	}

	self.view.SetOrigin(0, self.originY)
}

// maxOrigin calculates the maximum valid originY based on content and view size.
func (self *ScrollableTrait) maxOrigin() int {
	contentLines := len(self.view.ViewBufferLines())
	_, viewHeight := self.view.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	max := contentLines - innerHeight
	if max < 0 {
		max = 0
	}
	return max
}
