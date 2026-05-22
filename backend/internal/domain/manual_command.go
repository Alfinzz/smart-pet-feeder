package domain

import "time"

type CommandAction string

const (
	CommandActionFeed      CommandAction = "feed"
	CommandActionDrink     CommandAction = "drink"
	CommandActionServoTest CommandAction = "servo_test"
	CommandActionTare      CommandAction = "tare"
)

func (a CommandAction) Valid() bool {
	return a == CommandActionFeed ||
		a == CommandActionDrink ||
		a == CommandActionServoTest ||
		a == CommandActionTare
}

type CommandStatus string

const (
	CommandStatusQueued    CommandStatus = "queued"
	CommandStatusSent      CommandStatus = "sent"
	CommandStatusCompleted CommandStatus = "completed"
	CommandStatusFailed    CommandStatus = "failed"
)

func (s CommandStatus) ValidDeviceUpdate() bool {
	return s == CommandStatusCompleted || s == CommandStatusFailed
}

type ManualCommand struct {
	ID           int64
	OwnerID      int64
	DeviceID     string
	Action       CommandAction
	Status       CommandStatus
	AttemptCount int
	LastError    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}
