package tui

import (
	"lazyprisma/pkg/version"

	"github.com/gdamore/tcell/v2"
)

func (a *App) drawHelp(x1, y1, x2, y2 int) {
	// Display different guides based on active panel
	var helpText string
	switch a.activePanelIdx {
	case 0: // Info panel
		helpText = " ←/→: Switch | ↑/↓: Scroll | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 1: // Migrations panel
		helpText = " ←/→: Switch | ↑/↓: Select | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 2: // DB Only panel
		helpText = " ←/→: Switch | ↑/↓: Select | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 3: // Migration Detail panel
		helpText = " ←/→: Switch | ↑/↓: Scroll | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	case 4: // Output panel
		helpText = " ←/→: Switch | ↑/↓: Scroll | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	default:
		helpText = " ←/→: Switch | r: Refresh | g: Generate | f: Format | t: Studio | d: Dev | D: Deploy | h: Help | q: Quit"
	}

	// Display key commands in purple, descriptions in white
	x := x1
	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	keyStyle := tcell.StyleDefault.Foreground(tcell.ColorPurple)

	for _, ch := range helpText {
		style := normalStyle

		// Detect key parts (alphabet, arrow symbols, slash)
		if ch == '←' || ch == '→' || ch == '↑' || ch == '↓' ||
		   (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '/' {
			style = keyStyle
		}

		if x >= x2 {
			break
		}
		a.screen.SetContent(x, y1, ch, nil, style)
		x++
	}
}

// drawVersionInfo draws the DokaLab branding and version in the bottom right corner
func (a *App) drawVersionInfo(width, height int) {
	// Build version string: "DokaLab v0.1 Beta"
	author := version.Author
	ver := version.Version
	versionText := author + " v" + ver

	// Calculate starting position (bottom right corner, same line as help)
	x := width - len(versionText) - 1
	y := height - 2

	// Style: DokaLab in cyan, version in gray
	authorStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkCyan).Bold(true)
	versionStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	// Draw author name
	currentX := x
	for _, ch := range author {
		a.screen.SetContent(currentX, y, ch, nil, authorStyle)
		currentX++
	}

	// Draw " v" and version
	versionPrefix := " v" + ver
	for _, ch := range versionPrefix {
		a.screen.SetContent(currentX, y, ch, nil, versionStyle)
		currentX++
	}
}
