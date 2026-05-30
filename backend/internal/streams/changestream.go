package streams

import (
	"context"
	"errors"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func StartChangeStreamListener(ctx context.Context, db *mongo.Database, hub *Hub, logger *zap.Logger) {
	backoff := 3 * time.Second
	var resumeToken bson.Raw

	for {
		if ctx.Err() != nil {
			return
		}

		changeStream, err := openIncidentChangeStream(ctx, db, resumeToken)
		if err != nil {
			logger.Error("open incident change stream", zap.Error(err))
			if !sleepWithContext(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		backoff = 3 * time.Second
		for changeStream.Next(ctx) {
			var event incidentChangeEvent
			if err := changeStream.Decode(&event); err != nil {
				logger.Error("decode incident change event", zap.Error(err))
				continue
			}
			resumeToken = changeStream.ResumeToken()

			if event.FullDocument.ID.IsZero() {
				continue
			}

			hub.Publish(IncidentEvent{
				Type:       eventTypeForChange(event.OperationType, event.FullDocument.Status),
				IncidentID: event.FullDocument.ID.Hex(),
				TeamID:     event.FullDocument.TeamID.Hex(),
				Payload:    event.FullDocument,
			})
		}

		err = changeStream.Err()
		_ = changeStream.Close(ctx)
		if err == nil || errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return
		}

		logger.Error("incident change stream closed with error", zap.Error(err))
		if !sleepWithContext(ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff)
	}
}

type incidentChangeEvent struct {
	OperationType string                `bson:"operationType"`
	FullDocument  incidents.IncidentDoc `bson:"fullDocument"`
}

func openIncidentChangeStream(ctx context.Context, db *mongo.Database, resumeToken bson.Raw) (*mongo.ChangeStream, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"operationType": bson.M{"$in": []string{"insert", "update"}},
		}}},
	}
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	if len(resumeToken) > 0 {
		opts.SetResumeAfter(resumeToken)
	}

	return db.Collection("incidents").Watch(ctx, pipeline, opts)
}

func eventTypeForChange(operationType, status string) string {
	if operationType == "insert" {
		return "INCIDENT_CREATED"
	}

	switch status {
	case "ACKNOWLEDGED":
		return "INCIDENT_ACKNOWLEDGED"
	case "INVESTIGATING":
		return "INCIDENT_INVESTIGATING"
	case "RESOLVED", "CLOSED":
		return "INCIDENT_RESOLVED"
	case "ESCALATED":
		return "INCIDENT_ESCALATED"
	default:
		return "ALERT_ATTACHED"
	}
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > 30*time.Second {
		return 30 * time.Second
	}

	return next
}
