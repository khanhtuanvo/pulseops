package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleIssuer = "https://accounts.google.com"

var (
	googleJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"
	jwksCache     = cachedJWKS{}
)

type GoogleClaims struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

type cachedJWKS struct {
	mu        sync.Mutex
	keySet    *jose.JSONWebKeySet
	expiresAt time.Time
}

type googleIDTokenClaims struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Audience string `json:"aud"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Expiry   int64  `json:"exp"`
}

func NewGoogleOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func GenerateStateToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}

func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code, codeVerifier string) (*oauth2.Token, error) {
	return cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
}

func ValidateIDToken(ctx context.Context, rawIDToken string, clientID string) (*GoogleClaims, error) {
	keySet, err := getGoogleJWKS(ctx)
	if err != nil {
		return nil, err
	}

	jws, err := jose.ParseSigned(rawIDToken, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return nil, fmt.Errorf("parse id token: %w", err)
	}

	payload, err := verifySignedPayload(jws, keySet)
	if err != nil {
		return nil, err
	}

	var claims googleIDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("decode id token claims: %w", err)
	}
	if claims.Issuer != googleIssuer && claims.Issuer != "accounts.google.com" {
		return nil, errors.New("invalid id token issuer")
	}
	if claims.Audience != clientID {
		return nil, errors.New("invalid id token audience")
	}
	if claims.Expiry <= time.Now().Unix() {
		return nil, errors.New("id token expired")
	}
	if claims.Subject == "" || claims.Email == "" {
		return nil, errors.New("id token missing required claims")
	}

	return &GoogleClaims{
		Subject: claims.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
	}, nil
}

func verifySignedPayload(jws *jose.JSONWebSignature, keySet *jose.JSONWebKeySet) ([]byte, error) {
	if len(jws.Signatures) == 0 {
		return nil, errors.New("id token has no signatures")
	}

	keyID := jws.Signatures[0].Header.KeyID
	for _, key := range keySet.Keys {
		if keyID != "" && key.KeyID != "" && key.KeyID != keyID {
			continue
		}

		payload, err := jws.Verify(key)
		if err == nil {
			return payload, nil
		}
	}

	return nil, errors.New("verify id token signature")
}

func getGoogleJWKS(ctx context.Context) (*jose.JSONWebKeySet, error) {
	jwksCache.mu.Lock()
	if jwksCache.keySet != nil && time.Now().Before(jwksCache.expiresAt) {
		defer jwksCache.mu.Unlock()
		return jwksCache.keySet, nil
	}
	jwksCache.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleJWKSURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch google jwks: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var keySet jose.JSONWebKeySet
	if err := json.Unmarshal(body, &keySet); err != nil {
		return nil, err
	}
	jwksCache.mu.Lock()
	jwksCache.keySet = &keySet
	jwksCache.expiresAt = time.Now().Add(time.Hour)
	jwksCache.mu.Unlock()

	return &keySet, nil
}
