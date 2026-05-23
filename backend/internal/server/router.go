package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tuankhanhvo/pulseops/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func NewRouter(_ config.Config, _ *mongo.Database, _ *zap.Logger) http.Handler {
	router := chi.NewRouter()
	router.Get("/health", healthHandler)

	return router
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
