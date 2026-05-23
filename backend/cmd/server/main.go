package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tuankhanhvo/pulseops/internal/server"
	"github.com/tuankhanhvo/pulseops/pkg/config"
	"github.com/tuankhanhvo/pulseops/pkg/mongodb"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	logger, err := newLogger(cfg.Env)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("config loaded", zap.String("env", cfg.Env), zap.String("port", cfg.Port))

	db, err := mongodb.Connect(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		logger.Fatal("connect mongodb", zap.Error(err))
	}
	logger.Info("mongodb connected", zap.String("database", db.Name()))

	if err := mongodb.CreateIndexes(db); err != nil {
		logger.Fatal("create mongodb indexes", zap.Error(err))
	}
	logger.Info("mongodb indexes ready")

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      server.NewRouter(cfg, db, logger),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server starting", zap.String("addr", httpServer.Addr))
		serverErrors <- httpServer.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	case signal := <-shutdownSignals:
		logger.Info("shutdown signal received", zap.String("signal", signal.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("graceful shutdown failed", zap.Error(err))
	}

	mongodb.Disconnect(db.Client())
	logger.Info("server stopped")
}

func newLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}

	return zap.NewDevelopment()
}
