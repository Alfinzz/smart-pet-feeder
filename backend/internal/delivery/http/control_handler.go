package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
	"smart-pet-monitoring/backend/internal/usecase"
)

type createManualCommandRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	Action   string `json:"action" binding:"required,oneof=feed drink"`
}

type updateCommandStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=completed failed"`
}

func (h *Handler) createManualCommand(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	var req createManualCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "device_id and action feed or drink are required")
		return
	}

	command, err := h.control.CreateManualCommand(c.Request.Context(), usecase.CreateManualCommandInput{
		OwnerID:  ownerID,
		DeviceID: req.DeviceID,
		Action:   domain.CommandAction(req.Action),
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusCreated, manualCommandResponse(command))
}

func (h *Handler) getNextDeviceCommand(c *gin.Context) {
	command, err := h.control.GetNextCommand(c.Request.Context(), c.Param("deviceID"))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusOK, gin.H{"data": nil})
			return
		}
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": manualCommandResponse(command)})
}

func (h *Handler) updateDeviceCommandStatus(c *gin.Context) {
	commandID, err := strconv.ParseInt(c.Param("commandID"), 10, 64)
	if err != nil || commandID <= 0 {
		respondError(c, http.StatusBadRequest, "command_id must be a positive integer")
		return
	}

	var req updateCommandStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "status completed or failed is required")
		return
	}

	command, err := h.control.UpdateCommandStatus(c.Request.Context(), usecase.UpdateCommandStatusInput{
		DeviceID:  c.Param("deviceID"),
		CommandID: commandID,
		Status:    domain.CommandStatus(req.Status),
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, manualCommandResponse(command))
}

func manualCommandResponse(command domain.ManualCommand) gin.H {
	return gin.H{
		"id":         command.ID,
		"owner_id":   command.OwnerID,
		"device_id":  command.DeviceID,
		"action":     command.Action,
		"status":     command.Status,
		"created_at": command.CreatedAt,
	}
}
