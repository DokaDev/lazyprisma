package prisma

import (
	"os"
	"sort"
	"strings"
)

type Migration struct {
	Name      string
	Timestamp string
}

type MigrationReader struct{}

func NewMigrationReader() *MigrationReader {
	return &MigrationReader{}
}

func (m *MigrationReader) GetMigrations() []Migration {
	migrationsPath := "prisma/migrations"

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return []Migration{}
	}

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return []Migration{}
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "migration_lock.toml" {
			name := entry.Name()

			if strings.HasPrefix(name, ".") {
				continue
			}

			parts := strings.SplitN(name, "_", 2)
			timestamp := ""
			displayName := name

			if len(parts) == 2 {
				timestamp = parts[0]
				displayName = parts[1]
			}

			migrations = append(migrations, Migration{
				Name:      displayName,
				Timestamp: timestamp,
			})
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Timestamp < migrations[j].Timestamp
	})

	return migrations
}
