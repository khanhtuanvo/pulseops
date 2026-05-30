package incidents

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	db *mongo.Database
}

type ListFilters struct {
	Status   *string
	Severity *string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

type StatusUpdate struct {
	Status            string
	AcknowledgedAt    *time.Time
	AcknowledgedBy    *primitive.ObjectID
	ResolvedAt        *time.Time
	ResolvedBy        *primitive.ObjectID
	StatusMessage     *string
	ResolutionSummary *string
	MTTR              *int
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, doc IncidentDoc) (*IncidentDoc, error) {
	if doc.ID.IsZero() {
		doc.ID = primitive.NewObjectID()
	}

	if _, err := r.collection().InsertOne(ctx, doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) GetByID(ctx context.Context, id, teamID string) (*IncidentDoc, error) {
	incidentID, teamObjectID, err := parseIncidentAndTeamIDs(id, teamID)
	if err != nil {
		return nil, err
	}

	var doc IncidentDoc
	err = r.collection().FindOne(ctx, bson.M{"_id": incidentID, "teamId": teamObjectID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) GetByFingerprint(ctx context.Context, fingerprint, teamID string) (*IncidentDoc, error) {
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	var doc IncidentDoc
	err = r.collection().FindOne(ctx, bson.M{"fingerprint": fingerprint, "teamId": teamObjectID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) List(ctx context.Context, teamID string, filters ListFilters) ([]*IncidentDoc, int64, error) {
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"teamId": teamObjectID}
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if filters.Severity != nil {
		filter["severity"] = *filters.Severity
	}
	if filters.From != nil || filters.To != nil {
		timeFilter := bson.M{}
		if filters.From != nil {
			timeFilter["$gte"] = *filters.From
		}
		if filters.To != nil {
			timeFilter["$lte"] = *filters.To
		}
		filter["triggeredAt"] = timeFilter
	}

	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "triggeredAt", Value: -1}})
	if filters.Limit > 0 {
		findOptions.SetLimit(int64(filters.Limit))
	}
	if filters.Offset > 0 {
		findOptions.SetSkip(int64(filters.Offset))
	}

	cursor, err := r.collection().Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var docs []*IncidentDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id, teamID string, update StatusUpdate) (*IncidentDoc, error) {
	incidentID, teamObjectID, err := parseIncidentAndTeamIDs(id, teamID)
	if err != nil {
		return nil, err
	}

	set := bson.M{"status": update.Status}
	if update.AcknowledgedAt != nil {
		set["acknowledgedAt"] = *update.AcknowledgedAt
	}
	if update.AcknowledgedBy != nil {
		set["acknowledgedBy"] = *update.AcknowledgedBy
	}
	if update.ResolvedAt != nil {
		set["resolvedAt"] = *update.ResolvedAt
	}
	if update.ResolvedBy != nil {
		set["resolvedBy"] = *update.ResolvedBy
	}
	if update.StatusMessage != nil {
		set["statusMessage"] = *update.StatusMessage
	}
	if update.ResolutionSummary != nil {
		set["resolutionSummary"] = *update.ResolutionSummary
	}
	if update.MTTR != nil {
		set["mttr"] = *update.MTTR
	}

	var doc IncidentDoc
	err = r.collection().FindOneAndUpdate(
		ctx,
		bson.M{"_id": incidentID, "teamId": teamObjectID},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) IncrementAlertCount(ctx context.Context, id, teamID string) error {
	incidentID, teamObjectID, err := parseIncidentAndTeamIDs(id, teamID)
	if err != nil {
		return err
	}

	_, err = r.collection().UpdateOne(
		ctx,
		bson.M{"_id": incidentID, "teamId": teamObjectID},
		bson.M{"$inc": bson.M{"alertCount": 1}},
	)
	return err
}

func (r *Repository) collection() *mongo.Collection {
	return r.db.Collection("incidents")
}

func parseIncidentAndTeamIDs(id, teamID string) (primitive.ObjectID, primitive.ObjectID, error) {
	incidentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	return incidentID, teamObjectID, nil
}
