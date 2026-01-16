package database

import "fmt"

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string            // disable, require, verify-ca, verify-full
	Extra    map[string]string // DBMS-specific options
}

// NewConfig creates a new Config with defaults
func NewConfig() *Config {
	return &Config{
		Host:    "localhost",
		Port:    5432,
		SSLMode: "disable",
		Extra:   make(map[string]string),
	}
}

// WithHost sets the host
func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

// WithPort sets the port
func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

// WithUser sets the user
func (c *Config) WithUser(user string) *Config {
	c.User = user
	return c
}

// WithPassword sets the password
func (c *Config) WithPassword(password string) *Config {
	c.Password = password
	return c
}

// WithDatabase sets the database name
func (c *Config) WithDatabase(database string) *Config {
	c.Database = database
	return c
}

// WithSSLMode sets the SSL mode
func (c *Config) WithSSLMode(mode string) *Config {
	c.SSLMode = mode
	return c
}

// WithExtra sets an extra option
func (c *Config) WithExtra(key, value string) *Config {
	if c.Extra == nil {
		c.Extra = make(map[string]string)
	}
	c.Extra[key] = value
	return c
}

// PostgresDSN returns a PostgreSQL connection string
func (c *Config) PostgresDSN() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)

	for k, v := range c.Extra {
		dsn += fmt.Sprintf(" %s=%s", k, v)
	}
	return dsn
}

// MySQLDSN returns a MySQL connection string
func (c *Config) MySQLDSN() string {
	// user:password@tcp(host:port)/database?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		c.User, c.Password, c.Host, c.Port, c.Database)

	// Default params for proper datetime parsing
	params := "parseTime=true"

	if c.SSLMode != "" && c.SSLMode != "disable" {
		params += "&tls=" + c.SSLMode
	}

	for k, v := range c.Extra {
		params += fmt.Sprintf("&%s=%s", k, v)
	}

	return dsn + "?" + params
}
