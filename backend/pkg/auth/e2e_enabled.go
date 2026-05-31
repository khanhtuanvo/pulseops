//go:build e2etest

package auth

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const e2eTestUserHeader = "X-E2E-Test-User"

func e2eClaimsFromRequest(r *http.Request) *Claims {
	if os.Getenv("ENV") != "test" {
		return nil
	}

	role := strings.ToUpper(strings.TrimSpace(r.Header.Get(e2eTestUserHeader)))
	if role == "" {
		return nil
	}
	if role != "OWNER" && role != "RESPONDER" && role != "VIEWER" {
		role = "OWNER"
	}

	return &Claims{
		UserID: "000000000000000000000001",
		TeamID: "000000000000000000000002",
		Role:   role,
		Email:  strings.ToLower(role) + "@e2e.pulseops.test",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "pulseops-e2e",
			Subject:   "000000000000000000000001",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
}
