package app

func (a *App) FocusNext() {
	// Ignore if modal is active
	if a.HasActiveModal() {
		return
	}

	if len(a.focusOrder) == 0 {
		return
	}

	// Blur current
	if panel, ok := a.panels[a.focusOrder[a.currentFocus]]; ok {
		panel.OnBlur()
	}

	// Next (circular)
	a.currentFocus = (a.currentFocus + 1) % len(a.focusOrder)

	// Focus new
	if panel, ok := a.panels[a.focusOrder[a.currentFocus]]; ok {
		panel.OnFocus()
	}
}

func (a *App) FocusPrevious() {
	// Ignore if modal is active
	if a.HasActiveModal() {
		return
	}

	if len(a.focusOrder) == 0 {
		return
	}

	// Blur current
	if panel, ok := a.panels[a.focusOrder[a.currentFocus]]; ok {
		panel.OnBlur()
	}

	// Previous (circular)
	a.currentFocus = (a.currentFocus - 1 + len(a.focusOrder)) % len(a.focusOrder)

	// Focus new
	if panel, ok := a.panels[a.focusOrder[a.currentFocus]]; ok {
		panel.OnFocus()
	}
}
