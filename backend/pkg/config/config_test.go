package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("PORT", "9090")
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017/pulseops?replicaSet=rs0")
	t.Setenv("MONGODB_DB", "pulseops_test")
	t.Setenv("JWT_SECRET", "replace-with-32-char-random-string")
	t.Setenv("JWT_EXPIRY_MINUTES", "30")
	t.Setenv("REFRESH_TOKEN_EXPIRY_DAYS", "14")
	t.Setenv("GOOGLE_CLIENT_ID", "client-id.apps.googleusercontent.com")
	t.Setenv("GOOGLE_CLIENT_SECRET", "client-secret")
	t.Setenv("OAUTH_REDIRECT_URL", "http://localhost:8080/auth/callback")
	t.Setenv("OTEL_SERVICE_NAME", "pulseops-api-test")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://example.monitor.azure.com")

	cfg := Load()

	require.Equal(t, "test", cfg.Env)
	require.Equal(t, "9090", cfg.Port)
	require.Equal(t, "http://localhost:5173,http://localhost:3000", cfg.AllowedOrigins)
	require.Equal(t, "mongodb://localhost:27017/pulseops?replicaSet=rs0", cfg.MongoURI)
	require.Equal(t, "pulseops_test", cfg.MongoDB)
	require.Equal(t, "replace-with-32-char-random-string", cfg.JWTSecret)
	require.Equal(t, 30, cfg.JWTExpiryMinutes)
	require.Equal(t, 14, cfg.RefreshTokenExpiryDays)
	require.Equal(t, "client-id.apps.googleusercontent.com", cfg.GoogleClientID)
	require.Equal(t, "client-secret", cfg.GoogleClientSecret)
	require.Equal(t, "http://localhost:8080/auth/callback", cfg.OAuthRedirectURL)
	require.Equal(t, "pulseops-api-test", cfg.ServiceName)
	require.Equal(t, "https://example.monitor.azure.com", cfg.OTLPEndpoint)
}
