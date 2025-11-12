// Package postgres provides utilities for managing PostgreSQL database connections.
// It supports environment-based configuration and connection string generation
// for use across different services.
package postgres

import (
	"database/sql"
	"fmt"
	"os"

	// Import the PostgreSQL driver
	_ "github.com/lib/pq"
)

// Connection defines parameters required to establish a connection
// to a PostgreSQL database.
type Connection struct {
	User     string // Database user
	Password string // Database password
	Host     string // Database host (e.g. "localhost" or remote host)
	Port     string // Database port (e.g. "5432")
	DB       string // Database name
	SSLMode  string // SSL mode (e.g. "disable", "require")
}

// String builds the PostgreSQL connection URI based on available fields.
// It supports cases where authentication may not include a password or even a username.
func (c *Connection) String() string {
	switch {
	case c.User == "" && c.Password == "":
		return fmt.Sprintf("postgres://%s:%s/%s?sslmode=%s", c.Host, c.Port, c.DB, c.SSLMode)
	case c.Password == "":
		return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", c.User, c.Host, c.Port, c.DB, c.SSLMode)
	default:
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.DB, c.SSLMode)
	}
}

// HostString returns a connection URI that omits authentication credentials.
// Useful for logging or non-sensitive operations.
func (c *Connection) HostString() string {
	return fmt.Sprintf("postgres://%s:%s/%s?sslmode=%s", c.Host, c.Port, c.DB, c.SSLMode)
}

// GetURIFromEnv constructs a Connection from standard environment variables:
//
//	DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME, DB_SSL_MODE
func GetURIFromEnv() *Connection {
	return &Connection{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DB:       os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	}
}

// Connect opens a connection to PostgreSQL using the provided Connection configuration.
// It returns a *sql.DB instance which can be used for executing queries and transactions.
func Connect(conn *Connection) (*sql.DB, error) {
	return sql.Open("postgres", conn.String())
}
