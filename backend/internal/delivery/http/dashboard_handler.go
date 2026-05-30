package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"smart-pet-monitoring/backend/internal/domain"
)

type updatePetProfileRequest struct {
	Name                 string  `json:"name" binding:"required"`
	Species              string  `json:"species"`
	Breed                string  `json:"breed"`
	AgeYears             int     `json:"age_years" binding:"gte=0"`
	WeightKG             float64 `json:"weight_kg" binding:"gte=0"`
	DailyFeedTargetGrams float64 `json:"daily_feed_target_grams" binding:"gte=0"`
	HealthScore          int     `json:"health_score" binding:"gte=0,lte=100"`
	HealthStatus         string  `json:"health_status"`
	HealthHeadline       string  `json:"health_headline"`
	HealthDescription    string  `json:"health_description"`
	ActivityMinutes      int     `json:"activity_minutes" binding:"gte=0"`
	SleepHours           float64 `json:"sleep_hours" binding:"gte=0"`
	DeviceID             string  `json:"device_id"`
}

type updateDeviceSettingsRequest struct {
	Name             string   `json:"name"`
	FoodStockPercent *float64 `json:"food_stock_percent" binding:"omitempty,gte=0,lte=100"`
	WaterAvailable   *bool    `json:"water_available"`
	WaterStatus      string   `json:"water_status"`
}

type careTaskRequest struct {
	Category    string `json:"category" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	DueLabel    string `json:"due_label"`
	DueAt       string `json:"due_at"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	SortOrder   int    `json:"sort_order"`
}

type careTaskStatusRequest struct {
	Status      string `json:"status" binding:"required"`
	AgeInMonths *int   `json:"age_in_months" binding:"omitempty,gte=0"`
	Age         *int   `json:"age" binding:"omitempty,gte=0"`
}

func (h *Handler) getDashboardOverview(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	overview, err := h.dashboard.GetOverview(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"greeting_title":    overview.GreetingTitle,
		"greeting_subtitle": overview.GreetingSubtitle,
		"pet":               h.petProfileResponse(c, overview.Pet),
		"device":            deviceStatusResponse(overview.Device),
	})
}

func (h *Handler) updatePetProfile(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	var req updatePetProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "valid pet profile fields are required")
		return
	}

	pet, err := h.dashboard.UpdatePetProfile(c.Request.Context(), ownerID, domain.PetProfileUpdate{
		Name:                 req.Name,
		Species:              req.Species,
		Breed:                req.Breed,
		AgeYears:             req.AgeYears,
		WeightKG:             req.WeightKG,
		DailyFeedTargetGrams: req.DailyFeedTargetGrams,
		HealthScore:          req.HealthScore,
		HealthStatus:         req.HealthStatus,
		HealthHeadline:       req.HealthHeadline,
		HealthDescription:    req.HealthDescription,
		ActivityMinutes:      req.ActivityMinutes,
		SleepHours:           req.SleepHours,
		DeviceID:             req.DeviceID,
	})
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.petProfileResponse(c, pet))
}

func (h *Handler) updateDeviceSettings(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	var req updateDeviceSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "valid device settings are required")
		return
	}

	device, err := h.dashboard.UpdateDeviceSettings(
		c.Request.Context(),
		ownerID,
		req.Name,
		req.FoodStockPercent,
		req.WaterAvailable,
		req.WaterStatus,
	)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, deviceStatusResponse(device))
}

func (h *Handler) getWeeklyConsumption(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	days, ok := parseOptionalIntQuery(c, "days")
	if !ok {
		return
	}

	weekly, err := h.dashboard.GetWeeklyConsumption(c.Request.Context(), ownerID, days)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	items := make([]gin.H, 0, len(weekly.Days))
	for _, item := range weekly.Days {
		items = append(items, gin.H{
			"date":        item.Date.Format("2006-01-02"),
			"day_label":   item.DayLabel,
			"total_grams": item.TotalGrams,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":                   items,
		"daily_target_grams":     weekly.DailyTargetGrams,
		"total_grams":            weekly.TotalGrams,
		"average_grams":          weekly.AverageGrams,
		"recommended_days_count": weekly.RecommendedDaysCount,
	})
}

func (h *Handler) getHealthSummary(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	summary, err := h.dashboard.GetHealthSummary(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	tasks := make([]gin.H, 0, len(summary.UpcomingTasks))
	for _, task := range summary.UpcomingTasks {
		tasks = append(tasks, careTaskResponse(task))
	}

	c.JSON(http.StatusOK, gin.H{
		"pet":          h.petProfileResponse(c, summary.Pet),
		"score":        summary.Score,
		"status_label": summary.StatusLabel,
		"headline":     summary.Headline,
		"description":  summary.Description,
		"vitals": gin.H{
			"weight_kg":        summary.Vitals.WeightKG,
			"activity_minutes": summary.Vitals.ActivityMinutes,
			"sleep_hours":      summary.Vitals.SleepHours,
		},
		"upcoming_tasks": tasks,
	})
}

func (h *Handler) listCareTasks(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	limit, ok := parseOptionalIntQuery(c, "limit")
	if !ok {
		return
	}

	tasks, err := h.dashboard.ListCareTasks(c.Request.Context(), ownerID, limit)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	items := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, careTaskResponse(task))
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) createCareTask(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	input, ok := bindCareTaskInput(c)
	if !ok {
		return
	}

	task, err := h.dashboard.CreateCareTask(c.Request.Context(), ownerID, input)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusCreated, careTaskResponse(task))
}

func (h *Handler) updateCareTask(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	taskID, ok := parseTaskID(c)
	if !ok {
		return
	}

	input, ok := bindCareTaskInput(c)
	if !ok {
		return
	}

	task, err := h.dashboard.UpdateCareTask(c.Request.Context(), ownerID, taskID, input)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, careTaskResponse(task))
}

func (h *Handler) updateCareTaskStatus(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	taskID, ok := parseTaskID(c)
	if !ok {
		return
	}

	var req careTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "status is required")
		return
	}

	ageInMonths := req.AgeInMonths
	if ageInMonths == nil {
		ageInMonths = req.Age
	}

	task, err := h.dashboard.UpdateCareTaskStatus(c.Request.Context(), ownerID, taskID, req.Status, ageInMonths)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, careTaskResponse(task))
}

func (h *Handler) deleteCareTask(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	taskID, ok := parseTaskID(c)
	if !ok {
		return
	}

	if err := h.dashboard.DeleteCareTask(c.Request.Context(), ownerID, taskID); err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) getProfile(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	profile, err := h.dashboard.GetProfile(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"owner":  ownerProfileResponse(profile.Owner),
		"pet":    h.petProfileResponse(c, profile.Pet),
		"device": deviceStatusResponse(profile.Device),
	})
}

func (h *Handler) listAlerts(c *gin.Context) {
	ownerID, ok := ownerIDFromContext(c)
	if !ok {
		respondUsecaseError(c, domain.ErrUnauthorized)
		return
	}

	alerts, err := h.dashboard.ListAlerts(c.Request.Context(), ownerID)
	if err != nil {
		respondUsecaseError(c, err)
		return
	}

	items := make([]gin.H, 0, len(alerts))
	for _, alert := range alerts {
		items = append(items, userAlertResponse(alert))
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func ownerProfileResponse(owner domain.OwnerProfile) gin.H {
	return gin.H{
		"id":    owner.ID,
		"name":  owner.Name,
		"email": owner.Email,
	}
}

func (h *Handler) petProfileResponse(c *gin.Context, pet domain.PetProfile) gin.H {
	return gin.H{
		"id":                      pet.ID,
		"owner_id":                pet.OwnerID,
		"device_id":               pet.DeviceID,
		"name":                    pet.Name,
		"species":                 pet.Species,
		"breed":                   pet.Breed,
		"age_years":               pet.AgeYears,
		"age_in_months":           pet.AgeYears * 12,
		"weight_kg":               pet.WeightKG,
		"daily_feed_target_grams": pet.DailyFeedTargetGrams,
		"health_score":            pet.HealthScore,
		"health_status":           pet.HealthStatus,
		"health_headline":         pet.HealthHeadline,
		"health_description":      pet.HealthDescription,
		"activity_minutes":        pet.ActivityMinutes,
		"sleep_hours":             pet.SleepHours,
		"photo_url":               buildPublicURL(publicBaseURLFromRequest(h.publicBaseURL, c.Request), pet.PhotoPath),
	}
}

func deviceStatusResponse(device domain.DeviceStatus) gin.H {
	return gin.H{
		"id":                 device.ID,
		"name":               device.Name,
		"food_stock_percent": device.FoodStockPercent,
		"food_stock_label":   device.FoodStockLabel,
		"water_available":    device.WaterAvailable,
		"water_status":       device.WaterStatus,
		"last_seen_at":       device.LastSeenAt,
	}
}

func careTaskResponse(task domain.CareTask) gin.H {
	var dueAt any
	if task.DueAt != nil {
		dueAt = task.DueAt.Format("2006-01-02")
	}
	dueLabel := task.DueLabel
	if dueLabel == "" {
		dueLabel = careTaskDueLabel(task.DueAt)
	}
	description := task.Description
	if description == "" {
		description = task.Subtitle
	}
	return gin.H{
		"id":          task.ID,
		"pet_id":      task.PetID,
		"category":    task.Category,
		"title":       task.Title,
		"subtitle":    description,
		"description": description,
		"due_label":   dueLabel,
		"due_at":      dueAt,
		"due_date":    dueAt,
		"status":      task.Status,
		"priority":    task.Priority,
		"sort_order":  task.SortOrder,
	}
}

func bindCareTaskInput(c *gin.Context) (domain.CareTaskInput, bool) {
	var req careTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "category and title are required")
		return domain.CareTaskInput{}, false
	}

	dueValue := req.DueDate
	if dueValue == "" {
		dueValue = req.DueAt
	}
	dueAt, ok := parseOptionalDate(dueValue)
	if !ok {
		respondError(c, http.StatusBadRequest, "due_date must use YYYY-MM-DD format")
		return domain.CareTaskInput{}, false
	}

	description := req.Description
	if description == "" {
		description = req.Subtitle
	}

	return domain.CareTaskInput{
		Category:    req.Category,
		Title:       req.Title,
		Description: description,
		DueLabel:    req.DueLabel,
		DueAt:       dueAt,
		Status:      req.Status,
		Priority:    req.Priority,
		SortOrder:   req.SortOrder,
	}, true
}

func parseTaskID(c *gin.Context) (int64, bool) {
	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil || taskID <= 0 {
		respondError(c, http.StatusBadRequest, "task_id must be a positive integer")
		return 0, false
	}
	return taskID, true
}

func parseOptionalDate(value string) (*time.Time, bool) {
	if value == "" {
		return nil, true
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, false
	}
	return &parsed, true
}

func careTaskDueLabel(dueAt *time.Time) string {
	if dueAt == nil {
		return ""
	}
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	dueDate := time.Date(dueAt.Year(), dueAt.Month(), dueAt.Day(), 0, 0, 0, 0, today.Location())
	days := int(dueDate.Sub(today).Hours() / 24)
	switch {
	case days < 0:
		return "Overdue"
	case days == 0:
		return "Due today"
	case days == 1:
		return "Due in 1 day"
	case days <= 7:
		return "Due in " + strconv.Itoa(days) + " days"
	default:
		return dueAt.Format("Jan 2")
	}
}

func userAlertResponse(alert domain.UserAlert) gin.H {
	var dueAt any
	if alert.DueAt != nil {
		dueAt = alert.DueAt.Format("2006-01-02")
	}
	return gin.H{
		"id":       alert.ID,
		"type":     alert.Type,
		"title":    alert.Title,
		"message":  alert.Message,
		"severity": alert.Severity,
		"due_date": dueAt,
	}
}
