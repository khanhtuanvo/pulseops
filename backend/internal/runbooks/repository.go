package runbooks

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Upsert(ctx context.Context, doc incidents.RunbookDoc) (*incidents.RunbookDoc, error) {
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = time.Now().UTC()
	}

	if doc.ID.IsZero() {
		doc.ID = primitive.NewObjectID()
		if _, err := r.collection().InsertOne(ctx, doc); err != nil {
			return nil, err
		}
		return &doc, nil
	}

	update := bson.M{
		"title":     doc.Title,
		"content":   doc.Content,
		"tags":      doc.Tags,
		"updatedAt": doc.UpdatedAt,
	}

	var updated incidents.RunbookDoc
	err := r.collection().FindOneAndUpdate(
		ctx,
		bson.M{"_id": doc.ID, "teamId": doc.TeamID},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *Repository) List(ctx context.Context, teamID string, query string) ([]*incidents.RunbookDoc, error) {
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"teamId": teamObjectID}
	if query != "" {
		// Escape the user input so it is treated as a literal substring rather
		// than a regular expression (prevents ReDoS and metacharacter surprises).
		pattern := primitive.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"title": pattern},
			bson.M{"tags": pattern},
		}
	}

	cursor, err := r.collection().Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*incidents.RunbookDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *Repository) GetByID(ctx context.Context, id, teamID string) (*incidents.RunbookDoc, error) {
	runbookID, teamObjectID, err := parseRunbookAndTeamIDs(id, teamID)
	if err != nil {
		return nil, err
	}

	var doc incidents.RunbookDoc
	err = r.collection().FindOne(ctx, bson.M{"_id": runbookID, "teamId": teamObjectID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) Delete(ctx context.Context, id, teamID string) error {
	runbookID, teamObjectID, err := parseRunbookAndTeamIDs(id, teamID)
	if err != nil {
		return err
	}

	_, err = r.collection().DeleteOne(ctx, bson.M{"_id": runbookID, "teamId": teamObjectID})
	return err
}

func (r *Repository) collection() *mongo.Collection {
	return r.db.Collection("runbooks")
}

func parseRunbookAndTeamIDs(id, teamID string) (primitive.ObjectID, primitive.ObjectID, error) {
	runbookID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	return runbookID, teamObjectID, nil
}
