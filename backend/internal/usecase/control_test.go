package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"smart-pet-monitoring/backend/internal/domain"
)

type fakeCommandRepository struct {
	ownerDeviceID   string
	hasDevice       bool
	createdCommand  *domain.ManualCommand
	nextStaleBefore time.Time
	nextMaxAttempts int
	statusLastError string
}

func (r *fakeCommandRepository) Create(ctx context.Context, command *domain.ManualCommand) error {
	r.createdCommand = command
	command.ID = 10
	command.CreatedAt = time.Now()
	command.UpdatedAt = command.CreatedAt
	return nil
}

func (r *fakeCommandRepository) GetByOwner(ctx context.Context, ownerID, commandID int64) (domain.ManualCommand, error) {
	return domain.ManualCommand{ID: commandID, OwnerID: ownerID}, nil
}

func (r *fakeCommandRepository) GetNextQueued(ctx context.Context, deviceID string, staleBefore time.Time, maxAttempts int) (domain.ManualCommand, error) {
	r.nextStaleBefore = staleBefore
	r.nextMaxAttempts = maxAttempts
	return domain.ManualCommand{ID: 1, DeviceID: deviceID, Status: domain.CommandStatusSent}, nil
}

func (r *fakeCommandRepository) GetOwnerDeviceID(ctx context.Context, ownerID int64) (string, error) {
	if r.ownerDeviceID == "" {
		return "", domain.ErrNotFound
	}
	return r.ownerDeviceID, nil
}

func (r *fakeCommandRepository) OwnerHasDevice(ctx context.Context, ownerID int64, deviceID string) (bool, error) {
	return r.hasDevice, nil
}

func (r *fakeCommandRepository) UpdateStatus(ctx context.Context, deviceID string, commandID int64, status domain.CommandStatus, lastError string) (domain.ManualCommand, error) {
	r.statusLastError = lastError
	return domain.ManualCommand{ID: commandID, DeviceID: deviceID, Status: status, LastError: lastError}, nil
}

func TestCreateManualCommandUsesOwnerDeviceWhenMissing(t *testing.T) {
	repo := &fakeCommandRepository{ownerDeviceID: "ESP32-001"}
	usecase := NewControlUsecase(repo)

	command, err := usecase.CreateManualCommand(context.Background(), CreateManualCommandInput{
		OwnerID: 1,
		Action:  domain.CommandActionFeed,
	})
	if err != nil {
		t.Fatalf("CreateManualCommand returned error: %v", err)
	}
	if command.DeviceID != "ESP32-001" {
		t.Fatalf("device id = %q, want ESP32-001", command.DeviceID)
	}
}

func TestCreateManualCommandRejectsForeignDevice(t *testing.T) {
	repo := &fakeCommandRepository{hasDevice: false}
	usecase := NewControlUsecase(repo)

	_, err := usecase.CreateManualCommand(context.Background(), CreateManualCommandInput{
		OwnerID:  1,
		DeviceID: "OTHER",
		Action:   domain.CommandActionFeed,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
	if repo.createdCommand != nil {
		t.Fatal("command should not be created for foreign device")
	}
}

func TestGetNextCommandUsesRetryDefaults(t *testing.T) {
	repo := &fakeCommandRepository{}
	usecase := NewControlUsecase(repo)

	if _, err := usecase.GetNextCommand(context.Background(), "ESP32-001"); err != nil {
		t.Fatalf("GetNextCommand returned error: %v", err)
	}
	if repo.nextMaxAttempts != 3 {
		t.Fatalf("max attempts = %d, want 3", repo.nextMaxAttempts)
	}
	if time.Since(repo.nextStaleBefore) < 89*time.Second {
		t.Fatalf("stale before = %v, want about 90 seconds ago", repo.nextStaleBefore)
	}
}

func TestUpdateCommandStatusTrimsLastError(t *testing.T) {
	repo := &fakeCommandRepository{}
	usecase := NewControlUsecase(repo)

	_, err := usecase.UpdateCommandStatus(context.Background(), UpdateCommandStatusInput{
		DeviceID:  "ESP32-001",
		CommandID: 1,
		Status:    domain.CommandStatusFailed,
		LastError: "  load cell pakan tidak siap  ",
	})
	if err != nil {
		t.Fatalf("UpdateCommandStatus returned error: %v", err)
	}
	if repo.statusLastError != "load cell pakan tidak siap" {
		t.Fatalf("last error = %q", repo.statusLastError)
	}
}
