package style

import (
	"fmt"
	"strings"
)

// Stylize applies combined ANSI styling (foreground colour code + bold flag).
// fgCode is a raw ANSI colour code such as "31" (red) or "38;5;208" (orange).
// If both fgCode and bold are empty/false the original text is returned unchanged.
func Stylize(text string, fgCode string, bold bool) string {
	if text == "" {
		return text
	}
	codes := make([]string, 0, 2)
	if fgCode != "" {
		codes = append(codes, fgCode)
	}
	if bold {
		codes = append(codes, "1")
	}
	if len(codes) == 0 {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", strings.Join(codes, ";"), text)
}

// ---------------------------------------------------------------------------
// Single-colour helpers
// ---------------------------------------------------------------------------

// Red colours text red (ANSI 31).
func Red(text string) string {
	return Stylize(text, "31", false)
}

// Green colours text green (ANSI 32).
func Green(text string) string {
	return Stylize(text, "32", false)
}

// Yellow colours text yellow (ANSI 33).
func Yellow(text string) string {
	return Stylize(text, "33", false)
}

// Blue colours text blue (ANSI 34).
func Blue(text string) string {
	return Stylize(text, "34", false)
}

// Magenta colours text magenta (ANSI 35).
func Magenta(text string) string {
	return Stylize(text, "35", false)
}

// Cyan colours text cyan (ANSI 36).
func Cyan(text string) string {
	return Stylize(text, "36", false)
}

// Orange colours text orange (256-colour ANSI 208).
func Orange(text string) string {
	return Stylize(text, "38;5;208", false)
}

// Gray colours text gray (256-colour ANSI 240).
func Gray(text string) string {
	return Stylize(text, "38;5;240", false)
}

// ---------------------------------------------------------------------------
// Compound helpers (colour + bold)
// ---------------------------------------------------------------------------

// RedBold colours text red and makes it bold.
func RedBold(text string) string {
	return Stylize(text, "31", true)
}

// GreenBold colours text green and makes it bold.
func GreenBold(text string) string {
	return Stylize(text, "32", true)
}

// YellowBold colours text yellow and makes it bold.
func YellowBold(text string) string {
	return Stylize(text, "33", true)
}

// CyanBold colours text cyan and makes it bold.
func CyanBold(text string) string {
	return Stylize(text, "36", true)
}

// OrangeBold colours text orange and makes it bold.
func OrangeBold(text string) string {
	return Stylize(text, "38;5;208", true)
}

// ---------------------------------------------------------------------------
// Attribute-only helpers
// ---------------------------------------------------------------------------

// Bold makes text bold (ANSI 1).
func Bold(text string) string {
	return Stylize(text, "", true)
}
