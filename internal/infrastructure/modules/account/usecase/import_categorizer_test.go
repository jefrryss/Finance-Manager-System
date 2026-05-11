package usecase

import (
	"testing"

	"github.com/google/uuid"

	categoryDomain "Finance-Manager-System/internal/infrastructure/modules/category/domain"
)

func TestResolveCategoryIDByMCCFallback(t *testing.T) {
	userID := uuid.New()
	cats := []categoryDomain.Category{
		{CategoryID: uuid.New(), UserID: userID, NameCategory: "Продукты", IsIncome: false},
		{CategoryID: uuid.New(), UserID: userID, NameCategory: "Другое", IsIncome: false},
	}
	mcc := "5411"
	got := resolveCategoryID(cats, "Random text", false, &mcc)
	if got == nil {
		t.Fatalf("expected category id")
	}
	if *got != cats[0].CategoryID {
		t.Fatalf("unexpected category id")
	}
}

func TestResolveCategoryIDByKeyword(t *testing.T) {
	userID := uuid.New()
	cats := []categoryDomain.Category{
		{CategoryID: uuid.New(), UserID: userID, NameCategory: "Кафе и рестораны", IsIncome: false},
	}
	got := resolveCategoryID(cats, "Burger place", false, nil)
	if got == nil || *got != cats[0].CategoryID {
		t.Fatalf("expected restaurants category")
	}
}

