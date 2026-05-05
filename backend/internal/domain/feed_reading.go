package domain

import "time"

type FeedReading struct {
	ID               int64
	DeviceID         string
	WeightGrams      float64
	FoodStockPercent *float64
	WaterAvailable   *bool
	WaterStatus      string
	RecordedAt       time.Time
	CreatedAt        time.Time
}

type FeedHistoryFilter struct {
	From  *time.Time
	To    *time.Time
	Limit int
}
