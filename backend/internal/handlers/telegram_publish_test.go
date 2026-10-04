package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"jobhub-ai/backend/internal/telegrampublisher"
)

type telegramRunnerStub struct {
	calls int
	err   error
}

func (s *telegramRunnerStub) ProcessOnce(context.Context) (telegrampublisher.Stats, error) {
	s.calls++
	return telegrampublisher.Stats{Checked: 1, Sent: 1}, s.err
}

func TestTelegramPublishRouteRequiresSecret(t *testing.T) {
	runner := &telegramRunnerStub{}
	router := gin.New()
	RegisterTelegramPublishRoute(router, "correct-secret", runner, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, secret := range []string{"", "wrong-secret"} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/internal/telegram-publish", nil)
		request.Header.Set("X-JobHub-Publisher-Secret", secret)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized || runner.calls != 0 {
			t.Fatalf("unauthorized request ran publisher: status=%d calls=%d", response.Code, runner.calls)
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/internal/telegram-publish", nil)
	request.Header.Set("X-JobHub-Publisher-Secret", "correct-secret")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || runner.calls != 1 || !strings.Contains(response.Body.String(), `"sent":1`) {
		t.Fatalf("authorized request failed: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTelegramPublishRouteDoesNotExposeErrors(t *testing.T) {
	runner := &telegramRunnerStub{err: errors.New("private database connection details")}
	router := gin.New()
	RegisterTelegramPublishRoute(router, "correct-secret", runner, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/internal/telegram-publish", nil)
	request.Header.Set("X-JobHub-Publisher-Secret", "correct-secret")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || strings.Contains(response.Body.String(), "private") {
		t.Fatalf("unexpected publisher error response: %d %s", response.Code, response.Body.String())
	}
}
