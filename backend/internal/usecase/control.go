package usecase

import (
	"context"
	"fmt"
	"strings"

	"smart-pet-monitoring/backend/internal/domain"
)

type CommandRepository interface {
	Create(ctx context.Context, command *domain.ManualCommand) error
	GetNextQueued(ctx context.Context, deviceID string) (domain.ManualCommand, error)
	UpdateStatus(ctx context.Context, deviceID string, commandID int64, status domain.CommandStatus) (domain.ManualCommand, error)
}

type ControlUsecase struct {
	repo CommandRepository
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
}

func NewControlUsecase(repo CommandRepository) *ControlUsecase {
	return &ControlUsecase{repo: repo}
}

func (u *ControlUsecase) GetNextCommand(ctx context.Context, deviceID string) (domain.ManualCommand, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return domain.ManualCommand{}, fmt.Errorf("%w: device_id is required", domain.ErrValidation)
	}
	return u.repo.GetNextQueued(ctx, deviceID)
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
	return u.repo.UpdateStatus(ctx, deviceID, input.CommandID, input.Status)
}

func (u *ControlUsecase) CreateManualCommand(ctx context.Context, input CreateManualCommandInput) (domain.ManualCommand, error) {
	deviceID := strings.TrimSpace(input.DeviceID)
	if input.OwnerID <= 0 {
		return domain.ManualCommand{}, domain.ErrUnauthorized
	}
	if deviceID == "" {
		return domain.ManualCommand{}, fmt.Errorf("%w: device_id is required", domain.ErrValidation)
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
