package mongo

import (
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// MongoConfig -- Database confugiration
type MongoConfig struct {
	DbUser     string // DB Username
	DbPassword string // DB Password
	DbHost     string // Host that the DB will listen on
	DbPort     string // Port that the DB will listen on
}

// initialize -- Reads ENV variables and saves them into config struct
func GetFromEnv() (*MongoConfig, error) {
	// Check all variables
	envs := []string{"DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT"}
	for _, env := range envs {
		if os.Getenv(env) == "" {
			return nil, fmt.Errorf("%s is required to initialize the connection", env)
		}
	}

	config := &MongoConfig{
		DbUser:     os.Getenv(envs[0]),
		DbPassword: os.Getenv(envs[1]),
		DbHost:     os.Getenv(envs[2]),
		DbPort:     os.Getenv(envs[3]),
	}

	return config, nil
}

// URI -- Generates Mongo DB Connection URI
func (config *MongoConfig) URI() string {
	// Handle No-Auth Connections
	if config.DbUser == "" {
		return fmt.Sprintf(
			"mongodb://%s:%s",
			config.DbHost,
			config.DbPort,
		)
	}

	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%s",
		config.DbUser,
		config.DbPassword,
		config.DbHost,
		config.DbPort,
	)
}

type MongoDB struct {
	Name       string
	Connection *mongo.Database
}

func New(name, uri string) (*MongoDB, error) {
	// Create the Mongo instance
	mongoDB := &MongoDB{
		Name: name,
	}

	// Create the Context in the Background
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create the client and catch errors
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	// Handle any errors
	if err != nil {
		return nil, err
	}

	// Create the database instance
	mongoDB.Connection = client.Database(name)

	// Return the Database
	return mongoDB, nil
}

// CreateIndex - creates an index for a specific field in a collection
func (db *MongoDB) CreateIndex(collectionName string, indexes []mongo.IndexModel) error {
	// Declare an options object
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	// 1. Create the context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Connect to the database and access the collection
	collection := db.Connection.Collection(collectionName)

	// 3. Create a single index
	_, err := collection.Indexes().CreateMany(ctx, indexes, opts)

	// 4. Return the error to be handled
	return err
}

// HealthCheck -- Handler to determine the service is healthy
func (db *MongoDB) HealthCheck() error {
	// Ping the DB
	return db.Connection.Client().Ping(context.Background(), nil)
}
