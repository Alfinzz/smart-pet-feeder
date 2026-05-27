package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
)

type petDetailsRequest struct {
	DeviceID             string  `json:"device_id"`
	Name                 string  `json:"name" binding:"required"`
	Species              string  `json:"species" binding:"required"`
	Breed                string  `json:"breed" binding:"required"`
	Gender               string  `json:"gender"`
	BirthDate            string  `json:"birth_date"`
	DailyFeedTargetGrams float64 `json:"daily_feed_target_grams" binding:"required,gt=0"`
}

type deviceSettingsRequest struct {
	Name                   string  `json:"name"`
	ManualFeedPortionGrams float64 `json:"manual_feed_portion_grams" binding:"gte=0"`
	ServoOpenDegrees       *int    `json:"servo_open_degrees" binding:"omitempty,gte=0,lte=180"`
	ServoClosedDegrees     *int    `json:"servo_closed_degrees" binding:"omitempty,gte=0,lte=180"`
	AutomationEnabled      *bool   `json:"automation_enabled"`
}

type notificationPreferencesRequest struct {
	LowFoodAlert         bool `json:"low_food_alert"`
	EmptyWaterAlert      bool `json:"empty_water_alert"`
	FeedingSuccessReport bool `json:"feeding_success_report"`
}

type notificationPreferencesPatchRequest struct {
	AlertLowFood         *bool `json:"alert_low_food"`
	AlertEmptyWater      *bool `json:"alert_empty_water"`
	AlertFeedSuccess     *bool `json:"alert_feed_success"`
	LowFoodAlert         *bool `json:"low_food_alert"`
	EmptyWaterAlert      *bool `json:"empty_water_alert"`
	FeedingSuccessReport *bool `json:"feeding_success_report"`
}

func (h *Handler) getPetDetails(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	pet, err := h.profile.GetPetDetails(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, petDetailsResponse(pet))
}

func (h *Handler) createPetDetails(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	input, ok := bindPetDetailsInput(c)
	if !ok {
		return
	}

	pet, err := h.profile.CreatePetDetails(c.Request.Context(), ownerID, input)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusCreated, petDetailsResponse(pet))
}

func (h *Handler) updatePetDetails(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	input, ok := bindPetDetailsInput(c)
	if !ok {
		return
	}

	pet, err := h.profile.UpdatePetDetails(c.Request.Context(), ownerID, input)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, petDetailsResponse(pet))
}

func (h *Handler) deletePetDetails(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	if err := h.profile.DeletePetDetails(c.Request.Context(), ownerID); err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) getDeviceSettings(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	settings, err := h.profile.GetDeviceSettings(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, deviceSettingsResponse(settings))
}

func (h *Handler) updateProfileDeviceSettings(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	var req deviceSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "valid device settings are required")
		return
	}

	settings, err := h.profile.UpdateDeviceSettings(c.Request.Context(), ownerID, domain.DeviceSettingsInput{
		Name:                   req.Name,
		ManualFeedPortionGrams: req.ManualFeedPortionGrams,
		ServoOpenDegrees:       req.ServoOpenDegrees,
		ServoClosedDegrees:     req.ServoClosedDegrees,
		AutomationEnabled:      req.AutomationEnabled,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, deviceSettingsResponse(settings))
}

func (h *Handler) getDeviceConfig(c *gin.Context) {
	settings, err := h.profile.GetDeviceConfig(c.Request.Context(), c.Param("deviceID"))
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, deviceSettingsResponse(settings))
}

func (h *Handler) calibrateDevice(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	settings, err := h.profile.CalibrateDevice(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, deviceSettingsResponse(settings))
}

func (h *Handler) getNotificationPreferences(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	preferences, err := h.profile.GetNotificationPreferences(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, notificationPreferencesResponse(preferences))
}

func (h *Handler) updateNotificationPreferences(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	var req notificationPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "valid notification preferences are required")
		return
	}

	preferences, err := h.profile.UpsertNotificationPreferences(c.Request.Context(), ownerID, domain.NotificationPreferencesInput{
		LowFoodAlert:         req.LowFoodAlert,
		EmptyWaterAlert:      req.EmptyWaterAlert,
		FeedingSuccessReport: req.FeedingSuccessReport,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, notificationPreferencesResponse(preferences))
}

func (h *Handler) patchNotificationPreferences(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	var req notificationPreferencesPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "valid notification preferences are required")
		return
	}

	current, err := h.profile.GetNotificationPreferences(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	input := domain.NotificationPreferencesInput{
		LowFoodAlert:         current.LowFoodAlert,
		EmptyWaterAlert:      current.EmptyWaterAlert,
		FeedingSuccessReport: current.FeedingSuccessReport,
	}
	if req.AlertLowFood != nil {
		input.LowFoodAlert = *req.AlertLowFood
	}
	if req.LowFoodAlert != nil {
		input.LowFoodAlert = *req.LowFoodAlert
	}
	if req.AlertEmptyWater != nil {
		input.EmptyWaterAlert = *req.AlertEmptyWater
	}
	if req.EmptyWaterAlert != nil {
		input.EmptyWaterAlert = *req.EmptyWaterAlert
	}
	if req.AlertFeedSuccess != nil {
		input.FeedingSuccessReport = *req.AlertFeedSuccess
	}
	if req.FeedingSuccessReport != nil {
		input.FeedingSuccessReport = *req.FeedingSuccessReport
	}

	preferences, err := h.profile.UpsertNotificationPreferences(c.Request.Context(), ownerID, input)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, notificationPreferencesResponse(preferences))
}

func bindPetDetailsInput(c *gin.Context) (domain.PetDetailsInput, bool) {
	var req petDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "name, species, breed, and daily_feed_target_grams are required")
		return domain.PetDetailsInput{}, false
	}

	birthDate, ok := parseOptionalDate(req.BirthDate)
	if !ok {
		respondError(c, http.StatusBadRequest, "birth_date must use YYYY-MM-DD format")
		return domain.PetDetailsInput{}, false
	}

	return domain.PetDetailsInput{
		DeviceID:             req.DeviceID,
		Name:                 req.Name,
		Species:              req.Species,
		Breed:                req.Breed,
		Gender:               req.Gender,
		BirthDate:            birthDate,
		DailyFeedTargetGrams: req.DailyFeedTargetGrams,
	}, true
}

func petDetailsResponse(pet domain.PetDetails) gin.H {
	var birthDate any
	if pet.BirthDate != nil {
		birthDate = pet.BirthDate.Format("2006-01-02")
	}
	return gin.H{
		"id":                      pet.ID,
		"owner_id":                pet.OwnerID,
		"device_id":               pet.DeviceID,
		"name":                    pet.Name,
		"species":                 pet.Species,
		"breed":                   pet.Breed,
		"gender":                  pet.Gender,
		"birth_date":              birthDate,
		"daily_feed_target_grams": pet.DailyFeedTargetGrams,
		"created_at":              pet.CreatedAt,
	}
}

func deviceSettingsResponse(settings domain.DeviceSettings) gin.H {
	var calibrationRequestedAt any
	if settings.CalibrationRequestedAt != nil {
		calibrationRequestedAt = settings.CalibrationRequestedAt
	}
	return gin.H{
		"id":                        settings.ID,
		"name":                      settings.Name,
		"manual_feed_portion_grams": settings.ManualFeedPortionGrams,
		"servo_open_degrees":        settings.ServoOpenDegrees,
		"servo_closed_degrees":      settings.ServoClosedDegrees,
		"automation_enabled":        settings.AutomationEnabled,
		"food_stock_percent":        settings.FoodStockPercent,
		"water_available":           settings.WaterAvailable,
		"water_status":              settings.WaterStatus,
		"calibration_status":        settings.CalibrationStatus,
		"calibration_requested_at":  calibrationRequestedAt,
		"last_seen_at":              settings.LastSeenAt,
		"config_updated_at":         settings.ConfigUpdatedAt,
	}
}

func notificationPreferencesResponse(preferences domain.NotificationPreferences) gin.H {
	return gin.H{
		"owner_id":               preferences.OwnerID,
		"low_food_alert":         preferences.LowFoodAlert,
		"empty_water_alert":      preferences.EmptyWaterAlert,
		"feeding_success_report": preferences.FeedingSuccessReport,
		"alert_low_food":         preferences.LowFoodAlert,
		"alert_empty_water":      preferences.EmptyWaterAlert,
		"alert_feed_success":     preferences.FeedingSuccessReport,
		"updated_at":             preferences.UpdatedAt,
	}
}
