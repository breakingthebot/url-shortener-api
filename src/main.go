// src/main.go
// Bootstraps configuration, PostgreSQL connectivity, routing, and the HTTP server entrypoint.
// Connects to: src/config/config.go, src/services/link_repository_postgres.go, src/components/httpapi/link_handler.go
// Created: 2026-06-28

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/breakingthebot/url-shortener-api/src/components/httpapi"
	"github.com/breakingthebot/url-shortener-api/src/config"
	"github.com/breakingthebot/url-shortener-api/src/services"
	"github.com/breakingthebot/url-shortener-api/src/utils/shortcode"
	"github.com/jackc/pgx/v5/pgxpool"
)

const databaseStartupTimeout = 10 * time.Second

// main starts the URL shortener HTTP API.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	applicationConfig, err := config.LoadConfig()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseStartupTimeout)
	defer cancel()

	connectionPool, err := pgxpool.New(ctx, applicationConfig.DatabaseURL)
	if err != nil {
		logger.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer connectionPool.Close()

	repository := services.NewPostgresLinkRepository(connectionPool)
	if err := repository.EnsureSchema(ctx); err != nil {
		logger.Error("ensure database schema", "error", err)
		os.Exit(1)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(applicationConfig.ShortCodeLength),
		logger,
	)

	handler := httpapi.NewLinkHandler(service, logger)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	loggedHandler := httpapi.RequestLoggingMiddleware(logger, mux)

	server := &http.Server{
		Addr:              applicationConfig.Address(),
		Handler:           loggedHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting server", "address", applicationConfig.Address(), "environment", applicationConfig.AppEnv)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", fmt.Errorf("listen and serve: %w", err))
		os.Exit(1)
	}
}
