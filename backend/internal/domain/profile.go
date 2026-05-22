package domain

import "time"

type PetDetails struct {
	ID                   int64
	OwnerID              int64
	DeviceID             string
	Name                 string
	Species              string
	Breed                string
	Gender               string
	BirthDate            *time.Time
	DailyFeedTargetGrams float64
	CreatedAt            time.Time
}

type PetDetailsInput struct {
	DeviceID             string
	Name                 string
	Species              string
	Breed                string
	Gender               string
	BirthDate            *time.Time
	DailyFeedTargetGrams float64
}

type DeviceSettings struct {
	ID                     string
	Name                   string
	ManualFeedPortionGrams float64
	ServoOpenDegrees       int
	ServoClosedDegrees     int
	AutomationEnabled      bool
	FoodStockPercent       float64
	WaterAvailable         bool
	WaterStatus            string
	CalibrationStatus      string
	CalibrationRequestedAt *time.Time
	LastSeenAt             time.Time
	ConfigUpdatedAt        time.Time
}

type DeviceSettingsInput struct {
	Name                   string
	ManualFeedPortionGrams float64
	ServoOpenDegrees       *int
	ServoClosedDegrees     *int
	AutomationEnabled      *bool
}

type NotificationPreferences struct {
	OwnerID              int64
	LowFoodAlert         bool
	EmptyWaterAlert      bool
	FeedingSuccessReport bool
	UpdatedAt            time.Time
}

type NotificationPreferencesInput struct {
	LowFoodAlert         bool
	EmptyWaterAlert      bool
	FeedingSuccessReport bool
}
