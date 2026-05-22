package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"smart-pet-monitoring/backend/internal/domain"
)

type CommandRepository interface {
	Create(ctx context.Context, command *domain.ManualCommand) error
	GetByOwner(ctx context.Context, ownerID, commandID int64) (domain.ManualCommand, error)
	GetNextQueued(ctx context.Context, deviceID string, staleBefore time.Time, maxAttempts int) (domain.ManualCommand, error)
	GetOwnerDeviceID(ctx context.Context, ownerID int64) (string, error)
	OwnerHasDevice(ctx context.Context, ownerID int64, deviceID string) (bool, error)
	UpdateStatus(ctx context.Context, deviceID string, commandID int64, status domain.CommandStatus, lastError string) (domain.ManualCommand, error)
}

type ControlUsecase struct {
	repo                CommandRepository
	deliveryTimeout     time.Duration
	maxDeliveryAttempts int
}

type CreateManualCommandInput struct {
	OwnerID  int64
	DeviceID string
	Action   domain.CommandAction
}

type UpdateCommandStatusInput struct {
	DeviceID  string
	CommandID int64
	Status    domain.CommandStatus
	LastError string
}

func NewControlUsecase(repo CommandRepository) *ControlUsecase {
	return &ControlUsecase{
		repo:                repo,
		deliveryTimeout:     90 * time.Second,
		maxDeliveryAttempts: 3,
	}
}

func (u *ControlUsecase) GetNextCommand(ctx context.Context, deviceID string) (domain.ManualCommand, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return domain.ManualCommand{}, fmt.Errorf("%w: device_id is required", domain.ErrValidation)
	}
	return u.repo.GetNextQueued(ctx, deviceID, time.Now().Add(-u.deliveryTimeout), u.maxDeliveryAttempts)
}

func (u *ControlUsecase) UpdateCommandStatus(ctx context.Context, input UpdateCommandStatusInput) (domain.ManualCommand, error) {
	deviceID := strings.TrimSpace(input.DeviceID)
	if deviceID == "" {
		return domain.ManualCommand{}, fmt.Errorf("%w: device_id is required", domain.ErrValidation)
	}
	if input.CommandID <= 0 {
		return domain.ManualCommand{}, fmt.Errorf("%w: command_id is required", domain.ErrValidation)
	}
	if !input.Status.ValidDeviceUpdate() {
		return domain.ManualCommand{}, fmt.Errorf("%w: status must be completed or failed", domain.ErrValidation)
	}
	return u.repo.UpdateStatus(ctx, deviceID, input.CommandID, input.Status, strings.TrimSpace(input.LastError))
}

func (u *ControlUsecase) GetManualCommand(ctx context.Context, ownerID, commandID int64) (domain.ManualCommand, error) {
	if ownerID <= 0 {
		return domain.ManualCommand{}, domain.ErrUnauthorized
	}
	if commandID <= 0 {
		return domain.ManualCommand{}, fmt.Errorf("%w: command_id is required", domain.ErrValidation)
	}
	return u.repo.GetByOwner(ctx, ownerID, commandID)
}

func (u *ControlUsecase) CreateManualCommand(ctx context.Context, input CreateManualCommandInput) (domain.ManualCommand, error) {
	deviceID := strings.TrimSpace(input.DeviceID)
	if input.OwnerID <= 0 {
		return domain.ManualCommand{}, domain.ErrUnauthorized
	}
	if deviceID == "" {
		var err error
		deviceID, err = u.repo.GetOwnerDeviceID(ctx, input.OwnerID)
		if err != nil {
			return domain.ManualCommand{}, err
		}
	} else {
		ok, err := u.repo.OwnerHasDevice(ctx, input.OwnerID, deviceID)
		if err != nil {
			return domain.ManualCommand{}, err
		}
		if !ok {
			return domain.ManualCommand{}, domain.ErrNotFound
		}
	}
	if !input.Action.Valid() {
		return domain.ManualCommand{}, fmt.Errorf("%w: action must be feed or drink", domain.ErrValidation)
	}

	command := domain.ManualCommand{
		OwnerID:  input.OwnerID,
		DeviceID: deviceID,
		Action:   input.Action,
		Status:   domain.CommandStatusQueued,
	}
	if err := u.repo.Create(ctx, &command); err != nil {
		return domain.ManualCommand{}, err
	}
	return command, nil
}
