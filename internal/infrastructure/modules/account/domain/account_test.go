package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewAccountManualSuccess(t *testing.T) {
	userID := uuid.New()
	acc, err := NewAccount(userID, "Main", "rub", "manual", "#aa00cc", false, nil, 1500)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if acc.UserID != userID {
		t.Fatalf("unexpected user id")
	}
	if acc.Currency != "RUB" {
		t.Fatalf("unexpected currency: %s", acc.Currency)
	}
	if acc.AccountType != "MANUAL" {
		t.Fatalf("unexpected account type: %s", acc.AccountType)
	}
	if acc.ColorHex != "#AA00CC" {
		t.Fatalf("unexpected color: %s", acc.ColorHex)
	}
	if acc.Balance != 1500 {
		t.Fatalf("unexpected balance: %d", acc.Balance)
	}
	if acc.ExternalAccountID != nil {
		t.Fatalf("manual account must not have external id")
	}
}

func TestNewAccountImportedRequiresExternalID(t *testing.T) {
	userID := uuid.New()
	_, err := NewAccount(userID, "Imported", "RUB", "imported_pdf", "#FFDD2D", true, nil, 0)
	if err != ErrMissingExternalID {
		t.Fatalf("expected ErrMissingExternalID, got %v", err)
	}
}

