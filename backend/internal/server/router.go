package server

import (
	"encoding/json"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/tuankhanhvo/pulseops/graph/generated"
	appgraph "github.com/tuankhanhvo/pulseops/internal/graph"
	"github.com/tuankhanhvo/pulseops/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func NewRouter(_ config.Config, _ *mongo.Database, _ *zap.Logger) http.Handler {
	router := chi.NewRouter()
	graphQLServer := handler.NewDefaultServer(
		generated.NewExecutableSchema(generated.Config{Resolvers: &appgraph.Resolver{}}),
	)

	router.Get("/health", healthHandler)
	router.Get("/query", playground.Handler("PulseOps GraphQL", "/query").ServeHTTP)
	router.Post("/query", graphQLServer.ServeHTTP)

	return router
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
