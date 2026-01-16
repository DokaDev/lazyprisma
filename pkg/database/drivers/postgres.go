package drivers

import (
	"database/sql"

	"github.com/dokadev/lazyprisma/pkg/database"

	_ "github.com/lib/pq" // PostgreSQL driver
)

const PostgresDriverName = "postgres"

// PostgresDriver implements DBDriver for PostgreSQL
type PostgresDriver struct {
	database.BaseDriver
}

// NewPostgres creates a new PostgreSQL driver
func NewPostgres() database.DBDriver {
	d := &PostgresDriver{}
	d.SetName(PostgresDriverName)
	return d
}

// Connect establishes a connection to PostgreSQL
func (d *PostgresDriver) Connect(cfg *database.Config) error {
	// Set default port for PostgreSQL
	if cfg.Port == 0 {
		cfg.Port = 5432
	}

	db, err := sql.Open("postgres", cfg.PostgresDSN())
	if err != nil {
		return err
	}

	d.SetDB(db)
	return d.Ping()
}

func init() {
	database.MustRegister(PostgresDriverName, NewPostgres)
}
