package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	authTokenMiddleware "github.com/eslupmi101/avito_merch_store/api/middleware"
	"github.com/eslupmi101/avito_merch_store/api/route"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.NewConfig()
	logger := setupLogger(cfg.Env)
	slog.SetDefault(logger)

	logger.Info("Starting merch store api", slog.String("env", cfg.Env))
	logger.Debug("Debug messages are enabled")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout*time.Second)
	defer cancel()

	connStr, err := cfg.BuildPGConnString()
	if err != nil {
		log.Fatalf("Error building connection to database string: %v", err)
	}
	db := config.NewPostgresDb(ctx, connStr)
	defer db.Close()

	router := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Origin", "X-Requested-With"},
		ExposedHeaders:   []string{"Link", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))

	router.Use(cors.Handler)

	route.Setup(&cfg, cfg.HTTPServer.Timeout, db, router)

	http.ListenAndServe(cfg.HTTPServer.Address, router)
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		)
	default:
		log.Fatalf("Invalid env provided: %s", env)
	}

	return logger
}
