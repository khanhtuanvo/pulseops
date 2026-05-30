package teams

import (
	"context"
	"errors"
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

func (r *Repository) FindUserByGoogleSubject(ctx context.Context, subject string) (*incidents.UserDoc, error) {
	return r.findUser(ctx, bson.M{"googleSubject": subject})
}

func (r *Repository) FindUserByID(ctx context.Context, id, teamID string) (*incidents.UserDoc, error) {
	userID, teamObjectID, err := parseUserAndTeamIDs(id, teamID)
	if err != nil {
		return nil, err
	}

	return r.findUser(ctx, bson.M{"_id": userID, "teamId": teamObjectID})
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*incidents.UserDoc, error) {
	return r.findUser(ctx, bson.M{"email": email})
}

func (r *Repository) CreateUser(ctx context.Context, doc incidents.UserDoc) (*incidents.UserDoc, error) {
	if doc.ID.IsZero() {
		doc.ID = primitive.NewObjectID()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}

	if _, err := r.db.Collection("users").InsertOne(ctx, doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) FindTeamByID(ctx context.Context, id string) (*incidents.TeamDoc, error) {
	teamID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var doc incidents.TeamDoc
	err = r.db.Collection("teams").FindOne(ctx, bson.M{"_id": teamID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) FindTeamByAPIKeyHash(ctx context.Context, hash string) (*incidents.TeamDoc, error) {
	var doc incidents.TeamDoc
	err := r.db.Collection("teams").FindOne(ctx, bson.M{"apiKeyHash": hash}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) CreateTeam(ctx context.Context, doc incidents.TeamDoc) (*incidents.TeamDoc, error) {
	if doc.ID.IsZero() {
		doc.ID = primitive.NewObjectID()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}

	if _, err := r.db.Collection("teams").InsertOne(ctx, doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *Repository) ListTeamMembers(ctx context.Context, teamID string) ([]*incidents.UserDoc, error) {
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.db.Collection("users").Find(ctx, bson.M{"teamId": teamObjectID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*incidents.UserDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID, teamID string, role string) error {
	userObjectID, teamObjectID, err := parseUserAndTeamIDs(userID, teamID)
	if err != nil {
		return err
	}

	_, err = r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userObjectID, "teamId": teamObjectID}, bson.M{"$set": bson.M{"role": role}})
	return err
}

func (r *Repository) MoveUserToTeam(ctx context.Context, userID, teamID string, role string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	_, err = r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userObjectID}, bson.M{"$set": bson.M{"teamId": teamObjectID, "role": role}})
	return err
}

func (r *Repository) RemoveUserFromTeam(ctx context.Context, userID, teamID string) error {
	userObjectID, teamObjectID, err := parseUserAndTeamIDs(userID, teamID)
	if err != nil {
		return err
	}

	_, err = r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userObjectID, "teamId": teamObjectID}, bson.M{"$unset": bson.M{"teamId": ""}})
	return err
}

func (r *Repository) RotateAPIKey(ctx context.Context, teamID, hash, hint string) error {
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	_, err = r.db.Collection("teams").UpdateOne(ctx, bson.M{"_id": teamObjectID}, bson.M{"$set": bson.M{"apiKeyHash": hash, "apiKeyHint": hint}})
	return err
}

func (r *Repository) findUser(ctx context.Context, filter bson.M) (*incidents.UserDoc, error) {
	var doc incidents.UserDoc
	err := r.db.Collection("users").FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func parseUserAndTeamIDs(userID, teamID string) (primitive.ObjectID, primitive.ObjectID, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	return userObjectID, teamObjectID, nil
}
