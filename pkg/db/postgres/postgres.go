package postgres

import (
	_ "github.com/lib/pq"

	"database/sql"
	"fmt"
	"os"
)

type Connection struct {
	User     string
	Password string
	Host     string
	Port     string
	DB       string
	SSLMode  string
}

func (c *Connection) String() string {
	if c.User == "" || c.Password == "" {
		if c.User != "" {
			return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", c.User, c.Host, c.Port, c.DB, c.SSLMode)
		}

		return fmt.Sprintf("postgres://%s:%s/%s?sslmode=%s", c.Host, c.Port, c.DB, c.SSLMode)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.DB, c.SSLMode)
}

func (c *Connection) HostString() string {
	return fmt.Sprintf("postgres://%s:%s/%s?sslmode=%s", c.Host, c.Port, c.DB, c.SSLMode)
}

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

func Connect(conn *Connection) (*sql.DB, error) {
	return sql.Open("postgres", conn.String())
}
