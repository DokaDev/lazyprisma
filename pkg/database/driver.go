package database

import "database/sql"

// DBDriver defines the interface for database drivers (Strategy pattern)
type DBDriver interface {
	// Name returns the driver name (e.g., "postgres", "mysql")
	Name() string

	// Connect establishes a connection using the provided config
	Connect(cfg *Config) error

	// Close closes the database connection
	Close() error

	// Ping verifies the connection is alive
	Ping() error

	// Query executes a query that returns rows
	Query(query string, args ...any) (*sql.Rows, error)

	// QueryRow executes a query that returns at most one row
	QueryRow(query string, args ...any) *sql.Row

	// Exec executes a query without returning rows
	Exec(query string, args ...any) (sql.Result, error)

	// DB returns the underlying *sql.DB for advanced operations
	DB() *sql.DB
}

// BaseDriver provides common functionality for SQL-based drivers
type BaseDriver struct {
	db   *sql.DB
	name string
}

// Name returns the driver name
func (d *BaseDriver) Name() string {
	return d.name
}

// Connect is a no-op for BaseDriver (used when connection is already established)
func (d *BaseDriver) Connect(cfg *Config) error {
	return nil
}

// Close closes the database connection
func (d *BaseDriver) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Ping verifies the connection is alive
func (d *BaseDriver) Ping() error {
	if d.db == nil {
		return ErrNotConnected
	}
	return d.db.Ping()
}

// Query executes a query that returns rows
func (d *BaseDriver) Query(query string, args ...any) (*sql.Rows, error) {
	if d.db == nil {
		return nil, ErrNotConnected
	}
	return d.db.Query(query, args...)
}

// QueryRow executes a query that returns at most one row
func (d *BaseDriver) QueryRow(query string, args ...any) *sql.Row {
	if d.db == nil {
		return nil
	}
	return d.db.QueryRow(query, args...)
}

// Exec executes a query without returning rows
func (d *BaseDriver) Exec(query string, args ...any) (sql.Result, error) {
	if d.db == nil {
		return nil, ErrNotConnected
	}
	return d.db.Exec(query, args...)
}

// DB returns the underlying *sql.DB
func (d *BaseDriver) DB() *sql.DB {
	return d.db
}

// SetDB sets the underlying database connection
func (d *BaseDriver) SetDB(db *sql.DB) {
	d.db = db
}

// SetName sets the driver name
func (d *BaseDriver) SetName(name string) {
	d.name = name
}
