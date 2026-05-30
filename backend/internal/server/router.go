package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/tuankhanhvo/pulseops/graph/generated"
	"github.com/tuankhanhvo/pulseops/internal/alerting"
	appgraph "github.com/tuankhanhvo/pulseops/internal/graph"
	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/internal/runbooks"
	"github.com/tuankhanhvo/pulseops/internal/streams"
	"github.com/tuankhanhvo/pulseops/internal/teams"
	"github.com/tuankhanhvo/pulseops/pkg/auth"
	"github.com/tuankhanhvo/pulseops/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func NewRouter(cfg *config.Config, db *mongo.Database, logger *zap.Logger, hub *streams.Hub) http.Handler {
	router := chi.NewRouter()
	incidentRepo := incidents.NewRepository(db)
	incidentService := incidents.NewService(incidentRepo)
	runbookRepo := runbooks.NewRepository(db)
	teamRepo := teams.NewRepository(db)

	graphQLServer := handler.New(
		generated.NewExecutableSchema(generated.Config{Resolvers: &appgraph.Resolver{
			DB:              db,
			Hub:             hub,
			IncidentRepo:    incidentRepo,
			IncidentService: incidentService,
			RunbookRepo:     runbookRepo,
			TeamRepo:        teamRepo,
		}}),
	)
	graphQLServer.AddTransport(transport.POST{})
	graphQLServer.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: allowedWebSocketOrigin(cfg.AllowedOrigins),
		},
	})

	router.Use(requestIDMiddleware)
	router.Use(recoveryMiddleware(logger))
	router.Use(requestLoggerMiddleware(logger))
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   splitOrigins(cfg.AllowedOrigins),
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type", "Authorization", requestIDHeader, "X-API-Key"},
		ExposedHeaders:   []string{requestIDHeader},
		AllowCredentials: true,
	}).Handler)
	router.Use(auth.Middleware(cfg.JWTSecret, logger))

	router.Get("/health", healthHandler)
	router.Get("/playground", playground.Handler("PulseOps GraphQL", "/query").ServeHTTP)
	router.Get("/query", graphQLServer.ServeHTTP)
	router.Post("/query", graphQLServer.ServeHTTP)

	NewAuthHandlers(*cfg, db, logger).RegisterRoutes(router)
	router.Post("/webhooks/alerts", alerting.NewWebhookHandler(db, logger))

	return router
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		w.Header().Set(requestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestLoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			logger.Info("http request",
				zap.String("requestId", requestIDFromContext(r.Context())),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", recorder.status),
				zap.Int64("duration_ms", time.Since(start).Milliseconds()),
				zap.String("userId", userIDFromContext(r.Context())),
				zap.String("teamId", teamIDFromContext(r.Context())),
			)
		})
	}
}

func recoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered",
						zap.Any("panic", recovered),
						zap.ByteString("stack", debug.Stack()),
						zap.String("requestId", requestIDFromContext(r.Context())),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}

	return hijacker.Hijack()
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func userIDFromContext(ctx context.Context) string {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return ""
	}

	return claims.UserID
}

func teamIDFromContext(ctx context.Context) string {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return ""
	}

	return claims.TeamID
}

func splitOrigins(origins string) []string {
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, origin := range parts {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			result = append(result, origin)
		}
	}

	return result
}

func allowedWebSocketOrigin(origins string) func(*http.Request) bool {
	allowed := make(map[string]struct{})
	for _, origin := range splitOrigins(origins) {
		allowed[origin] = struct{}{}
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		_, ok := allowed[origin]
		return ok
	}
}
