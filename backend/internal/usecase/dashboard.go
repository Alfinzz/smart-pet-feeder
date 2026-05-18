package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"smart-pet-monitoring/backend/internal/domain"
)

const (
	fallbackDeviceID            = "ESP32-001"
	defaultDailyFeedTargetGrams = 150
)

type DashboardRepository interface {
	CreateCareTask(ctx context.Context, petID int64, input domain.CareTaskInput) (domain.CareTask, error)
	DeleteCareTask(ctx context.Context, ownerID, taskID int64) error
	GetOwner(ctx context.Context, ownerID int64) (domain.OwnerProfile, error)
	GetPetProfile(ctx context.Context, ownerID int64) (domain.PetProfile, error)
	GetDeviceStatus(ctx context.Context, deviceID string) (domain.DeviceStatus, error)
	GetDailyConsumption(ctx context.Context, deviceID string, days int) ([]domain.DailyConsumption, error)
	ListUpcomingTasks(ctx context.Context, petID int64, limit int) ([]domain.CareTask, error)
	UpdateCareTask(ctx context.Context, ownerID, taskID int64, input domain.CareTaskInput) (domain.CareTask, error)
	UpdateDeviceSettings(ctx context.Context, deviceID, name string, foodStockPercent *float64, waterAvailable *bool, waterStatus string) (domain.DeviceStatus, error)
	UpdatePetPhoto(ctx context.Context, ownerID int64, photoPath string) (domain.PetProfile, string, error)
	UpsertPetProfile(ctx context.Context, ownerID int64, input domain.PetProfileUpdate) (domain.PetProfile, error)
}

type DashboardUsecase struct {
	repo            DashboardRepository
	defaultDeviceID string
}

func NewDashboardUsecase(repo DashboardRepository, defaultDeviceID string) *DashboardUsecase {
	return &DashboardUsecase{
		repo:            repo,
		defaultDeviceID: normalizeDefaultDeviceID(defaultDeviceID),
	}
}

func (u *DashboardUsecase) GetOverview(ctx context.Context, ownerID int64) (domain.DashboardOverview, error) {
	pet, device, err := u.getPetAndDevice(ctx, ownerID)
	if err != nil {
		return domain.DashboardOverview{}, err
	}

	return domain.DashboardOverview{
		Pet:              pet,
		Device:           device,
		GreetingTitle:    fmt.Sprintf("Hello, %s!", pet.Name),
		GreetingSubtitle: "Ready for breakfast?",
	}, nil
}

func (u *DashboardUsecase) GetWeeklyConsumption(ctx context.Context, ownerID int64, days int) (domain.WeeklyConsumption, error) {
	if days <= 0 {
		days = 7
	}
	if days > 31 {
		days = 31
	}

	pet, _, err := u.getPetAndDevice(ctx, ownerID)
	if err != nil {
		return domain.WeeklyConsumption{}, err
	}

	items, err := u.repo.GetDailyConsumption(ctx, pet.DeviceID, days)
	if err != nil {
		return domain.WeeklyConsumption{}, err
	}

	total := 0.0
	for i := range items {
		items[i].DayLabel = shortWeekdayLabel(items[i].Date.Weekday())
		total += items[i].TotalGrams
	}

	target := pet.DailyFeedTargetGrams
	if target <= 0 {
		target = defaultDailyFeedTargetGrams
	}

	average := 0.0
	if len(items) > 0 {
		average = total / float64(len(items))
	}

	return domain.WeeklyConsumption{
		Days:                 items,
		DailyTargetGrams:     target,
		TotalGrams:           total,
		AverageGrams:         average,
		RecommendedDaysCount: days,
	}, nil
}

func (u *DashboardUsecase) GetHealthSummary(ctx context.Context, ownerID int64) (domain.HealthSummary, error) {
	pet, _, err := u.getPetAndDevice(ctx, ownerID)
	if err != nil {
		return domain.HealthSummary{}, err
	}

	tasks := []domain.CareTask(nil)
	if pet.ID > 0 {
		tasks, err = u.repo.ListUpcomingTasks(ctx, pet.ID, 2)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return domain.HealthSummary{}, err
		}
	}
	if len(tasks) == 0 {
		tasks = defaultCareTasks()
	}

	return domain.HealthSummary{
		Pet:         pet,
		Score:       pet.HealthScore,
		StatusLabel: statusOrDefault(pet.HealthStatus, "Excellent"),
		Headline:    statusOrDefault(pet.HealthHeadline, "Optimal Wellness"),
		Description: statusOrDefault(
			pet.HealthDescription,
			"Your pet health metrics are stable this week. Keep maintaining the current diet and activity routines.",
		),
		Vitals: domain.HealthVitals{
			WeightKG:        pet.WeightKG,
			ActivityMinutes: pet.ActivityMinutes,
			SleepHours:      pet.SleepHours,
		},
		UpcomingTasks: tasks,
	}, nil
}

func (u *DashboardUsecase) GetProfile(ctx context.Context, ownerID int64) (domain.ProfileSummary, error) {
	owner, err := u.repo.GetOwner(ctx, ownerID)
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	pet, device, err := u.getPetAndDevice(ctx, ownerID)
	if err != nil {
		return domain.ProfileSummary{}, err
	}

	return domain.ProfileSummary{
		Owner:  owner,
		Pet:    pet,
		Device: device,
	}, nil
}

func (u *DashboardUsecase) EnsureDefaultProfile(ctx context.Context, ownerID int64) error {
	if ownerID <= 0 {
		return domain.ErrUnauthorized
	}
	_, err := u.repo.GetPetProfile(ctx, ownerID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	_, err = u.repo.UpsertPetProfile(ctx, ownerID, u.defaultPetUpdate(ownerID))
	return err
}

func (u *DashboardUsecase) UpdatePetProfile(ctx context.Context, ownerID int64, input domain.PetProfileUpdate) (domain.PetProfile, error) {
	if ownerID <= 0 {
		return domain.PetProfile{}, domain.ErrUnauthorized
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Species = strings.TrimSpace(input.Species)
	input.Breed = strings.TrimSpace(input.Breed)
	input.DeviceID = strings.TrimSpace(input.DeviceID)
	input.HealthStatus = strings.TrimSpace(input.HealthStatus)
	input.HealthHeadline = strings.TrimSpace(input.HealthHeadline)
	input.HealthDescription = strings.TrimSpace(input.HealthDescription)

	if input.Name == "" {
		return domain.PetProfile{}, fmt.Errorf("%w: pet name is required", domain.ErrValidation)
	}
	if input.Species == "" {
		input.Species = "Dog"
	}
	if input.Breed == "" {
		input.Breed = "Unknown"
	}
	if input.DeviceID == "" {
		input.DeviceID = u.defaultDeviceID
	}
	if input.AgeYears < 0 {
		return domain.PetProfile{}, fmt.Errorf("%w: age_years must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.WeightKG < 0 {
		return domain.PetProfile{}, fmt.Errorf("%w: weight_kg must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.DailyFeedTargetGrams < 0 {
		return domain.PetProfile{}, fmt.Errorf("%w: daily_feed_target_grams must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.HealthScore < 0 || input.HealthScore > 100 {
		return domain.PetProfile{}, fmt.Errorf("%w: health_score must be between 0 and 100", domain.ErrValidation)
	}
	if input.ActivityMinutes < 0 {
		return domain.PetProfile{}, fmt.Errorf("%w: activity_minutes must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.SleepHours < 0 {
		return domain.PetProfile{}, fmt.Errorf("%w: sleep_hours must be greater than or equal to 0", domain.ErrValidation)
	}

	return u.repo.UpsertPetProfile(ctx, ownerID, input)
}

func (u *DashboardUsecase) UpdatePetPhoto(ctx context.Context, ownerID int64, photoPath string) (domain.PetProfile, string, error) {
	if ownerID <= 0 {
		return domain.PetProfile{}, "", domain.ErrUnauthorized
	}
	photoPath = strings.TrimSpace(photoPath)
	if photoPath == "" {
		return domain.PetProfile{}, "", fmt.Errorf("%w: photo path is required", domain.ErrValidation)
	}
	if err := u.EnsureDefaultProfile(ctx, ownerID); err != nil {
		return domain.PetProfile{}, "", err
	}
	return u.repo.UpdatePetPhoto(ctx, ownerID, photoPath)
}

func (u *DashboardUsecase) UpdateDeviceSettings(ctx context.Context, ownerID int64, name string, foodStockPercent *float64, waterAvailable *bool, waterStatus string) (domain.DeviceStatus, error) {
	pet, _, err := u.getPetAndDevice(ctx, ownerID)
	if err != nil {
		return domain.DeviceStatus{}, err
	}
	if pet.ID == 0 {
		if err := u.EnsureDefaultProfile(ctx, ownerID); err != nil {
			return domain.DeviceStatus{}, err
		}
		pet, err = u.repo.GetPetProfile(ctx, ownerID)
		if err != nil {
			return domain.DeviceStatus{}, err
		}
	}
	if foodStockPercent != nil && (*foodStockPercent < 0 || *foodStockPercent > 100) {
		return domain.DeviceStatus{}, fmt.Errorf("%w: food_stock_percent must be between 0 and 100", domain.ErrValidation)
	}
	status, err := u.repo.UpdateDeviceSettings(ctx, pet.DeviceID, strings.TrimSpace(name), foodStockPercent, waterAvailable, strings.TrimSpace(waterStatus))
	if err != nil {
		return domain.DeviceStatus{}, err
	}
	return u.normalizeDeviceStatus(status), nil
}

func (u *DashboardUsecase) ListCareTasks(ctx context.Context, ownerID int64, limit int) ([]domain.CareTask, error) {
	pet, err := u.ensurePetForOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	tasks, err := u.repo.ListUpcomingTasks(ctx, pet.ID, limit)
	if errors.Is(err, domain.ErrNotFound) {
		return []domain.CareTask{}, nil
	}
	return tasks, err
}

func (u *DashboardUsecase) CreateCareTask(ctx context.Context, ownerID int64, input domain.CareTaskInput) (domain.CareTask, error) {
	pet, err := u.ensurePetForOwner(ctx, ownerID)
	if err != nil {
		return domain.CareTask{}, err
	}
	input, err = validateCareTaskInput(input)
	if err != nil {
		return domain.CareTask{}, err
	}
	return u.repo.CreateCareTask(ctx, pet.ID, input)
}

func (u *DashboardUsecase) UpdateCareTask(ctx context.Context, ownerID, taskID int64, input domain.CareTaskInput) (domain.CareTask, error) {
	if taskID <= 0 {
		return domain.CareTask{}, fmt.Errorf("%w: task_id is required", domain.ErrValidation)
	}
	input, err := validateCareTaskInput(input)
	if err != nil {
		return domain.CareTask{}, err
	}
	return u.repo.UpdateCareTask(ctx, ownerID, taskID, input)
}

func (u *DashboardUsecase) DeleteCareTask(ctx context.Context, ownerID, taskID int64) error {
	if taskID <= 0 {
		return fmt.Errorf("%w: task_id is required", domain.ErrValidation)
	}
	return u.repo.DeleteCareTask(ctx, ownerID, taskID)
}

func (u *DashboardUsecase) getPetAndDevice(ctx context.Context, ownerID int64) (domain.PetProfile, domain.DeviceStatus, error) {
	if ownerID <= 0 {
		return domain.PetProfile{}, domain.DeviceStatus{}, domain.ErrUnauthorized
	}

	pet, err := u.repo.GetPetProfile(ctx, ownerID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.PetProfile{}, domain.DeviceStatus{}, err
		}
		pet = defaultPet(ownerID)
	}
	if strings.TrimSpace(pet.DeviceID) == "" {
		pet.DeviceID = u.defaultDeviceID
	}

	device, err := u.repo.GetDeviceStatus(ctx, pet.DeviceID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.PetProfile{}, domain.DeviceStatus{}, err
		}
		device = u.defaultDeviceStatus(pet.DeviceID)
	}
	device = u.normalizeDeviceStatus(device)

	return pet, device, nil
}

func (u *DashboardUsecase) ensurePetForOwner(ctx context.Context, ownerID int64) (domain.PetProfile, error) {
	if ownerID <= 0 {
		return domain.PetProfile{}, domain.ErrUnauthorized
	}
	pet, err := u.repo.GetPetProfile(ctx, ownerID)
	if err == nil {
		return pet, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.PetProfile{}, err
	}
	return u.repo.UpsertPetProfile(ctx, ownerID, u.defaultPetUpdate(ownerID))
}

func defaultPet(ownerID int64) domain.PetProfile {
	return domain.PetProfile{
		OwnerID:              ownerID,
		DeviceID:             fallbackDeviceID,
		Name:                 "Fluffy",
		Species:              "Dog",
		Breed:                "Golden Retriever",
		AgeYears:             3,
		WeightKG:             25.4,
		DailyFeedTargetGrams: defaultDailyFeedTargetGrams,
		HealthScore:          92,
		HealthStatus:         "Excellent",
		HealthHeadline:       "Optimal Wellness",
		HealthDescription:    "Your pet health metrics are stable this week. Keep maintaining the current diet and activity routines.",
		ActivityMinutes:      45,
		SleepHours:           9.5,
	}
}

func (u *DashboardUsecase) defaultPetUpdate(ownerID int64) domain.PetProfileUpdate {
	pet := defaultPet(ownerID)
	pet.DeviceID = u.defaultDeviceID
	return domain.PetProfileUpdate{
		Name:                 pet.Name,
		Species:              pet.Species,
		Breed:                pet.Breed,
		AgeYears:             pet.AgeYears,
		WeightKG:             pet.WeightKG,
		DailyFeedTargetGrams: pet.DailyFeedTargetGrams,
		HealthScore:          pet.HealthScore,
		HealthStatus:         pet.HealthStatus,
		HealthHeadline:       pet.HealthHeadline,
		HealthDescription:    pet.HealthDescription,
		ActivityMinutes:      pet.ActivityMinutes,
		SleepHours:           pet.SleepHours,
		DeviceID:             pet.DeviceID,
	}
}

func (u *DashboardUsecase) defaultDeviceStatus(deviceID string) domain.DeviceStatus {
	if strings.TrimSpace(deviceID) == "" {
		deviceID = u.defaultDeviceID
	}
	return u.normalizeDeviceStatus(domain.DeviceStatus{
		ID:               deviceID,
		Name:             "Kitchen Smart Feeder",
		FoodStockPercent: 75,
		WaterAvailable:   true,
		WaterStatus:      "Clean & Fresh",
	})
}

func (u *DashboardUsecase) normalizeDeviceStatus(status domain.DeviceStatus) domain.DeviceStatus {
	return normalizeDeviceStatusWithDefault(status, u.defaultDeviceID)
}

func normalizeDeviceStatus(status domain.DeviceStatus) domain.DeviceStatus {
	return normalizeDeviceStatusWithDefault(status, fallbackDeviceID)
}

func normalizeDeviceStatusWithDefault(status domain.DeviceStatus, defaultDeviceID string) domain.DeviceStatus {
	defaultDeviceID = normalizeDefaultDeviceID(defaultDeviceID)
	if strings.TrimSpace(status.ID) == "" {
		status.ID = defaultDeviceID
	}
	if strings.TrimSpace(status.Name) == "" {
		status.Name = "Smart Feeder"
	}
	if status.FoodStockPercent < 0 {
		status.FoodStockPercent = 0
	}
	if status.FoodStockPercent > 100 {
		status.FoodStockPercent = 100
	}
	if status.LastSeenAt.IsZero() {
		status.LastSeenAt = time.Now().UTC()
	}
	status.FoodStockLabel = foodStockLabel(status.FoodStockPercent)
	if strings.TrimSpace(status.WaterStatus) == "" {
		if status.WaterAvailable {
			status.WaterStatus = "Clean & Fresh"
		} else {
			status.WaterStatus = "Unavailable"
		}
	}
	return status
}

func normalizeDefaultDeviceID(deviceID string) string {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return fallbackDeviceID
	}
	return deviceID
}

func foodStockLabel(percent float64) string {
	switch {
	case percent >= 70:
		return "Full"
	case percent >= 35:
		return "Medium"
	case percent > 0:
		return "Low"
	default:
		return "Empty"
	}
}

func shortWeekdayLabel(weekday time.Weekday) string {
	name := weekday.String()
	if len(name) < 3 {
		return name
	}
	return name[:3]
}

func statusOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func defaultCareTasks() []domain.CareTask {
	return []domain.CareTask{
		{
			ID:        1,
			Category:  "vaccination",
			Title:     "Vaccination",
			Subtitle:  "Annual Rabies Booster",
			DueLabel:  "Due in 5 days",
			Priority:  "high",
			SortOrder: 1,
		},
		{
			ID:        2,
			Category:  "checkup",
			Title:     "Vet Checkup",
			Subtitle:  "General Wellness Exam",
			DueLabel:  "Oct 24",
			Priority:  "normal",
			SortOrder: 2,
		},
	}
}

func validateCareTaskInput(input domain.CareTaskInput) (domain.CareTaskInput, error) {
	input.Category = strings.TrimSpace(input.Category)
	input.Title = strings.TrimSpace(input.Title)
	input.Subtitle = strings.TrimSpace(input.Subtitle)
	input.DueLabel = strings.TrimSpace(input.DueLabel)
	input.Priority = strings.TrimSpace(input.Priority)
	if input.Category == "" {
		return domain.CareTaskInput{}, fmt.Errorf("%w: category is required", domain.ErrValidation)
	}
	if input.Title == "" {
		return domain.CareTaskInput{}, fmt.Errorf("%w: title is required", domain.ErrValidation)
	}
	if input.Subtitle == "" {
		return domain.CareTaskInput{}, fmt.Errorf("%w: subtitle is required", domain.ErrValidation)
	}
	if input.DueLabel == "" {
		return domain.CareTaskInput{}, fmt.Errorf("%w: due_label is required", domain.ErrValidation)
	}
	if input.Priority == "" {
		input.Priority = "normal"
	}
	return input, nil
}
