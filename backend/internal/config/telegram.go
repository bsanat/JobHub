package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var telegramUsernamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{4,31}$`)

type TelegramConfig struct {
	Token       string
	Allowed     map[int64]struct{}
	Admins      map[int64]struct{}
	OffsetFile  string
	DatabaseURL string
}

type TelegramPublisherConfig struct {
	Enabled         bool
	DryRun          bool
	WithoutLink     bool
	TestSources     bool
	Token           string
	ChannelID       int64
	ChannelUsername string
	PublicBaseURL   string
	Interval        time.Duration
	BatchSize       int
	MaxAttempts     int
}

func LoadTelegram() (TelegramConfig, error) {
	c := TelegramConfig{
		Token: strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")), OffsetFile: value("TELEGRAM_OFFSET_FILE", "./storage/telegram-collector.offset"),
		DatabaseURL: strings.TrimSpace(os.Getenv("TELEGRAM_DEV_DATABASE_URL")), Allowed: map[int64]struct{}{}, Admins: map[int64]struct{}{},
	}
	if c.Token == "" {
		return TelegramConfig{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if strings.TrimSpace(c.OffsetFile) == "" || filepath.Clean(c.OffsetFile) == "." {
		return TelegramConfig{}, errors.New("TELEGRAM_OFFSET_FILE must name a file")
	}
	raw := strings.TrimSpace(os.Getenv("TELEGRAM_ALLOWED_CHANNEL_IDS"))
	parts := []string{}
	if raw != "" {
		parts = strings.Split(raw, ",")
	}
	if len(parts) > 20 {
		return TelegramConfig{}, errors.New("TELEGRAM_ALLOWED_CHANNEL_IDS exceeds 20 channels")
	}
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id >= 0 {
			return TelegramConfig{}, errors.New("TELEGRAM_ALLOWED_CHANNEL_IDS must contain negative numeric channel IDs")
		}
		if _, duplicate := c.Allowed[id]; duplicate {
			return TelegramConfig{}, errors.New("TELEGRAM_ALLOWED_CHANNEL_IDS contains a duplicate")
		}
		c.Allowed[id] = struct{}{}
	}
	adminRaw := strings.TrimSpace(os.Getenv("TELEGRAM_ALLOWED_ADMIN_IDS"))
	if adminRaw == "" {
		return c, nil
	}
	adminParts := strings.Split(adminRaw, ",")
	if len(adminParts) > 20 {
		return TelegramConfig{}, errors.New("TELEGRAM_ALLOWED_ADMIN_IDS exceeds 20 users")
	}
	for _, part := range adminParts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id <= 0 {
			return TelegramConfig{}, errors.New("TELEGRAM_ALLOWED_ADMIN_IDS must contain positive numeric user IDs")
		}
		if _, duplicate := c.Admins[id]; duplicate {
			return TelegramConfig{}, errors.New("TELEGRAM_ALLOWED_ADMIN_IDS contains a duplicate")
		}
		c.Admins[id] = struct{}{}
	}
	return c, nil
}

func ValidateTelegramDevelopmentDatabase(environment, raw string) error {
	return ValidateLocalPOCDatabase(environment, raw)
}

func LoadTelegramPublisher() (TelegramPublisherConfig, error) {
	enabled, err := optionalBool("TELEGRAM_PUBLISHER_ENABLED", false)
	if err != nil {
		return TelegramPublisherConfig{}, err
	}
	dryRun, err := optionalBool("TELEGRAM_PUBLISHER_DRY_RUN", false)
	if err != nil {
		return TelegramPublisherConfig{}, err
	}
	withoutLink, err := optionalBool("TELEGRAM_PUBLISH_WITHOUT_LINK", false)
	if err != nil {
		return TelegramPublisherConfig{}, err
	}
	testSources, err := optionalBool("TELEGRAM_PUBLISH_TEST_SOURCES", false)
	if err != nil {
		return TelegramPublisherConfig{}, err
	}
	interval, err := time.ParseDuration(value("TELEGRAM_PUBLISH_INTERVAL", "60s"))
	if err != nil || interval < 5*time.Second || interval > 24*time.Hour {
		return TelegramPublisherConfig{}, fmt.Errorf("TELEGRAM_PUBLISH_INTERVAL must be between 5s and 24h")
	}
	batchSize, err := optionalInt("TELEGRAM_PUBLISH_BATCH_SIZE", 10, 1, 50)
	if err != nil {
		return TelegramPublisherConfig{}, err
	}
	maxAttempts, err := optionalInt("TELEGRAM_PUBLISH_MAX_ATTEMPTS", 3, 1, 10)
	if err != nil {
		return TelegramPublisherConfig{}, err
	}
	channelID, err := strconv.ParseInt(strings.TrimSpace(os.Getenv("TELEGRAM_PUBLIC_CHANNEL_ID")), 10, 64)
	if err != nil || channelID >= 0 {
		return TelegramPublisherConfig{}, fmt.Errorf("TELEGRAM_PUBLIC_CHANNEL_ID must be a negative numeric channel ID")
	}
	if testSources {
		testChannelID, parseErr := strconv.ParseInt(strings.TrimSpace(os.Getenv("TELEGRAM_PUBLISH_TEST_CHANNEL_ID")), 10, 64)
		if strings.TrimSpace(os.Getenv("APP_ENV")) != "staging" || parseErr != nil || testChannelID != channelID {
			return TelegramPublisherConfig{}, fmt.Errorf("TELEGRAM_PUBLISH_TEST_SOURCES requires APP_ENV=staging and a matching TELEGRAM_PUBLISH_TEST_CHANNEL_ID")
		}
	}
	username := strings.TrimPrefix(strings.TrimSpace(os.Getenv("TELEGRAM_PUBLIC_CHANNEL_USERNAME")), "@")
	if username != "" && !telegramUsernamePattern.MatchString(username) {
		return TelegramPublisherConfig{}, fmt.Errorf("TELEGRAM_PUBLIC_CHANNEL_USERNAME is invalid")
	}
	if !dryRun && username == "" {
		return TelegramPublisherConfig{}, fmt.Errorf("TELEGRAM_PUBLIC_CHANNEL_USERNAME is required for publishing")
	}
	publicBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("JOBHUB_PUBLIC_BASE_URL")), "/")
	if withoutLink {
		publicBaseURL = ""
	}
	if publicBaseURL == "" && dryRun && !withoutLink {
		publicBaseURL = "http://localhost:5173"
	}
	if publicBaseURL == "" && !withoutLink {
		return TelegramPublisherConfig{}, fmt.Errorf("JOBHUB_PUBLIC_BASE_URL is required for publishing")
	}
	if publicBaseURL != "" {
		if err := validatePublicBaseURL(publicBaseURL); err != nil {
			return TelegramPublisherConfig{}, err
		}
	}
	if !dryRun && !withoutLink {
		parsed, _ := url.Parse(publicBaseURL)
		host := strings.ToLower(parsed.Hostname())
		ip := net.ParseIP(host)
		if parsed.Scheme != "https" || host == "localhost" || strings.HasSuffix(host, ".local") || ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified()) {
			return TelegramPublisherConfig{}, fmt.Errorf("JOBHUB_PUBLIC_BASE_URL must be a public HTTPS URL for publishing")
		}
	}
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if !dryRun && token == "" {
		return TelegramPublisherConfig{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required unless TELEGRAM_PUBLISHER_DRY_RUN=true or --dry-run is used")
	}
	return TelegramPublisherConfig{
		Enabled: enabled, DryRun: dryRun, WithoutLink: withoutLink, TestSources: testSources, Token: token, ChannelID: channelID, ChannelUsername: username,
		PublicBaseURL: publicBaseURL, Interval: interval, BatchSize: batchSize, MaxAttempts: maxAttempts,
	}, nil
}

func optionalBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}
	return value, nil
}

func optionalInt(key string, fallback, min, max int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be between %d and %d", key, min, max)
	}
	return value, nil
}

func validatePublicBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return fmt.Errorf("JOBHUB_PUBLIC_BASE_URL must be an HTTP(S) URL without credentials, query, or fragment")
	}
	return nil
}
