package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"github.com/tuankhanhvo/pulseops/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	stateCookieName   = "state"
	refreshCookieName = "refresh"
)

type AuthHandlers struct {
	cfg      config.Config
	db       *mongo.Database
	oauthCfg *oauth2.Config
	logger   *zap.Logger
	now      func() time.Time
}

type authUserDoc struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email         string             `bson:"email" json:"email"`
	Name          string             `bson:"name" json:"name"`
	AvatarURL     string             `bson:"avatarUrl,omitempty" json:"avatarUrl,omitempty"`
	TeamID        primitive.ObjectID `bson:"teamId" json:"teamId"`
	Role          string             `bson:"role" json:"role"`
	GoogleSubject string             `bson:"googleSubject" json:"googleSubject"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}

type authTeamDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"createdAt"`
	OwnerID   primitive.ObjectID `bson:"ownerId"`
}

type authSessionDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId"`
	TokenHash string             `bson:"tokenHash"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	CreatedAt time.Time          `bson:"createdAt"`
	UserAgent string             `bson:"userAgent"`
	IPAddress string             `bson:"ipAddress"`
}

type callbackRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"codeVerifier"`
	State        string `json:"state"`
}

func NewAuthHandlers(cfg config.Config, db *mongo.Database, loggers ...*zap.Logger) *AuthHandlers {
	logger := zap.NewNop()
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	}

	return &AuthHandlers{
		cfg:      cfg,
		db:       db,
		oauthCfg: auth.NewGoogleOAuthConfig(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.OAuthRedirectURL),
		logger:   logger,
		now:      time.Now,
	}
}

func (h *AuthHandlers) RegisterRoutes(router chi.Router) {
	router.Get("/auth/login", h.Login)
	router.Get("/auth/callback", h.Callback)
	router.Post("/auth/callback", h.Callback)
	router.Get("/auth/me", h.Me)
	router.Post("/auth/refresh", h.Refresh)
	router.Post("/auth/logout", h.Logout)
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GenerateStateToken()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, h.cookie(stateCookieName, state, "/auth/callback", 10*time.Minute))
	redirectURL := h.oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *AuthHandlers) Callback(w http.ResponseWriter, r *http.Request) {
	code, state, codeVerifier, err := h.callbackParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if code == "" || state == "" || codeVerifier == "" {
		http.Error(w, "missing oauth callback parameter", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		stateCookie, err := r.Cookie(stateCookieName)
		if err != nil || stateCookie.Value != state {
			http.Error(w, "invalid oauth state", http.StatusBadRequest)
			return
		}
		h.clearCookie(w, stateCookieName, "/auth/callback")
	}

	token, err := auth.ExchangeCode(r.Context(), h.oauthCfg, code, codeVerifier)
	if err != nil {
		h.logger.Error("oauth token exchange failed", zap.Error(err), zap.String("requestId", requestIDFromContext(r.Context())))
		http.Error(w, "failed to exchange oauth code", http.StatusBadGateway)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		http.Error(w, "missing id token", http.StatusBadGateway)
		return
	}

	googleClaims, err := auth.ValidateIDToken(r.Context(), rawIDToken, h.cfg.GoogleClientID)
	if err != nil {
		http.Error(w, "invalid id token", http.StatusUnauthorized)
		return
	}

	user, err := h.findOrCreateUser(r.Context(), googleClaims)
	if err != nil {
		http.Error(w, "failed to upsert user", http.StatusInternalServerError)
		return
	}
	if err := h.issueSessionCookies(w, r, user); err != nil {
		http.Error(w, "failed to issue session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, h.frontendCallbackURL(), http.StatusFound)
}

func (h *AuthHandlers) callbackParams(r *http.Request) (string, string, string, error) {
	if r.Method == http.MethodPost {
		defer r.Body.Close()
		var request callbackRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return "", "", "", errors.New("invalid oauth callback body")
		}

		return request.Code, request.State, request.CodeVerifier, nil
	}

	return r.URL.Query().Get("code"), r.URL.Query().Get("state"), r.URL.Query().Get("code_verifier"), nil
}

func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.RequireAuth(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.findUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessions := h.db.Collection("sessions")
	var session authSessionDoc
	err = sessions.FindOne(r.Context(), bson.M{"tokenHash": hashToken(cookie.Value)}).Decode(&session)
	if err != nil || !session.ExpiresAt.After(h.now()) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_, _ = sessions.DeleteOne(r.Context(), bson.M{"_id": session.ID})

	user, err := h.findUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.issueSessionCookies(w, r, user); err != nil {
		http.Error(w, "failed to refresh session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.RequireAuth(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err == nil {
		_, _ = h.db.Collection("sessions").DeleteMany(r.Context(), bson.M{"userId": userID})
	}

	h.clearCookie(w, auth.SessionCookieName, "/")
	h.clearCookie(w, refreshCookieName, "/auth/refresh")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) findOrCreateUser(ctx context.Context, claims *auth.GoogleClaims) (authUserDoc, error) {
	users := h.db.Collection("users")

	var user authUserDoc
	err := users.FindOne(ctx, bson.M{"googleSubject": claims.Subject}).Decode(&user)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return authUserDoc{}, err
	}

	now := h.now()
	userID := primitive.NewObjectID()
	teamID := primitive.NewObjectID()
	user = authUserDoc{
		ID:            userID,
		Email:         claims.Email,
		Name:          claims.Name,
		AvatarURL:     claims.Picture,
		TeamID:        teamID,
		Role:          "OWNER",
		GoogleSubject: claims.Subject,
		CreatedAt:     now,
	}

	if _, err := users.InsertOne(ctx, user); err != nil {
		return authUserDoc{}, err
	}
	if _, err := h.db.Collection("teams").InsertOne(ctx, authTeamDoc{
		ID:        teamID,
		Name:      teamNameForUser(claims),
		CreatedAt: now,
		OwnerID:   userID,
	}); err != nil {
		return authUserDoc{}, err
	}

	return user, nil
}

func (h *AuthHandlers) findUserByID(ctx context.Context, userID primitive.ObjectID) (authUserDoc, error) {
	var user authUserDoc
	err := h.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	return user, err
}

func (h *AuthHandlers) issueSessionCookies(w http.ResponseWriter, r *http.Request, user authUserDoc) error {
	sessionToken, err := auth.SignJWT(auth.Claims{
		UserID: user.ID.Hex(),
		TeamID: user.TeamID.Hex(),
		Role:   user.Role,
		Email:  user.Email,
	}, h.cfg.JWTSecret, h.cfg.JWTExpiryMinutes)
	if err != nil {
		return err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return err
	}

	now := h.now()
	_, err = h.db.Collection("sessions").InsertOne(r.Context(), authSessionDoc{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: now.Add(time.Duration(h.cfg.RefreshTokenExpiryDays) * 24 * time.Hour),
		CreatedAt: now,
		UserAgent: r.UserAgent(),
		IPAddress: clientIP(r),
	})
	if err != nil {
		return err
	}

	http.SetCookie(w, h.cookie(auth.SessionCookieName, sessionToken, "/", time.Duration(h.cfg.JWTExpiryMinutes)*time.Minute))
	http.SetCookie(w, h.cookie(refreshCookieName, refreshToken, "/auth/refresh", time.Duration(h.cfg.RefreshTokenExpiryDays)*24*time.Hour))
	return nil
}

func (h *AuthHandlers) cookie(name, value, path string, ttl time.Duration) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if h.cfg.Env == "production" {
		sameSite = http.SameSiteStrictMode
	}

	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   int(ttl.Seconds()),
		Expires:  h.now().Add(ttl),
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: sameSite,
	}
}

func (h *AuthHandlers) clearCookie(w http.ResponseWriter, name, path string) {
	cookie := h.cookie(name, "", path, -time.Hour)
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

func (h *AuthHandlers) frontendCallbackURL() string {
	origin := "http://localhost:5173"
	if h.cfg.AllowedOrigins != "" {
		origin = strings.TrimSpace(strings.Split(h.cfg.AllowedOrigins, ",")[0])
	}

	return strings.TrimRight(origin, "/") + "/auth/callback"
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func generateRefreshToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func teamNameForUser(claims *auth.GoogleClaims) string {
	if claims.Name != "" {
		return claims.Name + "'s Team"
	}

	return claims.Email + "'s Team"
}
