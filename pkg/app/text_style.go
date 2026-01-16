package app

import "fmt"

// ============================================================================
// Text Styling Utilities - ANSI Escape Code Helpers
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

// colorToFgCode converts Color to ANSI foreground code
func colorToFgCode(c Color) string {
	codes := map[Color]string{
		ColorDefault: "",
		ColorBlack:   "30",
		ColorRed:     "31",
		ColorGreen:   "32",
		ColorYellow:  "33",
		ColorBlue:    "34",
		ColorMagenta: "35",
		ColorCyan:    "36",
		ColorWhite:   "37",
	}
	return codes[c]
}

// colorToBgCode converts Color to ANSI background code
func colorToBgCode(c Color) string {
	codes := map[Color]string{
		ColorDefault: "",
		ColorBlack:   "40",
		ColorRed:     "41",
		ColorGreen:   "42",
		ColorYellow:  "43",
		ColorBlue:    "44",
		ColorMagenta: "45",
		ColorCyan:    "46",
		ColorWhite:   "47",
	}
	return codes[c]
}

// Style represents text styling options
type Style struct {
	FgColor    Color
	BgColor    Color
	Bold       bool
	Italic     bool
	Underline  bool
	Dim        bool
	Blink      bool
	Reverse    bool
	StrikeThru bool
}

// Stylize applies the given style to text using ANSI escape codes
func Stylize(text string, style Style) string {
	if text == "" {
		return text
	}

	codes := make([]string, 0, 5)

	// Foreground color
	if fgCode := colorToFgCode(style.FgColor); fgCode != "" {
		codes = append(codes, fgCode)
	}

	// Background color
	if bgCode := colorToBgCode(style.BgColor); bgCode != "" {
		codes = append(codes, bgCode)
	}

	// Attributes
	if style.Bold {
		codes = append(codes, "1")
	}
	if style.Dim {
		codes = append(codes, "2")
	}
	if style.Italic {
		codes = append(codes, "3")
	}
	if style.Underline {
		codes = append(codes, "4")
	}
	if style.Blink {
		codes = append(codes, "5")
	}
	if style.Reverse {
		codes = append(codes, "7")
	}
	if style.StrikeThru {
		codes = append(codes, "9")
	}

	// No styling needed
	if len(codes) == 0 {
		return text
	}

	// Build ANSI escape sequence
	var escape string
	for i, code := range codes {
		if i == 0 {
			escape = code
		} else {
			escape += ";" + code
		}
	}

	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", escape, text)
}

// ============================================================================
// Convenience Functions - Color Only
// ============================================================================

// Colorize applies a foreground color to text
func Colorize(text string, color Color) string {
	return Stylize(text, Style{FgColor: color})
}

// Red colors text red
func Red(text string) string {
	return Colorize(text, ColorRed)
}

// Green colors text green
func Green(text string) string {
	return Colorize(text, ColorGreen)
}

// Yellow colors text yellow
func Yellow(text string) string {
	return Colorize(text, ColorYellow)
}

// Blue colors text blue
func Blue(text string) string {
	return Colorize(text, ColorBlue)
}

// Magenta colors text magenta
func Magenta(text string) string {
	return Colorize(text, ColorMagenta)
}

// Cyan colors text cyan
func Cyan(text string) string {
	return Colorize(text, ColorCyan)
}

// White colors text white
func White(text string) string {
	return Colorize(text, ColorWhite)
}

// Black colors text black
func Black(text string) string {
	return Colorize(text, ColorBlack)
}

// Orange colors text orange (using 256-color ANSI code)
func Orange(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[38;5;208m%s\x1b[0m", text)
}

// Gray colors text gray (using 256-color ANSI code)
func Gray(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[38;5;240m%s\x1b[0m", text)
}

// ============================================================================
// Convenience Functions - Attributes Only
// ============================================================================

// Bold makes text bold
func Bold(text string) string {
	return Stylize(text, Style{Bold: true})
}

// Italic makes text italic
func Italic(text string) string {
	return Stylize(text, Style{Italic: true})
}

// Underline underlines text
func Underline(text string) string {
	return Stylize(text, Style{Underline: true})
}

// Dim makes text dim/faint
func Dim(text string) string {
	return Stylize(text, Style{Dim: true})
}

// Blink makes text blink
func Blink(text string) string {
	return Stylize(text, Style{Blink: true})
}

// Reverse reverses foreground and background colors
func Reverse(text string) string {
	return Stylize(text, Style{Reverse: true})
}

// StrikeThru strikes through text
func StrikeThru(text string) string {
	return Stylize(text, Style{StrikeThru: true})
}
