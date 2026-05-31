//go:build !e2etest

package auth

import "net/http"

func e2eClaimsFromRequest(_ *http.Request) *Claims {
	return nil
}
