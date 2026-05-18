package domain

import "time"

type OwnerProfile struct {
	ID    int64
	Name  string
	Email string
}

type PetProfile struct {
	ID                   int64
	OwnerID              int64
	DeviceID             string
	Name                 string
	Species              string
	Breed                string
	AgeYears             int
	WeightKG             float64
	DailyFeedTargetGrams float64
	HealthScore          int
	HealthStatus         string
	HealthHeadline       string
	HealthDescription    string
	ActivityMinutes      int
	SleepHours           float64
	PhotoPath            string
}

type PetProfileUpdate struct {
	Name                 string
	Species              string
	Breed                string
	AgeYears             int
	WeightKG             float64
	DailyFeedTargetGrams float64
	HealthScore          int
	HealthStatus         string
	HealthHeadline       string
	HealthDescription    string
	ActivityMinutes      int
	SleepHours           float64
	DeviceID             string
}

type DeviceStatus struct {
	ID               string
	Name             string
	FoodStockPercent float64
	FoodStockLabel   string
	WaterAvailable   bool
	WaterStatus      string
	LastSeenAt       time.Time
}

type DeviceStatusUpdate struct {
	ID               string
	FoodStockPercent *float64
	WaterAvailable   *bool
	WaterStatus      string
	LastSeenAt       time.Time
}

type DashboardOverview struct {
	Pet              PetProfile
	Device           DeviceStatus
	GreetingTitle    string
	GreetingSubtitle string
}

type DailyConsumption struct {
	Date       time.Time
	DayLabel   string
	TotalGrams float64
}

type WeeklyConsumption struct {
	Days                 []DailyConsumption
	DailyTargetGrams     float64
	TotalGrams           float64
	AverageGrams         float64
	RecommendedDaysCount int
}

type HealthVitals struct {
	WeightKG        float64
	ActivityMinutes int
	SleepHours      float64
}

type CareTask struct {
	ID        int64
	PetID     int64
	Category  string
	Title     string
	Subtitle  string
	DueLabel  string
	DueAt     *time.Time
	Priority  string
	SortOrder int
}

type CareTaskInput struct {
	Category  string
	Title     string
	Subtitle  string
	DueLabel  string
	DueAt     *time.Time
	Priority  string
	SortOrder int
}

type HealthSummary struct {
	Pet           PetProfile
	Score         int
	StatusLabel   string
	Headline      string
	Description   string
	Vitals        HealthVitals
	UpcomingTasks []CareTask
}

type ProfileSummary struct {
	Owner  OwnerProfile
	Pet    PetProfile
	Device DeviceStatus
}
