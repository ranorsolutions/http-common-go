package mongo

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//
// --- Helpers ---
//

func resetEnv(keys ...string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

//
// --- GetFromEnv() tests ---
//

func TestGetFromEnv_Success(t *testing.T) {
	defer resetEnv("DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT")

	os.Setenv("DB_USER", "envuser")
	os.Setenv("DB_PASSWORD", "envpass")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "27017")

	cfg, err := GetFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := &MongoConfig{
		DbUser:     "envuser",
		DbPassword: "envpass",
		DbHost:     "localhost",
		DbPort:     "27017",
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("expected %+v, got %+v", want, cfg)
	}
}

func TestGetFromEnv_MissingVars(t *testing.T) {
	defer resetEnv("DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT")

	os.Setenv("DB_USER", "envuser") // Missing others

	_, err := GetFromEnv()
	if err == nil {
		t.Fatal("expected error for missing environment variables")
	}
}

//
// --- URI() tests ---
//

func TestURI_WithAuth(t *testing.T) {
	cfg := &MongoConfig{
		DbUser:     "user",
		DbPassword: "pass",
		DbHost:     "localhost",
		DbPort:     "27017",
	}

	want := "mongodb://user:pass@localhost:27017"
	got := cfg.URI()
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func TestURI_NoAuth(t *testing.T) {
	cfg := &MongoConfig{
		DbHost: "localhost",
		DbPort: "27017",
	}

	want := "mongodb://localhost:27017"
	got := cfg.URI()
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

//
// --- New() tests ---
//

func TestNew_EmptyURI(t *testing.T) {
	_, err := New("testdb", "")
	if err == nil {
		t.Fatal("expected error when URI is empty")
	}
}

func TestNew_ValidURIFormat(t *testing.T) {
	db, err := New("testdb", "mongodb://localhost:27017")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db == nil || db.Connection == nil {
		t.Fatalf("expected valid MongoDB struct, got nil")
	}
}

//
// --- Mocks implementing the new interfaces ---
//

type mockIndexes struct {
	createErr error
}

func (m *mockIndexes) CreateMany(ctx context.Context, models []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return []string{"idx_test"}, nil
}

type mockCollection struct {
	indexView IndexViewAdapter
}

func (m *mockCollection) Indexes() IndexViewAdapter {
	return m.indexView
}

type mockClient struct {
	pingErr error
	called  bool
}

func (m *mockClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	m.called = true
	return m.pingErr
}

type mockDatabase struct {
	col    CollectionAdapter
	client ClientAdapter
}

func (m *mockDatabase) Collection(name string) CollectionAdapter {
	return m.col
}

func (m *mockDatabase) Client() ClientAdapter {
	return m.client
}

//
// --- CreateIndex() tests ---
//

func TestCreateIndex_Success(t *testing.T) {
	idx := &mockIndexes{}
	col := &mockCollection{indexView: idx}
	client := &mockClient{}
	db := &MongoDB{
		Name:       "testdb",
		Connection: &mockDatabase{col: col, client: client},
	}

	err := db.CreateIndex("users", []mongo.IndexModel{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateIndex_Error(t *testing.T) {
	idx := &mockIndexes{createErr: errors.New("index creation failed")}
	col := &mockCollection{indexView: idx}
	client := &mockClient{}
	db := &MongoDB{
		Name:       "testdb",
		Connection: &mockDatabase{col: col, client: client},
	}

	err := db.CreateIndex("users", []mongo.IndexModel{})
	if err == nil {
		t.Fatal("expected error from CreateIndex()")
	}
}

//
// --- HealthCheck() tests ---
//

func TestHealthCheck_Success(t *testing.T) {
	client := &mockClient{}
	db := &MongoDB{
		Name:       "testdb",
		Connection: &mockDatabase{client: client},
	}

	err := db.HealthCheck()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !client.called {
		t.Error("expected Ping() to be called")
	}
}

func TestHealthCheck_Failure(t *testing.T) {
	client := &mockClient{pingErr: errors.New("ping failed")}
	db := &MongoDB{
		Name:       "testdb",
		Connection: &mockDatabase{client: client},
	}

	err := db.HealthCheck()
	if err == nil {
		t.Fatal("expected error from Ping()")
	}
}

//
// --- Performance sanity test (ensures non-blocking) ---
//

func TestHealthCheck_FastResponse(t *testing.T) {
	client := &mockClient{}
	db := &MongoDB{
		Name:       "testdb",
		Connection: &mockDatabase{client: client},
	}

	start := time.Now()
	_ = db.HealthCheck()
	duration := time.Since(start)
	if duration > time.Second {
		t.Errorf("expected test to complete fast, took %v", duration)
	}
}
