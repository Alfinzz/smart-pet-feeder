package usecase

import (
	"testing"

	"smart-pet-monitoring/backend/internal/domain"
)

func TestCalculateSAWHealthScoreUsesManualVitals(t *testing.T) {
	score := CalculateSAWHealthScore(&domain.VitalSigns{
		WeightKG:        10,
		ActivityMinutes: 30,
		SleepHours:      10,
	}, 10, 0)

	if score.Score != 100 {
		t.Fatalf("score = %d, want 100", score.Score)
	}
	if score.WeightComponent != 100 || score.ActivityComponent != 100 || score.SleepComponent != 100 {
		t.Fatalf("components = %.1f %.1f %.1f, want all 100", score.WeightComponent, score.ActivityComponent, score.SleepComponent)
	}
}

func TestCalculateSAWHealthScoreAppliesOverdueTaskPenalty(t *testing.T) {
	score := CalculateSAWHealthScore(&domain.VitalSigns{
		WeightKG:        10,
		ActivityMinutes: 30,
		SleepHours:      10,
	}, 10, 1)

	if score.Score != 85 {
		t.Fatalf("score = %d, want 85", score.Score)
	}
	if score.TaskPenalty != 15 {
		t.Fatalf("penalty = %d, want 15", score.TaskPenalty)
	}
}
