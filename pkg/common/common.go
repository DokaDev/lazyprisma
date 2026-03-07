package common

import (
	"github.com/dokadev/lazyprisma/pkg/i18n"
)

// Common provides shared dependencies used across the application.
// All components that need access to translations or configuration
// should receive a *Common reference.
type Common struct {
	Tr *i18n.TranslationSet
}

// NewCommon creates a new Common instance with the given TranslationSet.
func NewCommon(tr *i18n.TranslationSet) *Common {
	return &Common{
		Tr: tr,
	}
}
