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
	CountOverduePendingTasks(ctx context.Context, petID int64) (int, error)
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

	scoreVitals := vitals
	if scoreVitals == nil && (pet.WeightKG > 0 || pet.ActivityMinutes > 0 || pet.SleepHours > 0) {
		scoreVitals = &domain.VitalSigns{
			PetID:           pet.ID,
			WeightKG:        pet.WeightKG,
			ActivityMinutes: pet.ActivityMinutes,
			SleepHours:      pet.SleepHours,
		}
	}

	overdueTasks, err := u.repo.CountOverduePendingTasks(ctx, pet.ID)
	if err != nil {
		return domain.HealthOverview{}, err
	}

	score := CalculateSAWHealthScore(scoreVitals, pet.WeightKG, overdueTasks)
	score.AverageDailyFeedGrams = avgFeed
	score.DailyFeedTargetGrams = pet.DailyFeedTargetGrams

	return domain.HealthOverview{
		Pet:           pet,
		Vitals:        vitals,
		WellnessScore: score,
	}, nil
}

const (
	defaultTargetActivityMinutes = 30
	defaultTargetSleepHours      = 10
	overdueTaskPenaltyPoints     = 15
)

func CalculateSAWHealthScore(vitals *domain.VitalSigns, targetWeightKG float64, overdueTaskCount int) domain.WellnessScore {
	targetActivityMinutes := defaultTargetActivityMinutes
	targetSleepHours := float64(defaultTargetSleepHours)

	if vitals == nil {
		score := 0
		return domain.WellnessScore{
			Score:                 score,
			Label:                 wellnessLabel(score),
			RawScore:              0,
			TaskPenalty:           maxInt(overdueTaskCount, 0) * overdueTaskPenaltyPoints,
			OverdueTaskCount:      maxInt(overdueTaskCount, 0),
			TargetWeightKG:        targetWeightKG,
			TargetActivityMinutes: targetActivityMinutes,
			TargetSleepHours:      targetSleepHours,
		}
	}

	if targetWeightKG <= 0 {
		targetWeightKG = vitals.WeightKG
	}

	weightComponent := closenessComponent(vitals.WeightKG, targetWeightKG)
	activityComponent := benefitComponent(float64(vitals.ActivityMinutes), float64(targetActivityMinutes))
	sleepComponent := closenessComponent(vitals.SleepHours, targetSleepHours)
	rawScore := weightComponent*0.40 + activityComponent*0.35 + sleepComponent*0.25
	overdueTaskCount = maxInt(overdueTaskCount, 0)
	penalty := overdueTaskCount * overdueTaskPenaltyPoints
	finalScore := clampFloat(rawScore-float64(penalty), 0, 100)
	score := int(math.Round(finalScore))

	return domain.WellnessScore{
		Score:                 score,
		Label:                 wellnessLabel(score),
		RawScore:              rawScore,
		WeightComponent:       weightComponent,
		ActivityComponent:     activityComponent,
		SleepComponent:        sleepComponent,
		TaskPenalty:           penalty,
		OverdueTaskCount:      overdueTaskCount,
		TargetWeightKG:        targetWeightKG,
		TargetActivityMinutes: targetActivityMinutes,
		TargetSleepHours:      targetSleepHours,
	}
}

func closenessComponent(value, target float64) float64 {
	if value <= 0 || target <= 0 {
		return 0
	}
	return clampFloat(100-(math.Abs(value-target)/target*100), 0, 100)
}

func benefitComponent(value, target float64) float64 {
	if value <= 0 || target <= 0 {
		return 0
	}
	return clampFloat((value/target)*100, 0, 100)
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

func healthHeadline(score int, overdueTaskCount int) string {
	if overdueTaskCount > 0 {
		return "Care Tasks Overdue"
	}
	switch {
	case score >= 85:
		return "Optimal Wellness"
	case score >= 70:
		return "Healthy Routine"
	case score >= 50:
		return "Needs Monitoring"
	default:
		return "Needs Attention"
	}
}

func healthDescription(score int, overdueTaskCount int) string {
	if overdueTaskCount > 0 {
		return "Complete overdue medical tasks to restore the health score."
	}
	switch {
	case score >= 85:
		return "Weight, activity, and sleep are close to the ideal targets."
	case score >= 70:
		return "Most health metrics are on track. Keep the routine consistent."
	case score >= 50:
		return "Some vitals are outside target range. Review activity and rest."
	default:
		return "Vitals are far from target range. Consider checking your pet care routine."
	}
}

func maxInt(value, min int) int {
	if value < min {
		return min
	}
	return value
}
