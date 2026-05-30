package auth

import (
	"context"
	"errors"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnauthorized    = errors.New("unauthorized")
)

type contextKey struct{}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims)
}

func FromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(contextKey{}).(*Claims)
	if !ok {
		return nil
	}

	return claims
}

func RequireAuth(ctx context.Context) (*Claims, error) {
	claims := FromContext(ctx)
	if claims == nil {
		return nil, ErrUnauthenticated
	}

	return claims, nil
}

func RequireRole(ctx context.Context, roles ...string) (*Claims, error) {
	claims, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		if claims.Role == role {
			return claims, nil
		}
	}

	return nil, ErrUnauthorized
}
