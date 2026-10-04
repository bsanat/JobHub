package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"jobhub-ai/backend/internal/config"
	"jobhub-ai/backend/internal/providers/telegram"
	"jobhub-ai/backend/internal/telegrampublisher"
)

func startTelegramPublisher(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	raw := strings.TrimSpace(os.Getenv("TELEGRAM_PUBLISHER_ENABLED"))
	if raw == "" {
		return nil
	}
	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		return fmt.Errorf("TELEGRAM_PUBLISHER_ENABLED must be true or false")
	}
	if !enabled {
		return nil
	}
	cfg, err := config.LoadTelegramPublisher()
	if err != nil {
		return err
	}
	var ready bool
	if err := pool.QueryRow(ctx, "SELECT to_regclass('jobhub.telegram_publications') IS NOT NULL").Scan(&ready); err != nil {
		return fmt.Errorf("check Telegram publisher migration: %w", err)
	}
	if !ready {
		return fmt.Errorf("Telegram publisher migration 000014 is not applied")
	}

	var client *telegram.Client
	if !cfg.DryRun {
		client, err = telegram.NewClient(cfg.Token, &http.Client{Timeout: 15 * time.Second})
		if err != nil {
			return err
		}
	}
	publisher := telegrampublisher.New(telegrampublisher.Config{
		ChannelID: cfg.ChannelID, ChannelUsername: cfg.ChannelUsername,
		PublicBaseURL: cfg.PublicBaseURL, DryRun: cfg.DryRun,
		BatchSize: cfg.BatchSize, MaxAttempts: cfg.MaxAttempts,
	}, telegrampublisher.NewPostgresStore(pool, cfg.TestSources), client, os.Stdout, logger)
	go func() {
		if client != nil {
			for {
				check, cancel := context.WithTimeout(ctx, 10*time.Second)
				err := client.VerifyChannel(check, cfg.ChannelID, cfg.ChannelUsername)
				cancel()
				if err == nil {
					break
				}
				logger.Error("Telegram channel verification failed; publisher waiting", "error", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(cfg.Interval):
				}
			}
		}
		logger.Info("Telegram publisher started", "interval", cfg.Interval.String(), "dry_run", cfg.DryRun)
		_ = publisher.Run(ctx, cfg.Interval)
	}()
	return nil
}
