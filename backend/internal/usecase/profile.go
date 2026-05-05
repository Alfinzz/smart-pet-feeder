package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"smart-pet-monitoring/backend/internal/domain"
)

type ProfileRepository interface {
	CalibrateDevice(ctx context.Context, ownerID int64) (domain.DeviceSettings, error)
	CreatePetDetails(ctx context.Context, ownerID int64, input domain.PetDetailsInput) (domain.PetDetails, error)
	DeletePetDetails(ctx context.Context, ownerID int64) error
	GetDeviceSettings(ctx context.Context, ownerID int64) (domain.DeviceSettings, error)
	GetNotificationPreferences(ctx context.Context, ownerID int64) (domain.NotificationPreferences, error)
	GetPetDetails(ctx context.Context, ownerID int64) (domain.PetDetails, error)
	UpdateDeviceSettings(ctx context.Context, ownerID int64, input domain.DeviceSettingsInput) (domain.DeviceSettings, error)
	UpdatePetDetails(ctx context.Context, ownerID int64, input domain.PetDetailsInput) (domain.PetDetails, error)
	UpsertNotificationPreferences(ctx context.Context, ownerID int64, input domain.NotificationPreferencesInput) (domain.NotificationPreferences, error)
}

type ProfileUsecase struct {
	repo ProfileRepository
}

func NewProfileUsecase(repo ProfileRepository) *ProfileUsecase {
	return &ProfileUsecase{repo: repo}
}

func (u *ProfileUsecase) GetPetDetails(ctx context.Context, ownerID int64) (domain.PetDetails, error) {
	return u.repo.GetPetDetails(ctx, ownerID)
}

func (u *ProfileUsecase) CreatePetDetails(ctx context.Context, ownerID int64, input domain.PetDetailsInput) (domain.PetDetails, error) {
	if _, err := u.repo.GetPetDetails(ctx, ownerID); err == nil {
		return domain.PetDetails{}, fmt.Errorf("%w: pet details already exist", domain.ErrConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.PetDetails{}, err
	}

	input, err := validatePetDetailsInput(input)
	if err != nil {
		return domain.PetDetails{}, err
	}
	return u.repo.CreatePetDetails(ctx, ownerID, input)
}

func (u *ProfileUsecase) UpdatePetDetails(ctx context.Context, ownerID int64, input domain.PetDetailsInput) (domain.PetDetails, error) {
	input, err := validatePetDetailsInput(input)
	if err != nil {
		return domain.PetDetails{}, err
	}
	return u.repo.UpdatePetDetails(ctx, ownerID, input)
}

func (u *ProfileUsecase) DeletePetDetails(ctx context.Context, ownerID int64) error {
	if ownerID <= 0 {
		return domain.ErrUnauthorized
	}
	return u.repo.DeletePetDetails(ctx, ownerID)
}

func (u *ProfileUsecase) GetDeviceSettings(ctx context.Context, ownerID int64) (domain.DeviceSettings, error) {
	return u.repo.GetDeviceSettings(ctx, ownerID)
}

func (u *ProfileUsecase) UpdateDeviceSettings(ctx context.Context, ownerID int64, input domain.DeviceSettingsInput) (domain.DeviceSettings, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.ManualFeedPortionGrams < 0 {
		return domain.DeviceSettings{}, fmt.Errorf("%w: manual_feed_portion_grams must be greater than or equal to 0", domain.ErrValidation)
	}
	if input.Name == "" && input.ManualFeedPortionGrams == 0 {
		return domain.DeviceSettings{}, fmt.Errorf("%w: at least one device setting is required", domain.ErrValidation)
	}
	return u.repo.UpdateDeviceSettings(ctx, ownerID, input)
}

func (u *ProfileUsecase) CalibrateDevice(ctx context.Context, ownerID int64) (domain.DeviceSettings, error) {
	if ownerID <= 0 {
		return domain.DeviceSettings{}, domain.ErrUnauthorized
	}
	return u.repo.CalibrateDevice(ctx, ownerID)
}

func (u *ProfileUsecase) GetNotificationPreferences(ctx context.Context, ownerID int64) (domain.NotificationPreferences, error) {
	preferences, err := u.repo.GetNotificationPreferences(ctx, ownerID)
	if errors.Is(err, domain.ErrNotFound) {
		return u.repo.UpsertNotificationPreferences(ctx, ownerID, domain.NotificationPreferencesInput{
			LowFoodAlert:         true,
			EmptyWaterAlert:      true,
			FeedingSuccessReport: true,
		})
	}
	return preferences, err
}

func (u *ProfileUsecase) UpsertNotificationPreferences(ctx context.Context, ownerID int64, input domain.NotificationPreferencesInput) (domain.NotificationPreferences, error) {
	if ownerID <= 0 {
		return domain.NotificationPreferences{}, domain.ErrUnauthorized
	}
	return u.repo.UpsertNotificationPreferences(ctx, ownerID, input)
}

func validatePetDetailsInput(input domain.PetDetailsInput) (domain.PetDetailsInput, error) {
	input.DeviceID = strings.TrimSpace(input.DeviceID)
	input.Name = strings.TrimSpace(input.Name)
	input.Species = strings.TrimSpace(input.Species)
	input.Breed = strings.TrimSpace(input.Breed)
	input.Gender = strings.TrimSpace(input.Gender)

	if input.Name == "" {
		return domain.PetDetailsInput{}, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if input.Species == "" {
		return domain.PetDetailsInput{}, fmt.Errorf("%w: species is required", domain.ErrValidation)
	}
	if input.Breed == "" {
		return domain.PetDetailsInput{}, fmt.Errorf("%w: breed is required", domain.ErrValidation)
	}
	if input.Gender == "" {
		input.Gender = "unknown"
	}
	if input.DeviceID == "" {
		input.DeviceID = defaultDeviceID
	}
	if input.DailyFeedTargetGrams <= 0 {
		return domain.PetDetailsInput{}, fmt.Errorf("%w: daily_feed_target_grams must be greater than 0", domain.ErrValidation)
	}
	return input, nil
}
