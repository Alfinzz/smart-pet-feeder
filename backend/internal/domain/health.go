package domain

import "time"

type HealthPet struct {
	ID                   int64
	OwnerID              int64
	DeviceID             string
	Name                 string
	WeightKG             float64
	ActivityMinutes      int
	SleepHours           float64
	DailyFeedTargetGrams float64
}

type VitalSigns struct {
	ID              int64
	PetID           int64
	WeightKG        float64
	ActivityMinutes int
	SleepHours      float64
	RecordedAt      time.Time
	CreatedAt       time.Time
}

type VitalSignsInput struct {
	WeightKG        float64
	ActivityMinutes int
	SleepHours      float64
	RecordedAt      *time.Time
}

type WellnessScore struct {
	Score                    int
	Label                    string
	RawScore                 float64
	WeightComponent          float64
	ActivityComponent        float64
	SleepComponent           float64
	TaskPenalty              int
	OverdueTaskCount         int
	TargetWeightKG           float64
	TargetActivityMinutes    int
	TargetSleepHours         float64
	AverageDailyFeedGrams    float64
	DailyFeedTargetGrams     float64
	ConsumptionPercent       float64
	ConsumptionComponent     float64
	WeightStabilityComponent float64
}

type HealthOverview struct {
	Pet           HealthPet
	Vitals        *VitalSigns
	WellnessScore WellnessScore
}
