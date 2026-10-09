package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"jobhub-ai/backend/internal/config"
	"jobhub-ai/backend/internal/database"
	"jobhub-ai/backend/internal/providers/telegram"
	"jobhub-ai/backend/internal/telegrampublisher"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("Telegram publisher stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	once := flag.Bool("once", false, "publish one batch and exit")
	dryRun := flag.Bool("dry-run", false, "print messages without sending or marking publications")
	flag.Parse()
	if *dryRun {
		_ = os.Setenv("TELEGRAM_PUBLISHER_DRY_RUN", "true")
	}

	appConfig, err := config.Load()
	if err != nil {
		return err
	}
	publisherConfig, err := config.LoadTelegramPublisher()
	if err != nil {
		return err
	}
	if publisherConfig.DryRun {
		if err := config.ValidateTelegramDevelopmentDatabase(appConfig.Environment, appConfig.DatabaseURL); err != nil {
			return err
		}
	}
	if !publisherConfig.Enabled && !publisherConfig.DryRun {
		return errors.New("set TELEGRAM_PUBLISHER_ENABLED=true or pass --dry-run")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 10*time.Second)
	pool, err := database.Connect(startup, appConfig.DatabaseURL, appConfig.PGXQueryExecMode)
	cancel()
	if err != nil {
		return err
	}
	defer pool.Close()

	var sender telegrampublisher.Sender
	if !publisherConfig.DryRun {
		client, clientErr := telegram.NewClient(publisherConfig.Token, &http.Client{Timeout: 15 * time.Second})
		if clientErr != nil {
			return clientErr
		}
		check, cancelCheck := context.WithTimeout(ctx, 10*time.Second)
		err = client.VerifyChannel(check, publisherConfig.ChannelID, publisherConfig.ChannelUsername)
		cancelCheck()
		if err != nil {
			return err
		}
		sender = client
	}
	publisher := telegrampublisher.New(telegrampublisher.Config{
		ChannelID:       publisherConfig.ChannelID,
		ChannelUsername: publisherConfig.ChannelUsername,
		PublicBaseURL:   publisherConfig.PublicBaseURL,
		DryRun:          publisherConfig.DryRun,
		BatchSize:       publisherConfig.BatchSize,
		MaxAttempts:     publisherConfig.MaxAttempts,
	}, telegrampublisher.NewPostgresStore(pool, publisherConfig.TestSources, publisherConfig.PublishAfter), sender, os.Stdout, logger)

	if *once {
		stats, err := publisher.ProcessOnce(ctx)
		if err != nil {
			return err
		}
		logger.Info("Telegram publisher batch finished", "checked", stats.Checked, "sent", stats.Sent, "failed", stats.Failed, "unknown", stats.Unknown, "dry_run", stats.DryRun)
		if err := logKZCompanyPublicationStatus(ctx, pool, publisherConfig.ChannelID, logger); err != nil {
			logger.Warn("Kazakhstan company publication status unavailable", "error", err)
		}
		return nil
	}
	logger.Info("Telegram publisher started", "interval", publisherConfig.Interval.String(), "dry_run", publisherConfig.DryRun)
	return publisher.Run(ctx, publisherConfig.Interval)
}

func logKZCompanyPublicationStatus(ctx context.Context, pool *pgxpool.Pool, channelID int64, logger *slog.Logger) error {
	check, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := pool.Query(check, `SELECT j.source, COALESCE(tp.status, 'unrecorded'), count(*)
FROM jobhub.jobs j
JOIN jobhub.job_sources s ON s.source = j.source
LEFT JOIN jobhub.telegram_publications tp ON tp.job_id = j.id AND tp.channel_id = $1
WHERE s.provider = 'kzcompany' AND s.enabled AND s.production_permissions_confirmed
  AND jobhub.job_is_public(j)
GROUP BY j.source, COALESCE(tp.status, 'unrecorded')
ORDER BY j.source, COALESCE(tp.status, 'unrecorded')`, channelID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var source, status string
		var count int64
		if err := rows.Scan(&source, &status, &count); err != nil {
			return err
		}
		logger.Info("Kazakhstan company Telegram publication status", "source", source, "status", status, "count", count)
	}
	return rows.Err()
}
