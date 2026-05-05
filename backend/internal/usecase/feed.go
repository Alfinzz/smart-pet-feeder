package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"smart-pet-monitoring/backend/internal/domain"
)

type FeedRepository interface {
	Create(ctx context.Context, reading *domain.FeedReading) error
	ListHistory(ctx context.Context, filter domain.FeedHistoryFilter) ([]domain.FeedReading, error)
	UpsertDeviceStatus(ctx context.Context, update domain.DeviceStatusUpdate) (domain.DeviceStatus, error)
}

type FeedUsecase struct {
	repo         FeedRepository
	defaultLimit int
	maxLimit     int
}

type CreateFeedReadingInput struct {
	DeviceID         string
	WeightGrams      float64
	FoodStockPercent *float64
	WaterAvailable   *bool
	WaterStatus      string
	RecordedAt       *time.Time
}

type UpdateDeviceStatusInput struct {
	DeviceID         string
	FoodStockPercent *float64
	WaterAvailable   *bool
	WaterStatus      string
}

func NewFeedUsecase(repo FeedRepository, defaultLimit, maxLimit int) *FeedUsecase {
	if defaultLimit <= 0 {
		defaultLimit = 50
	}
	if maxLimit < defaultLimit {
		maxLimit = defaultLimit
	}
	return &FeedUsecase{
		repo:         repo,
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
	}
}

func (u *FeedUsecase) CreateReading(ctx context.Context, input CreateFeedReadingInput) (domain.FeedReading, error) {
	deviceID := strings.TrimSpace(input.DeviceID)
	if deviceID == "" {
		return domain.FeedReading{}, fmt.Errorf("%w: device_id is required", domain.ErrValidation)
	}
	if input.WeightGrams < 0 {
		return domain.FeedReading{}, fmt.Errorf("%w: weight_grams must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.FoodStockPercent != nil && (*input.FoodStockPercent < 0 || *input.FoodStockPercent > 100) {
		return domain.FeedReading{}, fmt.Errorf("%w: food_stock_percent must be between 0 and 100", domain.ErrValidation)
	}

	recordedAt := time.Now().UTC()
	if input.RecordedAt != nil {
		recordedAt = input.RecordedAt.UTC()
	}

	reading := domain.FeedReading{
		DeviceID:         deviceID,
		WeightGrams:      input.WeightGrams,
		FoodStockPercent: input.FoodStockPercent,
		WaterAvailable:   input.WaterAvailable,
		WaterStatus:      strings.TrimSpace(input.WaterStatus),
		RecordedAt:       recordedAt,
	}
	if err := u.repo.Create(ctx, &reading); err != nil {
		return domain.FeedReading{}, err
	}
	return reading, nil
}

func (u *FeedUsecase) UpdateDeviceStatus(ctx context.Context, input UpdateDeviceStatusInput) (domain.DeviceStatus, error) {
	deviceID := strings.TrimSpace(input.DeviceID)
	if deviceID == "" {
		return domain.DeviceStatus{}, fmt.Errorf("%w: device_id is required", domain.ErrValidation)
	}
	if input.FoodStockPercent == nil && input.WaterAvailable == nil && strings.TrimSpace(input.WaterStatus) == "" {
		return domain.DeviceStatus{}, fmt.Errorf("%w: at least one sensor field is required", domain.ErrValidation)
	}
	if input.FoodStockPercent != nil && (*input.FoodStockPercent < 0 || *input.FoodStockPercent > 100) {
		return domain.DeviceStatus{}, fmt.Errorf("%w: food_stock_percent must be between 0 and 100", domain.ErrValidation)
	}

	status, err := u.repo.UpsertDeviceStatus(ctx, domain.DeviceStatusUpdate{
		ID:               deviceID,
		FoodStockPercent: input.FoodStockPercent,
		WaterAvailable:   input.WaterAvailable,
		WaterStatus:      strings.TrimSpace(input.WaterStatus),
		LastSeenAt:       time.Now().UTC(),
	})
	if err != nil {
		return domain.DeviceStatus{}, err
	}
	return normalizeDeviceStatus(status), nil
}

func (u *FeedUsecase) ListHistory(ctx context.Context, filter domain.FeedHistoryFilter) ([]domain.FeedReading, error) {
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return nil, fmt.Errorf("%w: from must be before to", domain.ErrValidation)
	}
	if filter.Limit <= 0 {
		filter.Limit = u.defaultLimit
	}
	if filter.Limit > u.maxLimit {
		filter.Limit = u.maxLimit
	}
	return u.repo.ListHistory(ctx, filter)
}
