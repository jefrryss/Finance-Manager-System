package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/modules/account/domain"
)

type AccountRepository interface {
	AddAccount(ctx context.Context, acc *domain.Account) error
	ArchiveAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error
	GetAllAccountsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	ChangeNameAccount(ctx context.Context, name string, userID uuid.UUID, accountID uuid.UUID) error
}

type AccountUseCase struct {
	repo AccountRepository
}

func NewAccountUseCase(repo AccountRepository) *AccountUseCase {
	return &AccountUseCase{
		repo: repo,
	}
}

func (uc *AccountUseCase) CreateAccount(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	currency string,
	accountType string,
	colorHex string,
	isImported bool,
	externalAccountID *string,
) error {
	acc, err := domain.NewAccount(
		userID,
		name,
		currency,
		accountType,
		colorHex,
		isImported,
		externalAccountID,
	)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	err = uc.repo.AddAccount(ctx, acc)
	if err != nil {
		return fmt.Errorf("failed to save account: %w", err)
	}

	return nil
}

func (uc *AccountUseCase) GetUserAccounts(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrEmptyUserID
	}

	accounts, err := uc.repo.GetAllAccountsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	return accounts, nil
}

func (uc *AccountUseCase) RenameAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, newName string) error {
	if userID == uuid.Nil || accountID == uuid.Nil {
		return fmt.Errorf("user ID and account ID cannot be empty")
	}

	if newName == "" {
		return domain.ErrEmptyAccountName
	}

	err := uc.repo.ChangeNameAccount(ctx, newName, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to rename account: %w", err)
	}

	return nil
}

func (uc *AccountUseCase) ArchiveAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error {
	if userID == uuid.Nil || accountID == uuid.Nil {
		return fmt.Errorf("user ID and account ID cannot be empty")
	}

	err := uc.repo.ArchiveAccount(ctx, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to archive account: %w", err)
	}

	return nil
}
