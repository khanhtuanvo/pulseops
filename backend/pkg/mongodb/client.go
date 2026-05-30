package mongodb

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var pingClient = func(ctx context.Context, client *mongo.Client) error {
	return client.Ping(ctx, readpref.Primary())
}

func Connect(uri, dbName string, monitors ...*event.CommandMonitor) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	if len(monitors) > 0 && monitors[0] != nil {
		clientOptions.SetMonitor(monitors[0])
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := pingClient(ctx, client); err != nil {
		disconnectClient(client)
		return nil, err
	}

	return client.Database(dbName), nil
}

func Disconnect(client *mongo.Client) {
	disconnectClient(client)
}

func disconnectClient(client *mongo.Client) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Printf("disconnect mongodb: %v", err)
	}
}
