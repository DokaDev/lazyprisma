package i18n

// NewTranslationSet returns a TranslationSet for the given language.
// Currently only English is supported; other languages will be added later.
func NewTranslationSet(language string) *TranslationSet {
	// For now, always return English.
	// Future: load JSON translations and merge with English defaults.
	return EnglishTranslationSet()
}
