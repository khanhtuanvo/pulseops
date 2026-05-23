package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := map[string][]mongo.IndexModel{
		"incidents": {
			{Keys: bson.D{{Key: "teamId", Value: 1}, {Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "teamId", Value: 1}, {Key: "triggeredAt", Value: -1}}},
			{Keys: bson.D{{Key: "fingerprint", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"alerts": {
			{Keys: bson.D{{Key: "incidentId", Value: 1}}},
			{Keys: bson.D{{Key: "teamId", Value: 1}, {Key: "receivedAt", Value: -1}}},
		},
		"fingerprints": {
			{
				Keys:    bson.D{{Key: "createdAt", Value: 1}},
				Options: options.Index().SetExpireAfterSeconds(60),
			},
		},
		"sessions": {
			{
				Keys:    bson.D{{Key: "expiresAt", Value: 1}},
				Options: options.Index().SetExpireAfterSeconds(0),
			},
			{Keys: bson.D{{Key: "userId", Value: 1}}},
		},
		"rate_limits": {
			{
				Keys:    bson.D{{Key: "expiresAt", Value: 1}},
				Options: options.Index().SetExpireAfterSeconds(0),
			},
		},
		"users": {
			{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "googleSubject", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "teamId", Value: 1}}},
		},
		"on_call_schedules": {
			{Keys: bson.D{{Key: "teamId", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
	}

	for collection, models := range indexes {
		if _, err := db.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
			return err
		}
	}

	return nil
}
