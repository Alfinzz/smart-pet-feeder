package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/usecase"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "name, valid email, and password with at least 8 characters are required")
		return
	}

	output, err := h.auth.Register(c.Request.Context(), usecase.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}
	if err := h.dashboard.EnsureDefaultProfile(c.Request.Context(), output.Owner.ID); err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusCreated, authResponse(output))
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "email and password are required")
		return
	}

	output, err := h.auth.Login(c.Request.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse(output))
}

func authResponse(output usecase.LoginOutput) gin.H {
	return gin.H{
		"token":      output.Token,
		"expires_at": output.ExpiresAt,
		"owner": gin.H{
			"id":    output.Owner.ID,
			"name":  output.Owner.Name,
			"email": output.Owner.Email,
		},
	}
}
