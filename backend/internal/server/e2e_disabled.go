//go:build !e2etest

package server

import "github.com/tuankhanhvo/pulseops/pkg/auth"

func e2eUserFromClaims(_ auth.Claims) (authUserDoc, bool) {
	return authUserDoc{}, false
}
