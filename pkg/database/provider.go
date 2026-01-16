package database

import "strings"

// ProviderDisplayNames maps provider names to their display names
var ProviderDisplayNames = map[string]string{
	"postgresql": "PostgreSQL",
	"postgres":   "PostgreSQL",
	"mysql":      "MySQL",
	"sqlite":     "SQLite",
	"sqlserver":  "SQL Server",
	"mongodb":    "MongoDB",
	"cockroachdb": "CockroachDB",
}

// GetProviderDisplayName returns the formatted display name for a provider
func GetProviderDisplayName(provider string) string {
	// Handle empty provider
	if provider == "" {
		return "Not specified"
	}

	// Check if known provider
	if displayName, ok := ProviderDisplayNames[strings.ToLower(provider)]; ok {
		return displayName
	}

	// Unknown provider
	return "Unknown"
}
