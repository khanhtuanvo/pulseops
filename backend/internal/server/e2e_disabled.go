//go:build !e2etest

package server

import (
	"context"

	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"go.mongodb.org/mongo-driver/mongo"
)

func e2eUserFromClaims(_ auth.Claims) (authUserDoc, bool) {
	return authUserDoc{}, false
}

// SeedE2EData is a no-op in production builds.
func SeedE2EData(_ context.Context, _ *mongo.Database) error {
	return nil
}
