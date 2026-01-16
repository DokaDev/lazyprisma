package prisma

import (
	"os"
	"path/filepath"
)

const (
	// v7.0+ config file
	ConfigFileName = "prisma.config.ts"

	// v7.0 이전 schema file
	SchemaFileName = "schema.prisma"
	SchemaDirName  = "prisma"
)

// IsWorkspace checks if the given directory is a Prisma workspace
// For v7.0+: checks for prisma.config.ts in the current directory
// For v7.0-: checks for prisma/schema.prisma in the current directory
func IsWorkspace(dir string) bool {
	// Check for v7.0+ workspace (prisma.config.ts)
	configPath := filepath.Join(dir, ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	// Check for v7.0- workspace (prisma/schema.prisma)
	schemaPath := filepath.Join(dir, SchemaDirName, SchemaFileName)
	if _, err := os.Stat(schemaPath); err == nil {
		return true
	}

	return false
}

// GetWorkspaceType returns the type of Prisma workspace
// Returns "v7+" for prisma.config.ts, "v7-" for prisma/schema.prisma, "" if not a workspace
func GetWorkspaceType(dir string) string {
	// Check for v7.0+
	configPath := filepath.Join(dir, ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return "v7+"
	}

	// Check for v7.0-
	schemaPath := filepath.Join(dir, SchemaDirName, SchemaFileName)
	if _, err := os.Stat(schemaPath); err == nil {
		return "v7-"
	}

	return ""
}
