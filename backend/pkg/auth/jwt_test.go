package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestJWTRoundTrip(t *testing.T) {
	token, err := SignJWT(Claims{
		UserID: "user-1",
		TeamID: "team-1",
		Role:   "OWNER",
		Email:  "owner@example.com",
	}, "secret", 15)
	require.NoError(t, err)

	claims, err := ValidateJWT(token, "secret")
	require.NoError(t, err)
	require.Equal(t, "user-1", claims.UserID)
	require.Equal(t, "team-1", claims.TeamID)
	require.Equal(t, "OWNER", claims.Role)
	require.Equal(t, "owner@example.com", claims.Email)
	require.Equal(t, "user-1", claims.Subject)
}

func TestValidateJWTExpiredToken(t *testing.T) {
	token, err := SignJWT(Claims{
		UserID: "user-1",
		TeamID: "team-1",
		Role:   "OWNER",
		Email:  "owner@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}, "secret", 15)
	require.NoError(t, err)

	_, err = ValidateJWT(token, "secret")
	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestValidateJWTInvalidSignature(t *testing.T) {
	token, err := SignJWT(Claims{
		UserID: "user-1",
		TeamID: "team-1",
		Role:   "OWNER",
		Email:  "owner@example.com",
	}, "secret", 15)
	require.NoError(t, err)

	_, err = ValidateJWT(token, "other-secret")
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func TestRequireRoleWrongRole(t *testing.T) {
	ctx := WithClaims(context.Background(), &Claims{Role: "VIEWER"})

	_, err := RequireRole(ctx, "OWNER", "RESPONDER")
	require.True(t, errors.Is(err, ErrUnauthorized))
}
