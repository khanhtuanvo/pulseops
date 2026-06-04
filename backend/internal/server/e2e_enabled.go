//go:build e2etest

package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"

	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// E2E fixtures. The team ID matches the synthetic claims minted by the auth
// e2e bypass, so incidents created via the seeded API key are scoped to the
// same team the E2E dashboard subscribes to.
const (
	e2eTeamID  = "000000000000000000000002"
	e2eAPIKey  = "e2e-test-api-key"
	e2eOwnerID = "000000000000000000000001"
)

// SeedE2EData upserts a deterministic team with a known API key so E2E tests
// can POST alerts to the webhook endpoint. It is a no-op unless ENV=test.
func SeedE2EData(ctx context.Context, db *mongo.Database) error {
	if os.Getenv("ENV") != "test" {
		return nil
	}

	teamID, err := primitive.ObjectIDFromHex(e2eTeamID)
	if err != nil {
		return err
	}
	ownerID, err := primitive.ObjectIDFromHex(e2eOwnerID)
	if err != nil {
		return err
	}

	sum := sha256.Sum256([]byte(e2eAPIKey))
	_, err = db.Collection("teams").UpdateOne(
		ctx,
		bson.M{"_id": teamID},
		bson.M{
			"$set":         bson.M{"name": "E2E Team", "apiKeyHash": hex.EncodeToString(sum[:]), "apiKeyHint": e2eAPIKey[len(e2eAPIKey)-4:], "ownerId": ownerID},
			"$setOnInsert": bson.M{"createdAt": time.Now().UTC()},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func e2eUserFromClaims(claims auth.Claims) (authUserDoc, bool) {
	if os.Getenv("ENV") != "test" || claims.Issuer != "pulseops-e2e" {
		return authUserDoc{}, false
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return authUserDoc{}, false
	}
	teamID, err := primitive.ObjectIDFromHex(claims.TeamID)
	if err != nil {
		return authUserDoc{}, false
	}

	return authUserDoc{
		ID:            userID,
		Email:         claims.Email,
		Name:          "E2E " + claims.Role,
		TeamID:        teamID,
		Role:          claims.Role,
		GoogleSubject: "e2e-" + claims.Role,
		CreatedAt:     time.Unix(0, 0).UTC(),
	}, true
}
