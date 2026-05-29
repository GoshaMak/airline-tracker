package mongo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(i do.Injector) (*mongo.Client, error) {
	host := os.Getenv("MONGO_HOST")
	port := os.Getenv("MONGO_PORT")
	user := os.Getenv("MONGO_USER")
	password := os.Getenv("MONGO_PASSWORD")
	dbName := os.Getenv("MONGO_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "27017"
	}

	var uri string

	if user != "" && password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
			user,
			password,
			host,
			port,
			dbName,
		)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s", host, port)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return client, nil
}

func NewMongoDatabase(i do.Injector) (*mongo.Database, error) {
	client := do.MustInvoke[*mongo.Client](i)
	dbName := os.Getenv("MONGO_NAME")
	if dbName == "" {
		dbName = os.Getenv("MONGO_DB")
	}
	if dbName == "" {
		dbName = "airline_tracker"
	}
	return client.Database(dbName), nil
}
