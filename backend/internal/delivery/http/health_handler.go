package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
)

type vitalSignsRequest struct {
	WeightKG        float64 `json:"weight_kg" binding:"required,gt=0"`
	ActivityMinutes int     `json:"activity_minutes" binding:"gte=0"`
	SleepHours      float64 `json:"sleep_hours" binding:"gte=0"`
	RecordedAt      string  `json:"recorded_at"`
}

func (h *Handler) getHealthOverview(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	days, ok := parseOptionalIntQuery(c, "days")
	if !ok {
		return
	}

	overview, err := h.health.GetOverview(c.Request.Context(), ownerID, days)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, healthOverviewResponse(overview))
}

func (h *Handler) updateVitalSigns(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	var req vitalSignsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "weight_kg, activity_minutes, and sleep_hours are required")
		return
	}

	recordedAt, ok := parseOptionalDateTime(req.RecordedAt)
	if !ok {
		respondError(c, http.StatusBadRequest, "recorded_at must use RFC3339 format")
		return
	}

	days, ok := parseOptionalIntQuery(c, "days")
	if !ok {
		return
	}

	overview, err := h.health.UpdateVitalSigns(c.Request.Context(), ownerID, domain.VitalSignsInput{
		WeightKG:        req.WeightKG,
		ActivityMinutes: req.ActivityMinutes,
		SleepHours:      req.SleepHours,
		RecordedAt:      recordedAt,
	}, days)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, healthOverviewResponse(overview))
}

func healthOverviewResponse(overview domain.HealthOverview) gin.H {
	var vitals any
	if overview.Vitals != nil {
		vitals = vitalSignsResponse(*overview.Vitals)
	}

	score := overview.WellnessScore
	return gin.H{
		"pet": gin.H{
			"id":                      overview.Pet.ID,
			"name":                    overview.Pet.Name,
			"device_id":               overview.Pet.DeviceID,
			"daily_feed_target_grams": overview.Pet.DailyFeedTargetGrams,
		},
		"vitals": vitals,
		"wellness_score": gin.H{
			"score":                      score.Score,
			"label":                      score.Label,
			"raw_score":                  score.RawScore,
			"weight_component":           score.WeightComponent,
			"activity_component":         score.ActivityComponent,
			"sleep_component":            score.SleepComponent,
			"task_penalty":               score.TaskPenalty,
			"overdue_task_count":         score.OverdueTaskCount,
			"target_weight_kg":           score.TargetWeightKG,
			"target_activity_minutes":    score.TargetActivityMinutes,
			"target_sleep_hours":         score.TargetSleepHours,
			"average_daily_feed_grams":   score.AverageDailyFeedGrams,
			"daily_feed_target_grams":    score.DailyFeedTargetGrams,
			"consumption_percent":        score.ConsumptionPercent,
			"consumption_component":      score.ConsumptionComponent,
			"weight_stability_component": score.WeightStabilityComponent,
		},
	}
}

func vitalSignsResponse(vitals domain.VitalSigns) gin.H {
	return gin.H{
		"id":               vitals.ID,
		"pet_id":           vitals.PetID,
		"weight_kg":        vitals.WeightKG,
		"activity_minutes": vitals.ActivityMinutes,
		"sleep_hours":      vitals.SleepHours,
		"recorded_at":      vitals.RecordedAt,
		"created_at":       vitals.CreatedAt,
	}
}

func parseOptionalDateTime(value string) (*time.Time, bool) {
	if value == "" {
		return nil, true
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, false
	}
	parsed = parsed.UTC()
	return &parsed, true
}
