package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
	"smart-pet-monitoring/backend/internal/usecase"
)

type createFeedReadingRequest struct {
	DeviceID         string     `json:"device_id" binding:"required"`
	WeightGrams      *float64   `json:"weight_grams" binding:"required,gte=0"`
	FoodStockPercent *float64   `json:"food_stock_percent" binding:"omitempty,gte=0,lte=100"`
	WaterAvailable   *bool      `json:"water_available"`
	WaterStatus      string     `json:"water_status"`
	RecordedAt       *time.Time `json:"recorded_at"`
}

type updateDeviceStatusRequest struct {
	DeviceID         string   `json:"device_id" binding:"required"`
	FoodStockPercent *float64 `json:"food_stock_percent" binding:"omitempty,gte=0,lte=100"`
	WaterAvailable   *bool    `json:"water_available"`
	WaterStatus      string   `json:"water_status"`
}

func (h *Handler) createFeedReading(c *gin.Context) {
	var req createFeedReadingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "device_id and non-negative weight_grams are required")
		return
	}

	reading, err := h.feed.CreateReading(c.Request.Context(), usecase.CreateFeedReadingInput{
		DeviceID:         req.DeviceID,
		WeightGrams:      *req.WeightGrams,
		FoodStockPercent: req.FoodStockPercent,
		WaterAvailable:   req.WaterAvailable,
		WaterStatus:      req.WaterStatus,
		RecordedAt:       req.RecordedAt,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusCreated, feedReadingResponse(reading))
}

func (h *Handler) updateDeviceStatus(c *gin.Context) {
	var req updateDeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "device_id and at least one valid sensor field are required")
		return
	}

	status, err := h.feed.UpdateDeviceStatus(c.Request.Context(), usecase.UpdateDeviceStatusInput{
		DeviceID:         req.DeviceID,
		FoodStockPercent: req.FoodStockPercent,
		WaterAvailable:   req.WaterAvailable,
		WaterStatus:      req.WaterStatus,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, deviceStatusResponse(status))
}

func (h *Handler) getFeedHistory(c *gin.Context) {
	from, ok := parseOptionalTimeQuery(c, "from")
	if !ok {
		return
	}
	to, ok := parseOptionalTimeQuery(c, "to")
	if !ok {
		return
	}

	limit, ok := parseOptionalIntQuery(c, "limit")
	if !ok {
		return
	}

	readings, err := h.feed.ListHistory(c.Request.Context(), domain.FeedHistoryFilter{
		From:  from,
		To:    to,
		Limit: limit,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	items := make([]gin.H, 0, len(readings))
	for _, reading := range readings {
		items = append(items, feedReadingResponse(reading))
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func parseOptionalTimeQuery(c *gin.Context, key string) (*time.Time, bool) {
	value := c.Query(key)
	if value == "" {
		return nil, true
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		respondError(c, http.StatusBadRequest, key+" must use RFC3339 format")
		return nil, false
	}
	parsed = parsed.UTC()
	return &parsed, true
}

func parseOptionalIntQuery(c *gin.Context, key string) (int, bool) {
	value := c.Query(key)
	if value == "" {
		return 0, true
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		respondError(c, http.StatusBadRequest, key+" must be a positive integer")
		return 0, false
	}
	return parsed, true
}

func feedReadingResponse(reading domain.FeedReading) gin.H {
	return gin.H{
		"id":                 reading.ID,
		"device_id":          reading.DeviceID,
		"weight_grams":       reading.WeightGrams,
		"food_stock_percent": reading.FoodStockPercent,
		"water_available":    reading.WaterAvailable,
		"water_status":       reading.WaterStatus,
		"recorded_at":        reading.RecordedAt,
		"created_at":         reading.CreatedAt,
	}
}
