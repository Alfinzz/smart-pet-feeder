package http

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
)

const ownerIDContextKey = "ownerID"

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	if strings.TrimSpace(allowOrigin) == "" {
		allowOrigin = "*"
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Device-Key")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Max-Age", "86400")
		c.Next()
	}
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if len(authHeader) < 8 || !strings.EqualFold(authHeader[:7], "Bearer ") {
			respondError(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(authHeader[7:])
		if tokenString == "" {
			respondError(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		claims, err := h.jwt.VerifyOwnerToken(tokenString)
		if err != nil {
			respondUsecaseError(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set(ownerIDContextKey, claims.OwnerID)
		c.Next()
	}
}

func (h *Handler) deviceAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(c.GetHeader("X-Device-Key"))
		expected := strings.TrimSpace(h.deviceAPIKey)
		if apiKey == "" || expected == "" || subtle.ConstantTimeCompare([]byte(apiKey), []byte(expected)) != 1 {
			respondError(c, http.StatusUnauthorized, "invalid device api key")
			c.Abort()
			return
		}
		c.Next()
	}
}

func ownerIDFromContext(c *gin.Context) (int64, bool) {
	value, ok := c.Get(ownerIDContextKey)
	if !ok {
		return 0, false
	}
	ownerID, ok := value.(int64)
	return ownerID, ok
}
