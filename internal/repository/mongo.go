package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ConnectMongo initializes and tests a MongoDB connection with timeout
func ConnectMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri).SetTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping mongo database: %w", err)
	}

	return client, nil
}
