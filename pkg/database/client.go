package database

import (
	"database/sql"
	"net/url"
	"strings"
)

// Client is a high-level wrapper around DBDriver
type Client struct {
	driver     DBDriver
	config     *Config
	driverName string // store for DSN connections
}

// NewClient creates a new database client for the specified driver
func NewClient(driverName string) (*Client, error) {
	driver, err := Get(driverName)
	if err != nil {
		return nil, err
	}

	return &Client{
		driver:     driver,
		driverName: driverName,
	}, nil
}

// sqlDriverName maps Prisma provider names to Go sql driver names
var sqlDriverName = map[string]string{
	"postgresql": "postgres",
	"postgres":   "postgres",
	"mysql":      "mysql",
	"sqlite":     "sqlite3",
	"sqlserver":  "sqlserver",
	"cockroachdb": "postgres", // CockroachDB uses postgres protocol
}

// NewClientFromDSN creates a new client and connects using a DSN string
// driverName should be the Prisma provider name (postgresql, mysql, etc.)
func NewClientFromDSN(driverName, dsn string) (*Client, error) {
	// Map Prisma driver name to Go sql driver name
	sqlDriver, ok := sqlDriverName[driverName]
	if !ok {
		sqlDriver = driverName // fallback to same name
	}

	// Convert Prisma URL to Go driver DSN format
	goDSN, err := convertPrismaURLToGoDSN(driverName, dsn)
	if err != nil {
		return nil, err
	}

	// Open connection directly
	db, err := sql.Open(sqlDriver, goDSN)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Create a base driver to wrap the connection
	baseDriver := &BaseDriver{}
	baseDriver.SetDB(db)
	baseDriver.SetName(driverName)

	return &Client{
		driver:     baseDriver,
		driverName: driverName,
	}, nil
}

// convertPrismaURLToGoDSN converts Prisma database URL to Go driver DSN format
func convertPrismaURLToGoDSN(driverName, prismaURL string) (string, error) {
	switch driverName {
	case "mysql":
		return convertMySQLURL(prismaURL)
	case "postgresql", "postgres", "cockroachdb":
		return convertPostgresURL(prismaURL)
	case "sqlite":
		// SQLite: remove file:// prefix if present
		return strings.TrimPrefix(prismaURL, "file://"), nil
	default:
		return prismaURL, nil
	}
}

// convertPostgresURL converts Prisma PostgreSQL URL to lib/pq compatible format
func convertPostgresURL(prismaURL string) (string, error) {
	// lib/pq doesn't support:
	// - sslmode=prefer (Prisma default) - only disable, require, verify-ca, verify-full
	// - schema parameter (Prisma extension for namespace)

	// Parse URL to manipulate query parameters
	u, err := url.Parse(prismaURL)
	if err != nil {
		// If parsing fails, try simple string replacement
		return convertPostgresURLSimple(prismaURL), nil
	}

	// Get query parameters
	query := u.Query()

	// Remove unsupported "schema" parameter
	query.Del("schema")

	// Handle sslmode
	sslmode := query.Get("sslmode")
	if sslmode == "prefer" || sslmode == "" {
		// Replace "prefer" with "disable" for local development
		// Set "disable" if not specified
		query.Set("sslmode", "disable")
	}

	// Rebuild URL with modified parameters
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// convertPostgresURLSimple is a fallback for simple string-based conversion
func convertPostgresURLSimple(prismaURL string) string {
	result := prismaURL

	// Remove schema parameter (with various separators)
	result = strings.ReplaceAll(result, "&schema=public", "")
	result = strings.ReplaceAll(result, "?schema=public&", "?")
	result = strings.ReplaceAll(result, "?schema=public", "")

	// Replace prefer with disable
	result = strings.ReplaceAll(result, "sslmode=prefer", "sslmode=disable")

	// Add sslmode if not present
	if !strings.Contains(result, "sslmode=") {
		separator := "?"
		if strings.Contains(result, "?") {
			separator = "&"
		}
		result += separator + "sslmode=disable"
	}

	return result
}

// convertMySQLURL converts mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
func convertMySQLURL(prismaURL string) (string, error) {
	u, err := url.Parse(prismaURL)
	if err != nil {
		return "", err
	}

	// Extract user:password
	user := u.User.Username()
	password, _ := u.User.Password()

	// Extract host:port
	host := u.Host
	if host == "" {
		host = "localhost:3306"
	}

	// Extract database name (path without leading /)
	dbName := strings.TrimPrefix(u.Path, "/")

	// Build MySQL DSN: user:password@tcp(host:port)/dbname?params
	dsn := ""
	if password != "" {
		dsn = user + ":" + password + "@tcp(" + host + ")/" + dbName
	} else {
		dsn = user + "@tcp(" + host + ")/" + dbName
	}

	// Add query params
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	} else {
		// Default params for proper datetime parsing
		dsn += "?parseTime=true"
	}

	return dsn, nil
}

// Connect connects to the database
func (c *Client) Connect(cfg *Config) error {
	c.config = cfg
	return c.driver.Connect(cfg)
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.driver.Close()
}

// Ping verifies the connection is alive
func (c *Client) Ping() error {
	return c.driver.Ping()
}

// Query executes a query that returns rows
func (c *Client) Query(query string, args ...any) (*sql.Rows, error) {
	return c.driver.Query(query, args...)
}

// QueryRow executes a query that returns at most one row
func (c *Client) QueryRow(query string, args ...any) *sql.Row {
	return c.driver.QueryRow(query, args...)
}

// Exec executes a query without returning rows
func (c *Client) Exec(query string, args ...any) (sql.Result, error) {
	return c.driver.Exec(query, args...)
}

// DB returns the underlying *sql.DB for advanced operations
func (c *Client) DB() *sql.DB {
	return c.driver.DB()
}

// Driver returns the underlying driver
func (c *Client) Driver() DBDriver {
	return c.driver
}

// DriverName returns the name of the driver
func (c *Client) DriverName() string {
	return c.driver.Name()
}

// Config returns the connection config
func (c *Client) Config() *Config {
	return c.config
}
