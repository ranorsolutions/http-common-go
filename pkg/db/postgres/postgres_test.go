package postgres

import (
	"os"
	"reflect"
	"testing"
)

// Utility to reset env vars between tests
func resetEnv(keys ...string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

// --- String() and HostString() tests ---

func TestConnectionStringVariants(t *testing.T) {
	tests := []struct {
		name     string
		conn     Connection
		expected string
	}{
		{
			name: "Full credentials",
			conn: Connection{
				User:     "user",
				Password: "pass",
				Host:     "localhost",
				Port:     "5432",
				DB:       "testdb",
				SSLMode:  "disable",
			},
			expected: "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "User without password",
			conn: Connection{
				User:    "user",
				Host:    "localhost",
				Port:    "5432",
				DB:      "testdb",
				SSLMode: "disable",
			},
			expected: "postgres://user@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "No user or password",
			conn: Connection{
				Host:    "localhost",
				Port:    "5432",
				DB:      "testdb",
				SSLMode: "disable",
			},
			expected: "postgres://localhost:5432/testdb?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.conn.String()
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestHostString(t *testing.T) {
	conn := Connection{
		Host:    "localhost",
		Port:    "5432",
		DB:      "testdb",
		SSLMode: "disable",
	}
	expected := "postgres://localhost:5432/testdb?sslmode=disable"

	got := conn.HostString()
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

// --- GetURIFromEnv() tests ---

func TestGetURIFromEnv(t *testing.T) {
	defer resetEnv("DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT", "DB_NAME", "DB_SSL_MODE")

	os.Setenv("DB_USER", "envuser")
	os.Setenv("DB_PASSWORD", "envpass")
	os.Setenv("DB_HOST", "envhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "envdb")
	os.Setenv("DB_SSL_MODE", "require")

	expected := &Connection{
		User:     "envuser",
		Password: "envpass",
		Host:     "envhost",
		Port:     "5432",
		DB:       "envdb",
		SSLMode:  "require",
	}

	got := GetURIFromEnv()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %+v, got %+v", expected, got)
	}
}

// --- Connect() tests ---

func TestConnect_InvalidConnection(t *testing.T) {
	conn := &Connection{
		Host:    "invalid-host",
		Port:    "9999",
		DB:      "testdb",
		SSLMode: "disable",
	}
	db, err := Connect(conn)
	if err == nil {
		// Even though sql.Open doesn’t check the connection immediately,
		// we still expect db.Ping() to fail for an invalid host.
		defer db.Close()
		if pingErr := db.Ping(); pingErr == nil {
			t.Errorf("expected connection failure, but Ping succeeded")
		}
	}
}

func TestConnect_ValidDSNFormat(t *testing.T) {
	conn := &Connection{
		Host:    "localhost",
		Port:    "5432",
		DB:      "postgres",
		SSLMode: "disable",
	}
	db, err := Connect(conn)
	if err != nil && db != nil {
		t.Errorf("expected sql.Open to return a valid *sql.DB, got error: %v", err)
	}
	if db != nil {
		_ = db.Close()
	}
}
