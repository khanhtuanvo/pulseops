package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareNoCookie(t *testing.T) {
	called := false
	handler := Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		require.Nil(t, FromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	require.True(t, called)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Empty(t, recorder.Result().Cookies())
}

func TestMiddlewareValidToken(t *testing.T) {
	token := mustTestJWT(t, "secret", time.Now().Add(15*time.Minute))
	handler := Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "user-1", FromContext(r.Context()).UserID)
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Empty(t, recorder.Result().Cookies())
}

func TestMiddlewareExpiredTokenClearsCookie(t *testing.T) {
	token := mustTestJWT(t, "secret", time.Now().Add(-time.Minute))
	handler := Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Nil(t, FromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, recorder.Result().Cookies(), 1)
	require.Equal(t, -1, recorder.Result().Cookies()[0].MaxAge)
}

func TestMiddlewareInvalidTokenClearsCookie(t *testing.T) {
	handler := Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Nil(t, FromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "not-a-jwt"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, recorder.Result().Cookies(), 1)
	require.Equal(t, -1, recorder.Result().Cookies()[0].MaxAge)
}

func TestMiddlewareWrongSignatureClearsCookie(t *testing.T) {
	token := mustTestJWT(t, "other-secret", time.Now().Add(15*time.Minute))
	handler := Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Nil(t, FromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, recorder.Result().Cookies(), 1)
	require.Equal(t, -1, recorder.Result().Cookies()[0].MaxAge)
}

func mustTestJWT(t *testing.T, secret string, expiresAt time.Time) string {
	t.Helper()

	token, err := SignJWT(Claims{
		UserID: "user-1",
		TeamID: "team-1",
		Role:   "OWNER",
		Email:  "owner@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}, secret, 15)
	require.NoError(t, err)

	return token
}
