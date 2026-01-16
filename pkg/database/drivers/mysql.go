package drivers

import (
	"database/sql"

	"github.com/dokadev/lazyprisma/pkg/database"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

const MySQLDriverName = "mysql"

// MySQLDriver implements DBDriver for MySQL/MariaDB
type MySQLDriver struct {
	database.BaseDriver
}

// NewMySQL creates a new MySQL driver
func NewMySQL() database.DBDriver {
	d := &MySQLDriver{}
	d.SetName(MySQLDriverName)
	return d
}

// Connect establishes a connection to MySQL
func (d *MySQLDriver) Connect(cfg *database.Config) error {
	// Set default port for MySQL
	if cfg.Port == 0 {
		cfg.Port = 3306
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN())
	if err != nil {
		return err
	}

	d.SetDB(db)
	return d.Ping()
}

func init() {
	database.MustRegister(MySQLDriverName, NewMySQL)
}
