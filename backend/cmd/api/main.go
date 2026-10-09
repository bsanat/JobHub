package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"jobhub-ai/backend/internal/auth"
	"jobhub-ai/backend/internal/candidates"
	"jobhub-ai/backend/internal/companies"
	"jobhub-ai/backend/internal/config"
	"jobhub-ai/backend/internal/database"
	"jobhub-ai/backend/internal/handlers"
	"jobhub-ai/backend/internal/jobs"
	"jobhub-ai/backend/internal/services"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 10*time.Second)
	pool, err := database.Connect(startup, cfg.DatabaseURL, cfg.PGXQueryExecMode)
	cancel()
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := startTelegramPublisher(ctx, pool, logger); err != nil {
		logger.Error("Telegram publisher disabled", "error", err)
	}
	jobStore := jobs.NewPostgresStore(pool, cfg.Environment == "production")
	authService := auth.NewService(pool)
	if os.Getenv("NEON_AUTH_BASE_URL") != "" {
		oauthVerifier, err := auth.NewNeonOAuthVerifier(ctx, os.Getenv("NEON_AUTH_BASE_URL"), pool)
		if err != nil {
			return err
		}
		authService.SetOAuthVerifier(oauthVerifier)
		authService.SetEmailVerifier(oauthVerifier)
	}
	var resumeFiles candidates.FileStore
	if os.Getenv("AWS_ENDPOINT_URL_S3") != "" {
		resumeFiles, err = candidates.NewS3FileStore(os.Getenv("AWS_ENDPOINT_URL_S3"), os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "uploads", nil)
		if err != nil {
			return err
		}
	} else if os.Getenv("VERCEL") == "1" {
		resumeFiles = candidates.UnavailableFileStore{}
	} else {
		resumeFiles, err = candidates.NewLocalFileStore(cfg.ResumeStorageDir)
		if err != nil {
			return err
		}
	}
	candidateService := candidates.NewService(candidates.NewPostgresStore(pool), resumeFiles)
	companyStore := companies.NewPostgresStore(pool, cfg.Environment == "production")
	router := handlers.NewRouter(services.NewHealthService(pool), logger, cfg.FrontendOrigin, jobStore, authService, candidateService, companyStore, cfg.Environment)
	if err := registerTelegramWebhook(router, pool, logger); err != nil {
		logger.Error("Telegram webhook disabled", "error", err)
	}
	if err := router.SetTrustedProxies(nil); err != nil {
		return err
	}
	// Profile updates can span several remote Postgres round trips. Keep the API and
	// browser timeouts aligned so a committed request is not reported as failed.
	server := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	failures := make(chan error, 1)
	go func() { failures <- server.ListenAndServe() }()
	logger.Info("API started", "port", cfg.Port, "environment", cfg.Environment)
	select {
	case err := <-failures:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
	}
	shutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	return server.Shutdown(shutdown)
}
