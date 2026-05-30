package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tuankhanhvo/pulseops/internal/streams"
	"github.com/tuankhanhvo/pulseops/pkg/config"
	"go.uber.org/zap"
)

func TestRouterHealth(t *testing.T) {
	router := NewRouter(testConfig(), nil, zap.NewNop(), streams.NewHub())

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
	require.NotEmpty(t, recorder.Header().Get(requestIDHeader))
}

func TestRouterLogoutWithoutCookieReturnsUnauthorized(t *testing.T) {
	router := NewRouter(testConfig(), nil, zap.NewNop(), streams.NewHub())

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/auth/logout", nil))

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func testConfig() *config.Config {
	return &config.Config{
		Env:                    "test",
		AllowedOrigins:         "http://localhost:5173",
		JWTSecret:              "test-secret",
		JWTExpiryMinutes:       15,
		RefreshTokenExpiryDays: 7,
		GoogleClientID:         "client-id",
		GoogleClientSecret:     "client-secret",
		OAuthRedirectURL:       "http://localhost:8080/auth/callback",
	}
}
