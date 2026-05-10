package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/account/domain"
	categoryDomain "Finance-Manager-System/internal/infrastructure/modules/category/domain"
	"Finance-Manager-System/internal/infrastructure/modules/tbankpdf"
	transactionDomain "Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type AccountRepository interface {
	AddAccount(ctx context.Context, acc *domain.Account) (uuid.UUID, error)
	ArchiveAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error
	GetAllAccountsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	ChangeNameAccount(ctx context.Context, name string, userID uuid.UUID, accountID uuid.UUID) error
}

type AccountCategoryRepository interface {
	GetCategoriesByUser(ctx context.Context, userID uuid.UUID) ([]categoryDomain.Category, error)
}

type AccountTransactionRepository interface {
	AddTransactions(ctx context.Context, transactions []*transactionDomain.Transaction) error
}

type AccountUseCase struct {
	repo      AccountRepository
	catRepo   AccountCategoryRepository
	transRepo AccountTransactionRepository
	txManager database.TxManager
}

func NewAccountUseCase(
	repo AccountRepository,
	catRepo AccountCategoryRepository,
	transRepo AccountTransactionRepository,
	txManager database.TxManager,
) *AccountUseCase {
	return &AccountUseCase{
		repo:      repo,
		catRepo:   catRepo,
		transRepo: transRepo,
		txManager: txManager,
	}
}

type ImportPDFResult struct {
	AccountID            uuid.UUID `json:"account_id"`
	ImportedTransactions int       `json:"imported_transactions"`
	Balance              int64     `json:"balance"`
	AccountNumber        string    `json:"account_number,omitempty"`
	ContractNumber       string    `json:"contract_number,omitempty"`
}

func (uc *AccountUseCase) ImportAccountFromTBankPDF(ctx context.Context, userID uuid.UUID, customName string, pdfData []byte) (*ImportPDFResult, error) {
	statement, err := tbankpdf.ParseStatement(pdfData)
	if err != nil {
		return nil, err
	}

	if statement.AccountNumber == "" {
		statement.AccountNumber = statement.ContractNumber
	}
	if statement.AccountNumber == "" {
		statement.AccountNumber = "tbank-pdf"
	}

	accountName := customName
	if accountName == "" {
		accountName = buildImportedAccountName(statement)
	}

	var result ImportPDFResult
	err = uc.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		categories, catErr := uc.catRepo.GetCategoriesByUser(txCtx, userID)
		if catErr != nil {
			return catErr
		}

		externalID := statement.AccountNumber
		acc, accErr := domain.NewAccount(
			userID,
			accountName,
			"RUB",
			"imported_pdf",
			"#FFDD2D",
			true,
			&externalID,
		)
		if accErr != nil {
			return fmt.Errorf("validation failed: %w", accErr)
		}
		acc.Balance = statement.Balance

		accountID, addErr := uc.repo.AddAccount(txCtx, acc)
		if addErr != nil {
			return fmt.Errorf("failed to save account: %w", addErr)
		}

		trans := make([]*transactionDomain.Transaction, 0, len(statement.Transactions))
		for _, rawTx := range statement.Transactions {
			categoryID := resolveCategoryID(categories, rawTx.Description, rawTx.IsIncome)
			tx, txErr := transactionDomain.NewTransaction(
				userID,
				accountID,
				categoryID,
				rawTx.Description,
				rawTx.IsIncome,
				rawTx.Amount,
				rawTx.CompletedAt,
				true,
				nil,
			)
			if txErr != nil {
				continue
			}
			trans = append(trans, tx)
		}

		if len(trans) > 0 {
			if insertErr := uc.transRepo.AddTransactions(txCtx, trans); insertErr != nil {
				return fmt.Errorf("failed to import transactions: %w", insertErr)
			}
		}

		result = ImportPDFResult{
			AccountID:            accountID,
			ImportedTransactions: len(trans),
			Balance:              statement.Balance,
			AccountNumber:        statement.AccountNumber,
			ContractNumber:       statement.ContractNumber,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
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

	_, err = uc.repo.AddAccount(ctx, acc)
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
