package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewGoalSuccess(t *testing.T) {
	targetDate := time.Now().AddDate(0, 1, 0)
	goal, err := NewGoal(uuid.New(), "Phone", 100000, &targetDate)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if goal.NameGoal != "Phone" {
		t.Fatalf("unexpected name: %s", goal.NameGoal)
	}
	if goal.TargetAmount != 100000 {
		t.Fatalf("unexpected target amount: %d", goal.TargetAmount)
	}
}

func TestNewGoalContributionInvalid(t *testing.T) {
	_, err := NewGoalContribution(uuid.New(), uuid.New(), -1, nil, nil)
	if err != ErrGoalInvalidContributionAmount {
		t.Fatalf("expected ErrGoalInvalidContributionAmount, got %v", err)
	}
}

