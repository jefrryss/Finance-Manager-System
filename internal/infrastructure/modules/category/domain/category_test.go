package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewCategorySuccess(t *testing.T) {
	icon := " https://example.com/icon.svg "
	cat, err := NewCategory(uuid.New(), " Продукты ", false, true, &icon)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cat.NameCategory != "Продукты" {
		t.Fatalf("unexpected category name: %q", cat.NameCategory)
	}
	if cat.IconURL == nil || *cat.IconURL != "https://example.com/icon.svg" {
		t.Fatalf("unexpected icon value: %#v", cat.IconURL)
	}
}

func TestNewCategoryInvalidName(t *testing.T) {
	_, err := NewCategory(uuid.New(), "   ", false, false, nil)
	if err != ErrCatEmptyName {
		t.Fatalf("expected ErrCatEmptyName, got %v", err)
	}
}

