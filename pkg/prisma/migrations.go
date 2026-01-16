package prisma

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	MigrationsDirName = "migrations"
)

// Migration represents a migration
type Migration struct {
	Name             string
	Path             string
	AppliedAt        *time.Time // nil if not applied to DB
	IsEmpty          bool       // true if migration folder is empty or has no migration.sql
	HasDownSQL       bool       // true if migration folder has down.sql
	Checksum         string     // SHA-256 checksum of local migration.sql
	DBChecksum       string     // SHA-256 checksum from DB (if exists)
	IsFailed         bool       // true if migration failed (finished_at IS NULL AND rolled_back_at IS NULL)
	ChecksumMismatch bool       // true if local checksum != DB checksum
	Logs             *string    // Migration logs from DB (if failed)
	StartedAt        *time.Time // Migration start time from DB (for in-transaction migrations)
}

// GetLocalMigrations returns a list of local migrations from the prisma/migrations directory
func GetLocalMigrations(projectDir string) ([]Migration, error) {
	// Build migrations directory path
	migrationsPath := filepath.Join(projectDir, SchemaDirName, MigrationsDirName)

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return []Migration{}, nil // Return empty list if no migrations directory
	}

	// Read migrations directory
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	// Collect migration directories
	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			migrationPath := filepath.Join(migrationsPath, entry.Name())

			// Check if migration.sql exists
			sqlFile := filepath.Join(migrationPath, "migration.sql")
			isEmpty := true
			checksum := ""
			if stat, err := os.Stat(sqlFile); err == nil && !stat.IsDir() {
				// Check if file has content
				if stat.Size() > 0 {
					isEmpty = false
					// Calculate checksum for non-empty migrations
					if cs, err := calculateChecksum(migrationPath); err == nil {
						checksum = cs
					}
				}
			}

			// Check if down.sql exists
			downSQLFile := filepath.Join(migrationPath, "down.sql")
			hasDownSQL := false
			if stat, err := os.Stat(downSQLFile); err == nil && !stat.IsDir() && stat.Size() > 0 {
				hasDownSQL = true
			}

			migrations = append(migrations, Migration{
				Name:       entry.Name(),
				Path:       migrationPath,
				IsEmpty:    isEmpty,
				HasDownSQL: hasDownSQL,
				Checksum:   checksum,
			})
		}
	}

	// Sort migrations by name (chronological order due to timestamp prefix)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

// calculateChecksum computes SHA-256 checksum of migration.sql file
func calculateChecksum(migrationPath string) (string, error) {
	sqlFile := filepath.Join(migrationPath, "migration.sql")

	content, err := os.ReadFile(sqlFile)
	if err != nil {
		return "", err
	}

	// Normalize line endings: replace CRLF with LF
	normalized := make([]byte, 0, len(content))
	for i := 0; i < len(content); i++ {
		if content[i] == '\r' {
			if i+1 < len(content) && content[i+1] == '\n' {
				continue // Skip CR if followed by LF
			}
		}
		normalized = append(normalized, content[i])
	}

	hash := sha256.Sum256(normalized)
	return hex.EncodeToString(hash[:]), nil
}

// DBMigration represents a migration from the database
type DBMigration struct {
	Name         string
	Checksum     string
	StartedAt    *time.Time // Migration start time
	FinishedAt   *time.Time // nil if migration failed
	RolledBackAt *time.Time // nil if not rolled back
	Logs         *string    // Migration logs (especially for failed migrations)
}

// GetDBMigrations returns migrations from the _prisma_migrations table
func GetDBMigrations(db *sql.DB) ([]DBMigration, error) {
	query := `
		SELECT migration_name, checksum, started_at, finished_at, rolled_back_at, logs
		FROM _prisma_migrations
		ORDER BY started_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []DBMigration
	for rows.Next() {
		var m DBMigration
		if err := rows.Scan(&m.Name, &m.Checksum, &m.StartedAt, &m.FinishedAt, &m.RolledBackAt, &m.Logs); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, rows.Err()
}

// MigrationCategory represents different migration types
type MigrationCategory struct {
	Local   []Migration // All local migrations
	Pending []Migration // Local but not in DB
	DBOnly  []Migration // In DB but not local
}

// CompareMigrations classifies migrations into categories
func CompareMigrations(localMigrations []Migration, dbMigrations []DBMigration) MigrationCategory {
	category := MigrationCategory{
		Local:   make([]Migration, 0),
		Pending: make([]Migration, 0),
		DBOnly:  make([]Migration, 0),
	}

	// Create map for quick lookup
	dbMap := make(map[string]DBMigration)
	for _, dbMig := range dbMigrations {
		dbMap[dbMig.Name] = dbMig
	}

	localMap := make(map[string]Migration)
	for _, localMig := range localMigrations {
		localMap[localMig.Name] = localMig
	}

	// Classify local migrations
	for _, localMig := range localMigrations {
		mig := localMig

		if dbMig, exists := dbMap[localMig.Name]; exists {
			// Migration exists in DB

			// Store DB checksum
			mig.DBChecksum = dbMig.Checksum

			// Check if failed (finished_at IS NULL AND rolled_back_at IS NULL)
			if dbMig.FinishedAt == nil && dbMig.RolledBackAt == nil {
				mig.IsFailed = true
				mig.Logs = dbMig.Logs
				mig.StartedAt = dbMig.StartedAt
			} else if dbMig.FinishedAt != nil {
				// Migration is applied
				mig.AppliedAt = dbMig.FinishedAt
			}

			// Check checksum mismatch (only for successfully applied migrations)
			if !mig.IsEmpty && mig.AppliedAt != nil && mig.Checksum != "" && dbMig.Checksum != "" {
				if mig.Checksum != dbMig.Checksum {
					mig.ChecksumMismatch = true
				}
			}

			category.Local = append(category.Local, mig)
		} else {
			// Migration is pending (not in DB)
			category.Pending = append(category.Pending, mig)
			category.Local = append(category.Local, mig)
		}
	}

	// Find DB-only migrations
	for _, dbMig := range dbMigrations {
		if _, exists := localMap[dbMig.Name]; !exists {
			mig := Migration{
				Name:       dbMig.Name,
				Path:       "",
				DBChecksum: dbMig.Checksum,
			}

			// Check if failed
			if dbMig.FinishedAt == nil && dbMig.RolledBackAt == nil {
				mig.IsFailed = true
				mig.Logs = dbMig.Logs
				mig.StartedAt = dbMig.StartedAt
			} else if dbMig.FinishedAt != nil {
				mig.AppliedAt = dbMig.FinishedAt
			}

			category.DBOnly = append(category.DBOnly, mig)
		}
	}

	return category
}
