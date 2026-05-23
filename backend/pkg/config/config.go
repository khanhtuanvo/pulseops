package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Env                    string
	Port                   string
	AllowedOrigins         string
	MongoURI               string
	MongoDB                string
	JWTSecret              string
	JWTExpiryMinutes       int
	RefreshTokenExpiryDays int
	GoogleClientID         string
	GoogleClientSecret     string
	OAuthRedirectURL       string
}

func Load() Config {
	return Config{
		Env:                    getEnv("ENV", "development"),
		Port:                   getEnv("PORT", "8080"),
		AllowedOrigins:         getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		MongoURI:               mustEnv("MONGODB_URI"),
		MongoDB:                mustEnv("MONGODB_DB"),
		JWTSecret:              mustEnv("JWT_SECRET"),
		JWTExpiryMinutes:       mustIntEnv("JWT_EXPIRY_MINUTES", 15),
		RefreshTokenExpiryDays: mustIntEnv("REFRESH_TOKEN_EXPIRY_DAYS", 7),
		GoogleClientID:         mustEnv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:     mustEnv("GOOGLE_CLIENT_SECRET"),
		OAuthRedirectURL:       mustEnv("OAUTH_REDIRECT_URL"),
	}
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing required environment variable %s", key)
	}

	return value
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func mustIntEnv(key string, fallback int) int {
	value := getEnv(key, strconv.Itoa(fallback))

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("invalid integer environment variable %s: %v", key, err)
	}

	return parsed
}
