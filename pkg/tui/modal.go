package tui

import (
	"strings"

	"lazyprisma/pkg/env"

	"github.com/gdamore/tcell/v2"
)

func (a *App) handleModalKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		// Close modal
		a.showModal = false
		a.modalInput = ""
		a.modalType = ""
		a.modalCallback = nil
		a.modalSelectedButton = 0
		a.pendingMigrationName = ""
		a.helpScroll = 0
		a.draw()
	case tcell.KeyUp:
		// Scroll in help modal
		if a.modalType == "help" {
			a.helpScroll--
			if a.helpScroll < 0 {
				a.helpScroll = 0
			}
			a.draw()
		}
	case tcell.KeyDown:
		// Scroll in help modal
		if a.modalType == "help" {
			a.helpScroll++
			a.draw()
		}
	case tcell.KeyLeft:
		// Button selection in confirm_reset modal
		if a.modalType == "confirm_reset" {
			a.modalSelectedButton--
			if a.modalSelectedButton < 0 {
				a.modalSelectedButton = 0
			}
			a.draw()
		}
		// Button selection in migrate_resolve modal
		if a.modalType == "migrate_resolve" {
			a.modalSelectedButton--
			if a.modalSelectedButton < 0 {
				a.modalSelectedButton = 0
			}
			a.draw()
		}
	case tcell.KeyRight:
		// Button selection in confirm_reset modal
		if a.modalType == "confirm_reset" {
			a.modalSelectedButton++
			if a.modalSelectedButton > 1 {
				a.modalSelectedButton = 1
			}
			a.draw()
		}
		// Button selection in migrate_resolve modal
		if a.modalType == "migrate_resolve" {
			a.modalSelectedButton++
			if a.modalSelectedButton > 1 {
				a.modalSelectedButton = 1
			}
			a.draw()
		}
	case tcell.KeyEnter:
		// Just close if error display mode
		if a.modalType == "error" {
			a.showModal = false
			a.modalInput = ""
			a.modalType = ""
			a.draw()
			return
		}
		// Just close if help modal
		if a.modalType == "help" {
			a.showModal = false
			a.modalType = ""
			a.helpScroll = 0
			a.draw()
			return
		}
		// confirm_reset: Reset/Cancel selection
		if a.modalType == "confirm_reset" {
			if a.modalSelectedButton == 0 {
				// Reset selected
				a.showModal = false
				a.modalType = ""
				a.modalSelectedButton = 0
				go a.executeReset()
			} else {
				// Cancel selected
				a.showModal = false
				a.modalType = ""
				a.modalSelectedButton = 0
				a.pendingMigrationName = ""
			}
			a.draw()
			return
		}
		// confirm_dev: Switch to input modal after confirmation
		if a.modalType == "confirm_dev" {
			a.showModal = true
			a.modalTitle = "Migration Name"
			a.modalType = "input"
			a.modalInput = ""
			a.draw()
			return
		}
		// confirm_deploy: Execute immediately after confirmation
		if a.modalType == "confirm_deploy" {
			a.showModal = false
			a.modalType = ""
			go a.executeMigrateDeploy()
			return
		}
		// migrate_resolve: Execute migrate resolve with selected option
		if a.modalType == "migrate_resolve" {
			resolveType := "applied"
			if a.modalSelectedButton == 1 {
				resolveType = "rolled-back"
			}
			a.showModal = false
			a.modalType = ""
			a.modalSelectedButton = 0
			go a.executeMigrateResolve(a.pendingMigrationName, resolveType)
			a.pendingMigrationName = ""
			return
		}
		// input: Input complete
		if a.modalType == "input" && a.modalInput != "" {
			a.showModal = false
			a.modalType = ""
			go a.executeMigrateDev(a.modalInput)
			a.modalInput = ""
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Allow backspace only in input mode
		if a.modalType == "input" && len(a.modalInput) > 0 {
			a.modalInput = a.modalInput[:len(a.modalInput)-1]
			a.draw()
		}
	case tcell.KeyRune:
		// Allow input only in input mode
		if a.modalType == "input" {
			r := ev.Rune()
			if r == ' ' {
				r = '_'
			}
			a.modalInput += string(r)
			a.draw()
		}
	}
}

func (a *App) drawModal() {
	width, height := a.screen.Size()

	// Modal size and position
	modalWidth := 60
	modalHeight := 7

	// Adjust size for error, confirm_reset, help modals
	if a.modalType == "error" || a.modalType == "confirm_reset" {
		lines := strings.Split(a.modalInput, "\n")
		contentHeight := len(lines) + 6 // Include padding
		if a.modalType == "confirm_reset" {
			contentHeight += 2 // Button space
		}
		if contentHeight > modalHeight {
			modalHeight = contentHeight
		}
		if modalHeight > height-6 {
			modalHeight = height - 6
		}
	} else if a.modalType == "help" {
		// Help modal uses most of terminal size
		modalWidth = width - 10
		modalHeight = height - 6
		if modalWidth < 60 {
			modalWidth = 60
		}
		if modalHeight < 20 {
			modalHeight = 20
		}
	}

	if modalWidth > width-4 {
		modalWidth = width - 4
	}

	x1 := (width - modalWidth) / 2
	y1 := (height - modalHeight) / 2
	x2 := x1 + modalWidth
	y2 := y1 + modalHeight

	// Background (gray background instead of translucent effect)
	bgStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			a.screen.SetContent(x, y, ' ', nil, bgStyle)
		}
	}

	// Border
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
	// Top
	a.screen.SetContent(x1, y1, '╭', nil, boxStyle)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y1, '─', nil, boxStyle)
	}
	a.screen.SetContent(x2, y1, '╮', nil, boxStyle)

	// Left and right
	for y := y1 + 1; y < y2; y++ {
		a.screen.SetContent(x1, y, '│', nil, boxStyle)
		a.screen.SetContent(x2, y, '│', nil, boxStyle)
	}

	// Bottom
	a.screen.SetContent(x1, y2, '╰', nil, boxStyle)
	for x := x1 + 1; x < x2; x++ {
		a.screen.SetContent(x, y2, '─', nil, boxStyle)
	}
	a.screen.SetContent(x2, y2, '╯', nil, boxStyle)

	// Title
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack).Bold(true)
	titleText := " " + a.modalTitle + " "
	titleX := x1 + (modalWidth-len(titleText))/2
	for i, r := range titleText {
		a.screen.SetContent(titleX+i, y1, r, nil, titleStyle)
	}

	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	if a.modalType == "error" {
		// Display error message (multi-line support)
		lines := strings.Split(a.modalInput, "\n")
		msgY := y1 + 2

		for _, line := range lines {
			if msgY >= y2-2 {
				break // Stop if out of space
			}

			// Truncate long lines
			displayLine := line
			maxLineWidth := modalWidth - 4
			if len(displayLine) > maxLineWidth {
				displayLine = displayLine[:maxLineWidth]
			}

			// Display line (left-aligned)
			lineX := x1 + 2
			for _, r := range displayLine {
				if lineX >= x2-1 {
					break
				}
				a.screen.SetContent(lineX, msgY, r, nil, textStyle)
				lineX++
			}
			msgY++
		}

		// Help message
		hintText := "Press ESC or Enter to close"
		hintY := y2 - 1
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "confirm_reset" {
		// Display reset confirmation message
		lines := strings.Split(a.modalInput, "\n")
		msgY := y1 + 2

		for _, line := range lines {
			if msgY >= y2-4 {
				break
			}

			displayLine := line
			maxLineWidth := modalWidth - 4
			if len(displayLine) > maxLineWidth {
				displayLine = displayLine[:maxLineWidth]
			}

			lineX := x1 + 2
			for _, r := range displayLine {
				if lineX >= x2-1 {
					break
				}
				a.screen.SetContent(lineX, msgY, r, nil, textStyle)
				lineX++
			}
			msgY++
		}

		// Draw buttons
		buttonY := y2 - 2
		buttonSpacing := 4
		resetBtn := " Reset "
		cancelBtn := " Cancel "

		totalWidth := len(resetBtn) + buttonSpacing + len(cancelBtn)
		startX := x1 + (modalWidth-totalWidth)/2

		// Reset button
		resetStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		if a.modalSelectedButton == 0 {
			resetStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorRed).Bold(true)
		}
		for i, r := range resetBtn {
			a.screen.SetContent(startX+i, buttonY, r, nil, resetStyle)
		}

		// Cancel button
		cancelStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		if a.modalSelectedButton == 1 {
			cancelStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen).Bold(true)
		}
		cancelX := startX + len(resetBtn) + buttonSpacing
		for i, r := range cancelBtn {
			a.screen.SetContent(cancelX+i, buttonY, r, nil, cancelStyle)
		}

		// Help message
		hintText := "←/→: Select | Enter: Confirm | ESC: Cancel"
		hintY := y2 - 1
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "help" {
		// Display help content
		lines := strings.Split(a.modalInput, "\n")

		// Apply scroll
		visibleStart := a.helpScroll
		visibleEnd := len(lines)
		contentY := y1 + 2
		maxContentY := y2 - 2

		for i := visibleStart; i < visibleEnd && contentY < maxContentY; i++ {
			line := lines[i]

			// Truncate long lines
			maxLineWidth := modalWidth - 4
			if len(line) > maxLineWidth {
				line = line[:maxLineWidth]
			}

			// Draw line
			lineX := x1 + 2
			for _, r := range line {
				if lineX >= x2-1 {
					break
				}
				a.screen.SetContent(lineX, contentY, r, nil, textStyle)
				lineX++
			}
			contentY++
		}

		// Display scrollbar
		totalLines := len(lines)
		visibleLines := maxContentY - (y1 + 2)
		if totalLines > visibleLines {
			scrollbarX := x2 - 1
			scrollbarStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)

			scrollbarHeight := (visibleLines * visibleLines) / totalLines
			if scrollbarHeight < 1 {
				scrollbarHeight = 1
			}

			scrollbarPos := (a.helpScroll * visibleLines) / totalLines
			scrollbarY := y1 + 2 + scrollbarPos

			for i := 0; i < scrollbarHeight; i++ {
				if scrollbarY+i < maxContentY {
					a.screen.SetContent(scrollbarX, scrollbarY+i, '█', nil, scrollbarStyle)
				}
			}
		}

		// Help message
		hintText := "↑/↓: Scroll | ESC or Enter: Close"
		hintY := y2 - 1
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "confirm_dev" || a.modalType == "confirm_deploy" {
		// Display confirmation message
		envVarName := a.status.SchemaInfo.DatasourceEnvVar
		envReader := env.NewDotEnvReader(envVarName)
		maskedURL := envReader.MaskDatabaseURL(a.status.DatabaseURL)

		var msg1, msg2 string
		if a.modalType == "confirm_dev" {
			msg1 = "This will create a new migration and apply it to:"
			msg2 = "Press Enter to continue, ESC to cancel"
		} else {
			msg1 = "This will apply pending migrations to:"
			msg2 = "Press Enter to continue, ESC to cancel"
		}

		msgY := y1 + 2
		msgX := x1 + (modalWidth-len(msg1))/2
		for i, r := range msg1 {
			a.screen.SetContent(msgX+i, msgY, r, nil, textStyle)
		}

		// Display DB URL
		urlStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua).Background(tcell.ColorBlack)
		urlY := y1 + 3
		urlX := x1 + 2
		displayURL := maskedURL
		if len(displayURL) > modalWidth-4 {
			displayURL = displayURL[:modalWidth-4]
		}
		for i, r := range displayURL {
			a.screen.SetContent(urlX+i, urlY, r, nil, urlStyle)
		}

		// Help message
		hintY := y1 + 5
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(msg2))/2
		for i, r := range msg2 {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "migrate_resolve" {
		// Display message
		msg1 := "Select resolve status for migration:"
		msgY := y1 + 2
		msgX := x1 + (modalWidth-len(msg1))/2
		for i, r := range msg1 {
			a.screen.SetContent(msgX+i, msgY, r, nil, textStyle)
		}

		// Display migration name
		migrationStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
		migrationY := y1 + 3
		migrationX := x1 + 2
		displayName := a.pendingMigrationName
		if len(displayName) > modalWidth-4 {
			displayName = displayName[:modalWidth-4]
		}
		for i, r := range displayName {
			a.screen.SetContent(migrationX+i, migrationY, r, nil, migrationStyle)
		}

		// Draw buttons
		buttonY := y2 - 2
		buttonSpacing := 4
		appliedBtn := " Applied "
		rolledBackBtn := " Rolled-back "

		totalWidth := len(appliedBtn) + buttonSpacing + len(rolledBackBtn)
		startX := x1 + (modalWidth-totalWidth)/2

		// Applied button
		appliedStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		if a.modalSelectedButton == 0 {
			appliedStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen).Bold(true)
		}
		for i, r := range appliedBtn {
			a.screen.SetContent(startX+i, buttonY, r, nil, appliedStyle)
		}

		// Rolled-back button
		rolledBackStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		if a.modalSelectedButton == 1 {
			rolledBackStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow).Bold(true)
		}
		rolledBackX := startX + len(appliedBtn) + buttonSpacing
		for i, r := range rolledBackBtn {
			a.screen.SetContent(rolledBackX+i, buttonY, r, nil, rolledBackStyle)
		}

		// Help message
		hintText := "←/→: Select | Enter: Confirm | ESC: Cancel"
		hintY := y2 - 1
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	} else if a.modalType == "input" {
		// Input field label
		labelText := "Enter migration name:"
		labelY := y1 + 2
		for i, r := range labelText {
			a.screen.SetContent(x1+2+i, labelY, r, nil, textStyle)
		}

		// Input field
		inputY := y1 + 3
		inputStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
		inputFieldWidth := modalWidth - 4
		for x := 0; x < inputFieldWidth; x++ {
			a.screen.SetContent(x1+2+x, inputY, ' ', nil, inputStyle)
		}

		// Display input content
		for i, r := range a.modalInput {
			if i < inputFieldWidth {
				a.screen.SetContent(x1+2+i, inputY, r, nil, inputStyle)
			}
		}

		// Display cursor
		cursorX := x1 + 2 + len(a.modalInput)
		if cursorX < x2-1 {
			a.screen.SetContent(cursorX, inputY, '▏', nil, inputStyle.Foreground(tcell.ColorYellow))
		}

		// Help message
		hintStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		hintText := "Press Enter to confirm, Esc to cancel"
		hintY := y1 + 5
		hintX := x1 + (modalWidth-len(hintText))/2
		for i, r := range hintText {
			a.screen.SetContent(hintX+i, hintY, r, nil, hintStyle)
		}
	}
}
