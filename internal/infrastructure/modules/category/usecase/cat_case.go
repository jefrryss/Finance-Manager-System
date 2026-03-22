package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/category/domain"
)

var (
	ErrCannotModifyDefaultCategory = errors.New("cannot modify or delete default categories")
)

type CategoryRepository interface {
	AddCategory(ctx context.Context, category *domain.Category) (uuid.UUID, error)
	GetCategoriesByUser(ctx context.Context, userID uuid.UUID) ([]domain.Category, error)
	GetCategoryByID(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) (*domain.Category, error)
	UpdateCategory(ctx context.Context, categoryID uuid.UUID, userID uuid.UUID, newName string, newIconURL *string) error
	DeleteCategory(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) error
}

type TransactionCategoryUpdater interface {
	MoveTransactionsCategory(ctx context.Context, userID uuid.UUID, oldCategoryID uuid.UUID, newCategoryID uuid.UUID) error
}

type CategoryUseCase struct {
	catRepo   CategoryRepository
	transRepo TransactionCategoryUpdater
	txManager database.TxManager
}

func NewCategoryUseCase(cr CategoryRepository, tr TransactionCategoryUpdater, tm database.TxManager) *CategoryUseCase {
	return &CategoryUseCase{
		catRepo:   cr,
		transRepo: tr,
		txManager: tm,
	}
}

func (uc *CategoryUseCase) CreateCustomCategory(ctx context.Context, userID uuid.UUID, name string, isIncome bool, iconURL *string) (uuid.UUID, error) {
	cat, err := domain.NewCategory(userID, name, isIncome, true, iconURL)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validation failed: %w", err)
	}

	generatedID, err := uc.catRepo.AddCategory(ctx, cat)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to save category: %w", err)
	}

	return generatedID, nil
}

func (uc *CategoryUseCase) GetUserCategories(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrCatEmptyUserID
	}

	categories, err := uc.catRepo.GetCategoriesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	return categories, nil
}

func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, newName string, newIconURL *string) error {
	cat, err := uc.catRepo.GetCategoryByID(ctx, userID, categoryID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	if !cat.IsCustom {
		return ErrCannotModifyDefaultCategory
	}

	if _, err := domain.NewCategory(userID, newName, cat.IsIncome, cat.IsCustom, newIconURL); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := uc.catRepo.UpdateCategory(ctx, categoryID, userID, newName, newIconURL); err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, replacementCategoryID *uuid.UUID) error {
	return uc.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		cat, err := uc.catRepo.GetCategoryByID(ctx, userID, categoryID)
		if err != nil {
			return fmt.Errorf("category not found: %w", err)
		}

		if !cat.IsCustom {
			return ErrCannotModifyDefaultCategory
		}

		if replacementCategoryID != nil {
			replacementCat, err := uc.catRepo.GetCategoryByID(ctx, userID, *replacementCategoryID)
			if err != nil {
				return fmt.Errorf("replacement category not found: %w", err)
			}
			if replacementCat.IsIncome != cat.IsIncome {
				return fmt.Errorf("cannot move transactions to a category of a different type")
			}

			if err := uc.transRepo.MoveTransactionsCategory(ctx, userID, categoryID, *replacementCategoryID); err != nil {
				return fmt.Errorf("failed to move transactions: %w", err)
			}
		}

		if err := uc.catRepo.DeleteCategory(ctx, userID, categoryID); err != nil {
			return fmt.Errorf("failed to delete category: %w", err)
		}

		return nil
	})
}
