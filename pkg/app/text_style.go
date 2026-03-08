package app

// ============================================================================
// Color enum — used by modal styling (MessageModalStyle, ColorToGocuiAttr)
// ============================================================================

// Color represents terminal colors
type Color int

const (
	ColorDefault Color = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)
