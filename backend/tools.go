//go:build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/go-jose/go-jose/v4"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/rs/cors"
	_ "github.com/stretchr/testify"
	_ "go.mongodb.org/mongo-driver/mongo"
	_ "go.uber.org/zap"
	_ "golang.org/x/oauth2"
	_ "golang.org/x/oauth2/google"
)
