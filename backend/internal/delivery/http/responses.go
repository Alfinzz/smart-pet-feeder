package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, errorResponse{Error: message})
}

func respondUsecaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrConflict):
		respondError(c, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		respondError(c, http.StatusUnauthorized, "invalid email or password")
	case errors.Is(err, domain.ErrUnauthorized):
		respondError(c, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, domain.ErrValidation):
		respondError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		respondError(c, http.StatusNotFound, "resource not found")
	default:
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}
