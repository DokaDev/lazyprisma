package i18n

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"dario.cat/mergo"
)

//go:embed translations/*.json
var translationsFS embed.FS

// NewTranslationSet returns a TranslationSet for the given language.
// If language is "auto", it detects the system language.
// Falls back to English if the language is not supported.
func NewTranslationSet(language string) *TranslationSet {
	if language == "auto" || language == "" {
		language = detectSystemLanguage()
	}

	if language == "en" {
		return EnglishTranslationSet()
	}

	base := EnglishTranslationSet()
	overlay, err := loadLanguageJSON(language)
	if err != nil {
		fmt.Printf("warning: failed to load translations for %q: %v\n", language, err)
		return base
	}

	if err := mergo.Merge(base, &overlay, mergo.WithOverride); err != nil {
		fmt.Printf("warning: failed to merge translations for %q: %v\n", language, err)
		return EnglishTranslationSet()
	}

	return base
}

// loadLanguageJSON reads a translation JSON file from the embedded filesystem.
// If the file does not exist, it returns an empty TranslationSet with nil error.
func loadLanguageJSON(language string) (TranslationSet, error) {
	filename := fmt.Sprintf("translations/%s.json", language)
	data, err := translationsFS.ReadFile(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return TranslationSet{}, nil
		}
		return TranslationSet{}, fmt.Errorf("reading %s: %w", filename, err)
	}

	var ts TranslationSet
	if err := json.Unmarshal(data, &ts); err != nil {
		return TranslationSet{}, fmt.Errorf("invalid JSON in %s: %w", filename, err)
	}

	return ts, nil
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
