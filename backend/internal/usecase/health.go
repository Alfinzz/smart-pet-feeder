package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"smart-pet-monitoring/backend/internal/domain"
)

type HealthRepository interface {
	AverageDailyFeedConsumption(ctx context.Context, deviceID string, days int) (float64, error)
	GetPrimaryHealthPet(ctx context.Context, ownerID int64) (domain.HealthPet, error)
	GetLatestVitalSigns(ctx context.Context, petID int64) (domain.VitalSigns, error)
	GetPreviousVitalSigns(ctx context.Context, petID, latestID int64) (domain.VitalSigns, error)
	SaveVitalSigns(ctx context.Context, petID int64, input domain.VitalSignsInput) (domain.VitalSigns, error)
}

type HealthUsecase struct {
	repo HealthRepository
}

func NewHealthUsecase(repo HealthRepository) *HealthUsecase {
	return &HealthUsecase{repo: repo}
}

func (u *HealthUsecase) GetOverview(ctx context.Context, ownerID int64, days int) (domain.HealthOverview, error) {
	pet, err := u.requirePet(ctx, ownerID)
	if err != nil {
		return domain.HealthOverview{}, err
	}
	return u.buildOverview(ctx, pet, days, nil)
}

func (u *HealthUsecase) UpdateVitalSigns(ctx context.Context, ownerID int64, input domain.VitalSignsInput, days int) (domain.HealthOverview, error) {
	pet, err := u.requirePet(ctx, ownerID)
	if err != nil {
		return domain.HealthOverview{}, err
	}
	if input.WeightKG <= 0 {
		return domain.HealthOverview{}, fmt.Errorf("%w: weight_kg must be greater than 0", domain.ErrValidation)
	}
	if input.ActivityMinutes < 0 {
		return domain.HealthOverview{}, fmt.Errorf("%w: activity_minutes must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.SleepHours < 0 {
		return domain.HealthOverview{}, fmt.Errorf("%w: sleep_hours must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.RecordedAt == nil {
		now := time.Now().UTC()
		input.RecordedAt = &now
	}

	vitals, err := u.repo.SaveVitalSigns(ctx, pet.ID, input)
	if err != nil {
		return domain.HealthOverview{}, err
	}
	return u.buildOverview(ctx, pet, days, &vitals)
}

func (u *HealthUsecase) requirePet(ctx context.Context, ownerID int64) (domain.HealthPet, error) {
	if ownerID <= 0 {
		return domain.HealthPet{}, domain.ErrUnauthorized
	}
	pet, err := u.repo.GetPrimaryHealthPet(ctx, ownerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.HealthPet{}, fmt.Errorf("%w: pet profile is required before health overview can be calculated", domain.ErrNotFound)
		}
		return domain.HealthPet{}, err
	}
	if pet.DailyFeedTargetGrams <= 0 {
		pet.DailyFeedTargetGrams = defaultDailyFeedTargetGrams
	}
	return pet, nil
}

func (u *HealthUsecase) buildOverview(ctx context.Context, pet domain.HealthPet, days int, currentVitals *domain.VitalSigns) (domain.HealthOverview, error) {
	if days <= 0 {
		days = 7
	}
	if days > 31 {
		days = 31
	}

	avgFeed, err := u.repo.AverageDailyFeedConsumption(ctx, pet.DeviceID, days)
	if err != nil {
		return domain.HealthOverview{}, err
	}

	vitals := currentVitals
	if vitals == nil {
		latest, err := u.repo.GetLatestVitalSigns(ctx, pet.ID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return domain.HealthOverview{}, err
		}
		if err == nil {
			vitals = &latest
		}
	}

	var currentWeight *float64
	var previousWeight *float64
	if vitals != nil {
		currentWeight = &vitals.WeightKG
		previous, err := u.repo.GetPreviousVitalSigns(ctx, pet.ID, vitals.ID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return domain.HealthOverview{}, err
		}
		if err == nil {
			previousWeight = &previous.WeightKG
		}
	}

	return domain.HealthOverview{
		Pet:           pet,
		Vitals:        vitals,
		WellnessScore: CalculateWellnessScore(avgFeed, pet.DailyFeedTargetGrams, currentWeight, previousWeight),
	}, nil
}

func CalculateWellnessScore(avgFeed, targetFeed float64, currentWeight, previousWeight *float64) domain.WellnessScore {
	if targetFeed <= 0 {
		targetFeed = defaultDailyFeedTargetGrams
	}

	consumptionPercent := (avgFeed / targetFeed) * 100
	consumptionComponent := 100 - math.Abs(100-consumptionPercent)*1.2
	if consumptionPercent < 60 {
		consumptionComponent = consumptionPercent * 0.9
	}
	if consumptionPercent > 130 {
		consumptionComponent = 70 - (consumptionPercent-130)*0.5
	}
	consumptionComponent = clampFloat(consumptionComponent, 0, 100)

	weightComponent := 85.0
	if currentWeight != nil && previousWeight != nil && *previousWeight > 0 {
		changePercent := math.Abs(*currentWeight-*previousWeight) / *previousWeight * 100
		switch {
		case changePercent <= 2:
			weightComponent = 100
		case changePercent <= 5:
			weightComponent = 82
		case changePercent <= 10:
			weightComponent = 55
		default:
			weightComponent = 30
		}
	}

	scoreValue := consumptionComponent*0.7 + weightComponent*0.3
	score := int(math.Round(clampFloat(scoreValue, 0, 100)))
	return domain.WellnessScore{
		Score:                    score,
		Label:                    wellnessLabel(score),
		AverageDailyFeedGrams:    avgFeed,
		DailyFeedTargetGrams:     targetFeed,
		ConsumptionPercent:       consumptionPercent,
		ConsumptionComponent:     consumptionComponent,
		WeightStabilityComponent: weightComponent,
	}
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func wellnessLabel(score int) string {
	switch {
	case score >= 85:
		return "Excellent"
	case score >= 70:
		return "Good"
	case score >= 50:
		return "Watch"
	default:
		return "Needs Attention"
	}
}
