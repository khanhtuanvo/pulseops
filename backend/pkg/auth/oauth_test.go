package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"
)

func TestGenerateStateToken(t *testing.T) {
	first, err := GenerateStateToken()
	require.NoError(t, err)
	second, err := GenerateStateToken()
	require.NoError(t, err)

	require.Len(t, first, 64)
	require.Regexp(t, regexp.MustCompile(`^[a-f0-9]{64}$`), first)
	require.NotEqual(t, first, second)
}

func TestValidateIDToken(t *testing.T) {
	privateKey := mustRSAKey(t)
	keySet := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key:       &privateKey.PublicKey,
		KeyID:     "test-key",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}}}
	server := serveJWKS(t, keySet)
	withTestJWKSURL(t, server.URL)

	rawIDToken := mustGoogleIDToken(t, privateKey, "client-id", time.Now().Add(time.Hour))

	claims, err := ValidateIDToken(context.Background(), rawIDToken, "client-id")
	require.NoError(t, err)
	require.Equal(t, "google-subject", claims.Subject)
	require.Equal(t, "user@example.com", claims.Email)
	require.Equal(t, "Test User", claims.Name)
	require.Equal(t, "https://example.com/avatar.png", claims.Picture)
}

func TestValidateIDTokenExpired(t *testing.T) {
	privateKey := mustRSAKey(t)
	server := serveJWKS(t, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key:       &privateKey.PublicKey,
		KeyID:     "test-key",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}}})
	withTestJWKSURL(t, server.URL)

	rawIDToken := mustGoogleIDToken(t, privateKey, "client-id", time.Now().Add(-time.Hour))

	_, err := ValidateIDToken(context.Background(), rawIDToken, "client-id")
	require.ErrorContains(t, err, "expired")
}

func TestValidateIDTokenWrongAudience(t *testing.T) {
	privateKey := mustRSAKey(t)
	server := serveJWKS(t, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key:       &privateKey.PublicKey,
		KeyID:     "test-key",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}}})
	withTestJWKSURL(t, server.URL)

	rawIDToken := mustGoogleIDToken(t, privateKey, "other-client-id", time.Now().Add(time.Hour))

	_, err := ValidateIDToken(context.Background(), rawIDToken, "client-id")
	require.ErrorContains(t, err, "audience")
}

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func serveJWKS(t *testing.T, keySet jose.JSONWebKeySet) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(keySet))
	}))
	t.Cleanup(server.Close)
	return server
}

func withTestJWKSURL(t *testing.T, url string) {
	t.Helper()

	originalURL := googleJWKSURL
	googleJWKSURL = url
	jwksCache = cachedJWKS{}

	t.Cleanup(func() {
		googleJWKSURL = originalURL
		jwksCache = cachedJWKS{}
	})
}

func mustGoogleIDToken(t *testing.T, privateKey *rsa.PrivateKey, audience string, expiresAt time.Time) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.RS256,
			Key: jose.JSONWebKey{
				Key:   privateKey,
				KeyID: "test-key",
			},
		},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	require.NoError(t, err)

	payload, err := json.Marshal(map[string]interface{}{
		"iss":     googleIssuer,
		"sub":     "google-subject",
		"aud":     audience,
		"email":   "user@example.com",
		"name":    "Test User",
		"picture": "https://example.com/avatar.png",
		"exp":     expiresAt.Unix(),
	})
	require.NoError(t, err)

	jws, err := signer.Sign(payload)
	require.NoError(t, err)

	token, err := jws.CompactSerialize()
	require.NoError(t, err)
	return token
}
