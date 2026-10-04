package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"jobhub-ai/backend/internal/auth"
)

func RegisterOnboardingRoutes(group *gin.RouterGroup, h authHandler, service auth.OnboardingService) {
	authenticated := func(c *gin.Context) {
		token, _ := c.Cookie(auth.SessionCookieName)
		session, err := h.service.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHENTICATED", "message": "Authentication required"}})
			return
		}
		c.Set("session", session)
	}
	respond := func(c *gin.Context, state auth.Onboarding, err error) {
		switch {
		case err == nil:
			c.JSON(http.StatusOK, state)
		case errors.Is(err, auth.ErrOnboardingValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid onboarding answers"}})
		case errors.Is(err, auth.ErrOnboardingConflict):
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "ONBOARDING_CONFLICT", "message": "Reload the latest onboarding state"}})
		default:
			h.logger.Error("onboarding request failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "ONBOARDING_UNAVAILABLE", "message": "Onboarding is temporarily unavailable"}})
		}
	}
	group.GET("/onboarding", authenticated, func(c *gin.Context) {
		state, err := service.GetOnboarding(c.Request.Context(), c.MustGet("session").(auth.Session).User.ID)
		respond(c, state, err)
	})
	group.PUT("/onboarding", authenticated, h.requireTrustedOrigin(), h.requireCSRF(), func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8*1024)
		var input auth.OnboardingInput
		if !decodeJSON(c, &input) {
			return
		}
		state, err := service.SaveOnboarding(c.Request.Context(), c.MustGet("session").(auth.Session).User.ID, input)
		respond(c, state, err)
	})
}
