package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"jobhub-ai/backend/internal/config"
	"jobhub-ai/backend/internal/handlers"
	"jobhub-ai/backend/internal/providers/telegram"
	"jobhub-ai/backend/internal/telegrampublisher"
)

func registerTelegramWebhook(router *gin.Engine, pool *pgxpool.Pool, logger *slog.Logger) error {
	secret := strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_SECRET"))
	if secret == "" {
		return nil
	}
	if len(secret) < 32 {
		return fmt.Errorf("TELEGRAM_WEBHOOK_SECRET must have at least 32 characters")
	}
	if os.Getenv("APP_ENV") == "production" && strings.TrimSpace(os.Getenv("TELEGRAM_PUBLISH_AFTER")) == "" {
		logger.Warn("Telegram webhook disabled until publication cutover is configured")
		return nil
	}
	cfg, err := config.LoadTelegramPublisher()
	if err != nil {
		return err
	}
	if cfg.DryRun || cfg.WithoutLink {
		return fmt.Errorf("Telegram webhook requires real delivery and public vacancy links")
	}
	client, err := telegram.NewClient(cfg.Token, &http.Client{Timeout: 15 * time.Second})
	if err != nil {
		return err
	}
	publisher := telegrampublisher.New(telegrampublisher.Config{
		ChannelID: cfg.ChannelID, ChannelUsername: cfg.ChannelUsername,
		PublicBaseURL: cfg.PublicBaseURL, BatchSize: 1, MaxAttempts: cfg.MaxAttempts,
	}, telegrampublisher.NewPostgresStore(pool, cfg.TestSources, cfg.PublishAfter), client, nil, logger)
	handlers.RegisterTelegramPublishRoute(router, secret, publisher, logger)
	return nil
}
