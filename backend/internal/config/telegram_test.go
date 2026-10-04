package config

import "testing"

func TestLoadTelegram(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret-value")
	t.Setenv("TELEGRAM_ALLOWED_CHANNEL_IDS", "-1001,-1002")
	t.Setenv("TELEGRAM_ALLOWED_ADMIN_IDS", "42,43")
	t.Setenv("TELEGRAM_OFFSET_FILE", t.TempDir()+"/offset")
	cfg, err := LoadTelegram()
	if err != nil || len(cfg.Allowed) != 2 || len(cfg.Admins) != 2 || cfg.Token != "secret-value" {
		t.Fatalf("unexpected config: %#v %v", cfg, err)
	}
}

func TestLoadTelegramRequiresTokenAndNumericChannels(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	if _, err := LoadTelegram(); err == nil {
		t.Fatal("expected missing token error")
	}
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret")
	t.Setenv("TELEGRAM_ALLOWED_CHANNEL_IDS", "public-name")
	if _, err := LoadTelegram(); err == nil {
		t.Fatal("expected invalid allowlist error")
	}
}

func TestValidateTelegramDevelopmentDatabase(t *testing.T) {
	good := "postgresql://user:password@127.0.0.1:5432/jobhub_telegram_poc_test?sslmode=disable"
	if err := ValidateTelegramDevelopmentDatabase("development", good); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"postgresql://host/jobhub_telegram_poc_test", "postgresql://127.0.0.1:5432/production", good + "&search_path=public"} {
		if ValidateTelegramDevelopmentDatabase("development", raw) == nil {
			t.Fatalf("accepted unsafe database %q", raw)
		}
	}
	if ValidateTelegramDevelopmentDatabase("production", good) == nil {
		t.Fatal("accepted production environment")
	}
}

func TestLoadTelegramPublisher(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret-value")
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_ID", "-10042")
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_USERNAME", "@jobhub_kz")
	t.Setenv("TELEGRAM_PUBLISHER_ENABLED", "true")
	t.Setenv("TELEGRAM_PUBLISH_INTERVAL", "30s")
	t.Setenv("JOBHUB_PUBLIC_BASE_URL", "https://jobhub.kz")
	cfg, err := LoadTelegramPublisher()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.ChannelID != -10042 || cfg.ChannelUsername != "jobhub_kz" || cfg.Interval.String() != "30s" || cfg.PublicBaseURL != "https://jobhub.kz" {
		t.Fatalf("unexpected publisher config: %#v", cfg)
	}
}

func TestLoadTelegramPublisherDryRunDoesNotRequireToken(t *testing.T) {
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_ID", "-10042")
	t.Setenv("TELEGRAM_PUBLISHER_DRY_RUN", "true")
	if _, err := LoadTelegramPublisher(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTelegramPublisherWithoutLink(t *testing.T) {
	t.Setenv("TELEGRAM_PUBLISHER_ENABLED", "true")
	t.Setenv("TELEGRAM_PUBLISHER_DRY_RUN", "false")
	t.Setenv("TELEGRAM_PUBLISH_WITHOUT_LINK", "true")
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret")
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_ID", "-10042")
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_USERNAME", "jobhub_test")
	t.Setenv("JOBHUB_PUBLIC_BASE_URL", "")
	cfg, err := LoadTelegramPublisher()
	if err != nil || !cfg.WithoutLink || cfg.PublicBaseURL != "" {
		t.Fatalf("unexpected linkless config: %#v %v", cfg, err)
	}
}

func TestLoadTelegramPublisherTestSourcesOnlyForMatchingStagingChannel(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("TELEGRAM_PUBLISHER_ENABLED", "true")
	t.Setenv("TELEGRAM_PUBLISH_WITHOUT_LINK", "true")
	t.Setenv("TELEGRAM_PUBLISH_TEST_SOURCES", "true")
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret")
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_ID", "-10042")
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_USERNAME", "jobhub_test")
	t.Setenv("TELEGRAM_PUBLISH_TEST_CHANNEL_ID", "-10042")
	cfg, err := LoadTelegramPublisher()
	if err != nil || !cfg.TestSources {
		t.Fatalf("expected staging test mode: %#v %v", cfg, err)
	}
	t.Setenv("APP_ENV", "production")
	if _, err := LoadTelegramPublisher(); err == nil {
		t.Fatal("test sources must not be enabled in production")
	}
	t.Setenv("APP_ENV", "staging")
	t.Setenv("TELEGRAM_PUBLISH_TEST_CHANNEL_ID", "-10043")
	if _, err := LoadTelegramPublisher(); err == nil {
		t.Fatal("test channel ID must match the publishing channel")
	}
}

func TestLoadTelegramPublisherValidation(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret")
	t.Setenv("TELEGRAM_PUBLIC_CHANNEL_USERNAME", "jobhub_kz")
	for name, env := range map[string]map[string]string{
		"positive channel": {"TELEGRAM_PUBLIC_CHANNEL_ID": "42"},
		"bad interval":     {"TELEGRAM_PUBLIC_CHANNEL_ID": "-10042", "TELEGRAM_PUBLISH_INTERVAL": "1s"},
		"bad base url":     {"TELEGRAM_PUBLIC_CHANNEL_ID": "-10042", "JOBHUB_PUBLIC_BASE_URL": "https://user@example.com"},
		"local live url":   {"TELEGRAM_PUBLIC_CHANNEL_ID": "-10042", "JOBHUB_PUBLIC_BASE_URL": "http://localhost:5173"},
		"missing username": {"TELEGRAM_PUBLIC_CHANNEL_ID": "-10042", "TELEGRAM_PUBLIC_CHANNEL_USERNAME": ""},
	} {
		t.Run(name, func(t *testing.T) {
			for key, value := range env {
				t.Setenv(key, value)
			}
			if _, err := LoadTelegramPublisher(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
