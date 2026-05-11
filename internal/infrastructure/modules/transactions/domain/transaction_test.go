package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTransactionDefaults(t *testing.T) {
	tr, err := NewTransaction(uuid.New(), uuid.New(), nil, "Coffee", false, 25000, time.Time{}, true, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tr.Currency != "RUB" {
		t.Fatalf("unexpected default currency: %s", tr.Currency)
	}
	if tr.Status != "completed" {
		t.Fatalf("unexpected default status: %s", tr.Status)
	}
	if tr.CompletedAt.IsZero() {
		t.Fatalf("completed_at must be auto-filled")
	}
}

func TestNewTransactionInvalidAmount(t *testing.T) {
	_, err := NewTransaction(uuid.New(), uuid.New(), nil, "Bad", true, 0, time.Now(), false, nil)
	if err != ErrTransInvalidAmount {
		t.Fatalf("expected ErrTransInvalidAmount, got %v", err)
	}
}

