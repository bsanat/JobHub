package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProductionWebhookWithoutCutoverDoesNotBlockAPI(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", strings.Repeat("x", 32))
	t.Setenv("TELEGRAM_PUBLISH_AFTER", "")
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := registerTelegramWebhook(router, nil, logger); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/internal/telegram-publish", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("publisher route should be disabled without cutover: %d", response.Code)
	}
}
