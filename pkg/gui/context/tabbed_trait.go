package context

// TabbedTrait provides shared tab management with per-tab scroll position saving.
// It replicates the exact tab origin save/restore pattern used in
// MigrationsPanel and DetailsPanel.
type TabbedTrait struct {
	tabs          []string
	currentTabIdx int
	tabOriginY    map[int]int // scroll position keyed by tab index
}

// NewTabbedTrait creates a TabbedTrait with the given initial tabs.
func NewTabbedTrait(tabs []string) TabbedTrait {
	return TabbedTrait{
		tabs:          tabs,
		currentTabIdx: 0,
		tabOriginY:    make(map[int]int),
	}
}

// SetTabs replaces the tab list. If the current index is out of bounds
// after the change, it resets to 0.
func (self *TabbedTrait) SetTabs(tabs []string) {
	self.tabs = tabs
	if self.currentTabIdx >= len(self.tabs) {
		self.currentTabIdx = 0
	}
}

// GetTabs returns the current tab names.
func (self *TabbedTrait) GetTabs() []string {
	return self.tabs
}

// GetCurrentTab returns the name of the active tab,
// or an empty string if there are no tabs.
func (self *TabbedTrait) GetCurrentTab() string {
	if self.currentTabIdx >= len(self.tabs) {
		return ""
	}
	return self.tabs[self.currentTabIdx]
}

// GetCurrentTabIdx returns the zero-based index of the active tab.
func (self *TabbedTrait) GetCurrentTabIdx() int {
	return self.currentTabIdx
}

// SetCurrentTabIdx sets the active tab index directly.
func (self *TabbedTrait) SetCurrentTabIdx(idx int) {
	if idx >= 0 && idx < len(self.tabs) {
		self.currentTabIdx = idx
	}
}

// NextTab advances to the next tab, wrapping around.
// The caller must call SaveTabOriginY before and RestoreTabOriginY after.
func (self *TabbedTrait) NextTab() {
	if len(self.tabs) == 0 {
		return
	}
	self.currentTabIdx = (self.currentTabIdx + 1) % len(self.tabs)
}

// PrevTab moves to the previous tab, wrapping around.
// The caller must call SaveTabOriginY before and RestoreTabOriginY after.
func (self *TabbedTrait) PrevTab() {
	if len(self.tabs) == 0 {
		return
	}
	self.currentTabIdx = (self.currentTabIdx - 1 + len(self.tabs)) % len(self.tabs)
}

// SaveTabOriginY saves the given scroll position for the current tab.
// Call this before switching tabs.
func (self *TabbedTrait) SaveTabOriginY(originY int) {
	self.tabOriginY[self.currentTabIdx] = originY
}

// RestoreTabOriginY returns the saved scroll position for the current tab.
// Returns 0 if no position was previously saved.
// Call this after switching tabs.
func (self *TabbedTrait) RestoreTabOriginY() int {
	if y, exists := self.tabOriginY[self.currentTabIdx]; exists {
		return y
	}
	return 0
}
