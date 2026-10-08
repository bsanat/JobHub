package telegrampublisher

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"
)

type Config struct {
	ChannelID       int64
	ChannelUsername string
	PublicBaseURL   string
	DryRun          bool
	BatchSize       int
	MaxAttempts     int
}

type Vacancy struct {
	ID              string
	Source          string
	SourceName      string
	Attribution     string
	OriginalURL     string
	Title           string
	Description     string
	CompanyName     string
	Location        string
	WorkMode        string
	ExperienceLevel string
	EmploymentType  string
	SalaryRaw       string
	SalaryMin       string
	SalaryMax       string
	SalaryCurrency  string
	SalaryPeriod    string
	SalaryVisible   bool
}

type Store interface {
	ClaimBatch(context.Context, int64, int, int) ([]Vacancy, error)
	ListEligible(context.Context, int64, int, int) ([]Vacancy, error)
	MarkSent(context.Context, string, int64, int64) error
	MarkFailed(context.Context, string, int64, string, bool) error
}

type Sender interface {
	SendHTMLMessageAndReturnID(context.Context, int64, string) (int64, error)
}

type Publisher struct {
	cfg    Config
	store  Store
	sender Sender
	out    io.Writer
	logger *slog.Logger
}

type Stats struct {
	Checked    int
	Sent       int
	Failed     int
	Unknown    int
	DryRun     int
	RetryAfter time.Duration
}

type EligibilityCandidate struct {
	Source            string
	SourceApproved    bool
	TestSources       bool
	DemoSource        bool
	PublicationStatus string
	ModerationStatus  string
	PublicVisible     bool
	AlreadySent       bool
	Attempts          int
	MaxAttempts       int
}

func EligibleForPublication(c EligibilityCandidate) bool {
	return (c.SourceApproved || c.TestSources) &&
		!c.DemoSource &&
		(c.Source != "jobhub" || c.PublicationStatus == "published") &&
		c.ModerationStatus == "approved" &&
		c.PublicVisible &&
		!c.AlreadySent &&
		c.Attempts < c.MaxAttempts
}

func New(cfg Config, store Store, sender Sender, out io.Writer, logger *slog.Logger) *Publisher {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Publisher{cfg: cfg, store: store, sender: sender, out: out, logger: logger}
}

func (p *Publisher) ProcessOnce(ctx context.Context) (Stats, error) {
	if p.cfg.DryRun {
		items, err := p.store.ListEligible(ctx, p.cfg.ChannelID, p.cfg.BatchSize, p.cfg.MaxAttempts)
		if err != nil {
			return Stats{}, err
		}
		stats := Stats{Checked: len(items)}
		for _, item := range items {
			stats.DryRun++
			if p.out != nil {
				_, _ = p.out.Write([]byte(FormatVacancy(item, p.cfg.PublicBaseURL) + "\n\n"))
			}
		}
		return stats, nil
	}

	stats := Stats{}
	for n := 0; n < p.cfg.BatchSize; n++ {
		if n > 0 {
			timer := time.NewTimer(1100 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return stats, ctx.Err()
			case <-timer.C:
			}
		}
		items, err := p.store.ClaimBatch(ctx, p.cfg.ChannelID, 1, p.cfg.MaxAttempts)
		if err != nil {
			return stats, err
		}
		if len(items) == 0 {
			break
		}
		item := items[0]
		stats.Checked++
		message := FormatVacancy(item, p.cfg.PublicBaseURL)
		messageID, err := p.sender.SendHTMLMessageAndReturnID(ctx, p.cfg.ChannelID, message)
		if err != nil {
			stats.Failed++
			code := errorCode(err)
			var rejected interface{ DeliveryKnownFailed() bool }
			if !errors.As(err, &rejected) || !rejected.DeliveryKnownFailed() {
				stats.Unknown++
				p.logger.Error("telegram publication outcome unknown; claim needs manual reconciliation", "job_id", item.ID, "error_code", code)
				continue
			}
			countAttempt := true
			var attemptPolicy interface{ CountFailureAttempt() bool }
			if errors.As(err, &attemptPolicy) {
				countAttempt = attemptPolicy.CountFailureAttempt()
			}
			if markErr := p.store.MarkFailed(ctx, item.ID, p.cfg.ChannelID, code, countAttempt); markErr != nil {
				return stats, markErr
			}
			p.logger.Warn("telegram publication failed", "job_id", item.ID, "error_code", code)
			var batchFailure interface {
				StopBatch() bool
				RetryDelay() time.Duration
			}
			if errors.As(err, &batchFailure) && batchFailure.StopBatch() {
				stats.RetryAfter = batchFailure.RetryDelay()
				break
			}
			continue
		}
		if err := p.store.MarkSent(ctx, item.ID, p.cfg.ChannelID, messageID); err != nil {
			return stats, err
		}
		stats.Sent++
		p.logger.Info("telegram publication sent", "job_id", item.ID, "message_id", messageID)
	}
	return stats, nil
}

func (p *Publisher) Run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = time.Minute
	}
	for {
		stats, err := p.ProcessOnce(ctx)
		if err != nil {
			p.logger.Error("telegram publisher iteration failed", "error", err)
		}
		wait := interval
		if stats.RetryAfter > wait {
			wait = stats.RetryAfter
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func errorCode(err error) string {
	code := "TELEGRAM_SEND_FAILED"
	if err != nil && err.Error() != "" {
		code = err.Error()
	}
	if len(code) > 120 {
		return code[:120]
	}
	return code
}
