package auth

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

const SessionCookieName = "session"

func Middleware(jwtSecret string, loggers ...*zap.Logger) func(http.Handler) http.Handler {
	var logger *zap.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := ValidateJWT(cookie.Value, jwtSecret)
			if err != nil {
				if errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrTokenInvalid) {
					clearSessionCookie(w)
				}
				if logger != nil {
					logger.Error("jwt validation failed",
						zap.Error(err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
					)
				}

				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   SessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
