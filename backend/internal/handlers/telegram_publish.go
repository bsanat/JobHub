package handlers

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"jobhub-ai/backend/internal/telegrampublisher"
)

type TelegramPublishRunner interface {
	ProcessOnce(context.Context) (telegrampublisher.Stats, error)
}

func RegisterTelegramPublishRoute(router *gin.Engine, secret string, runner TelegramPublishRunner, logger *slog.Logger) {
	router.POST("/api/v1/internal/telegram-publish", func(c *gin.Context) {
		provided := c.GetHeader("X-JobHub-Publisher-Secret")
		if secret == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		defer cancel()
		stats, err := runner.ProcessOnce(ctx)
		if err != nil {
			logger.Error("telegram publisher request failed", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "publisher unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"checked": stats.Checked, "sent": stats.Sent, "failed": stats.Failed, "unknown": stats.Unknown})
	})
}
