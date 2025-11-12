// Package mongo provides helper utilities to configure, connect to, and manage
// MongoDB databases in Go applications. It supports environment-based
// configuration and simplified connection management for other services.
package mongo

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//
// --- Configuration --
//

// MongoConfig holds configuration values required to establish a MongoDB connection.
// Values are typically sourced from environment variables.
type MongoConfig struct {
	DbUser     string // Database username
	DbPassword string // Database password
	DbHost     string // Database host (e.g. "localhost")
	DbPort     string // Database port (e.g. "27017")
}

// GetFromEnv constructs a MongoConfig from standard environment variables:
//
//	DB_USER, DB_PASSWORD, DB_HOST, DB_PORT
//
// It returns an error if any of these required variables are missing.
func GetFromEnv() (*MongoConfig, error) {
	required := []string{"DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT"}
	for _, env := range required {
		if os.Getenv(env) == "" {
			return nil, fmt.Errorf("%s is required to initialize the MongoDB connection", env)
		}
	}

	return &MongoConfig{
		DbUser:     os.Getenv("DB_USER"),
		DbPassword: os.Getenv("DB_PASSWORD"),
		DbHost:     os.Getenv("DB_HOST"),
		DbPort:     os.Getenv("DB_PORT"),
	}, nil
}

// URI generates a MongoDB connection URI from the configuration values.
// If no user is specified, it returns a no-auth connection URI.
func (config *MongoConfig) URI() string {
	if config.DbUser == "" {
		return fmt.Sprintf("mongodb://%s:%s", config.DbHost, config.DbPort)
	}

	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%s",
		config.DbUser,
		config.DbPassword,
		config.DbHost,
		config.DbPort,
	)
}

//
// --- Interfaces for Dependency Injection --
//

// DatabaseAdapter defines the minimal interface of mongo.Database needed by this package.
type DatabaseAdapter interface {
	Collection(name string) CollectionAdapter
	Client() ClientAdapter
}

// CollectionAdapter abstracts a MongoDB collection used for index creation.
type CollectionAdapter interface {
	Indexes() IndexViewAdapter
}

// IndexViewAdapter abstracts the index creation API.
type IndexViewAdapter interface {
	CreateMany(ctx context.Context, models []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error)
}

// ClientAdapter defines the minimal interface of mongo.Client for health checks.
type ClientAdapter interface {
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

//
// --- Concrete Implementations (wrappers around mongo.*) --
//

// realDatabase wraps mongo.Database to satisfy DatabaseAdapter.
type realDatabase struct {
	db *mongo.Database
}

func (r *realDatabase) Collection(name string) CollectionAdapter {
	return &realCollection{col: r.db.Collection(name)}
}

func (r *realDatabase) Client() ClientAdapter {
	return &realClient{client: r.db.Client()}
}

type realCollection struct {
	col *mongo.Collection
}

func (r *realCollection) Indexes() IndexViewAdapter {
	return &realIndexView{idx: r.col.Indexes()}
}

type realIndexView struct {
	idx mongo.IndexView
}

func (r *realIndexView) CreateMany(ctx context.Context, models []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	return r.idx.CreateMany(ctx, models, opts...)
}

type realClient struct {
	client *mongo.Client
}

func (r *realClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return r.client.Ping(ctx, rp)
}

//
// --- MongoDB wrapper struct ---
//

// MongoDB represents an active connection to a MongoDB database.
type MongoDB struct {
	Name       string
	Connection DatabaseAdapter
}

// New creates a new MongoDB client and connects to the database at the given URI.
// It applies a 10-second timeout for establishing the connection.
//
// Example:
//
//	db, err := mongo.New("appdb", "mongodb://localhost:27017")
//	if err != nil { ... }
func New(name, uri string) (*MongoDB, error) {
	if uri == "" {
		return nil, fmt.Errorf("MongoDB connection URI cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(name)
	return &MongoDB{
		Name:       name,
		Connection: &realDatabase{db: db},
	}, nil
}

// CreateIndex creates one or more indexes on a given collection.
//
// Example:
//
//	idx := mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)}
//	err := db.CreateIndex("users", []mongo.IndexModel{idx})
func (db *MongoDB) CreateIndex(collectionName string, indexes []mongo.IndexModel) error {
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Connection.Collection(collectionName)
	_, err := collection.Indexes().CreateMany(ctx, indexes, opts)
	return err
}

// HealthCheck verifies the connectivity to the MongoDB instance by pinging it.
// It returns nil if the connection is healthy.
func (db *MongoDB) HealthCheck() error {
	return db.Connection.Client().Ping(context.Background(), nil)
}
