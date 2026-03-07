package i18n

import (
	"os"
	"strings"
)

// Supported language codes
var supportedLanguages = map[string]func() *TranslationSet{
	"en": EnglishTranslationSet,
}

// NewTranslationSet returns a TranslationSet for the given language.
// If language is "auto", it detects the system language.
// Falls back to English if the language is not supported.
func NewTranslationSet(language string) *TranslationSet {
	if language == "auto" || language == "" {
		language = detectSystemLanguage()
	}

	if factory, ok := supportedLanguages[language]; ok {
		return factory()
	}

	return EnglishTranslationSet()
}

// detectSystemLanguage checks LANG, LC_ALL, LC_MESSAGES environment variables.
func detectSystemLanguage() string {
	for _, envVar := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if val := os.Getenv(envVar); val != "" {
			return parseLanguageCode(val)
		}
	}
	return "en"
}

// parseLanguageCode extracts the language code from locale strings like "ko_KR.UTF-8".
func parseLanguageCode(locale string) string {
	// Remove encoding (e.g., ".UTF-8")
	if idx := strings.Index(locale, "."); idx != -1 {
		locale = locale[:idx]
	}
	// Remove country (e.g., "_KR")
	if idx := strings.Index(locale, "_"); idx != -1 {
		locale = locale[:idx]
	}
	// Remove region variant (e.g., "-KR")
	if idx := strings.Index(locale, "-"); idx != -1 {
		locale = locale[:idx]
	}
	return strings.ToLower(locale)
}
