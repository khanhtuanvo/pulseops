//go:build e2etest

package server

import (
	"os"
	"time"

	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
