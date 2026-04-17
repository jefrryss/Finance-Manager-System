package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/account/domain"
)

type AccountRepo struct {
	db *sqlx.DB
}

func NewAccountRepo(db *sqlx.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, amountDelta int64) error {
	q := database.GetQueryer(ctx, r.db)
	query := `UPDATE Accounts SET balance = balance + $1 WHERE user_id = $2 AND account_id = $3 AND is_archived = false`

	result, err := q.ExecContext(ctx, query, amountDelta, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("account not found, belongs to another user, or archived")
	}

	return nil
}

func (r *AccountRepo) GetAllAccountsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	q := database.GetQueryer(ctx, r.db)
	accounts := make([]domain.Account, 0)

	query := `
        SELECT * FROM Accounts 
        WHERE user_id = $1 AND is_archived = false
        ORDER BY created_at ASC
    `

	err := q.SelectContext(ctx, &accounts, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user accounts: %w", err)
	}

	return accounts, nil
}

func (r *AccountRepo) ChangeNameAccount(ctx context.Context, name string, userID uuid.UUID, accountID uuid.UUID) error {
	q := database.GetQueryer(ctx, r.db)
	query := `
        UPDATE Accounts 
        SET name_account = $1 
        WHERE user_id = $2 AND account_id = $3 AND is_archived = false
    `

	result, err := q.ExecContext(ctx, query, name, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to change account name: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found or archived")
	}

	return nil
}

func (r *AccountRepo) AddAccount(ctx context.Context, acc *domain.Account) (uuid.UUID, error) {
	q := database.GetQueryer(ctx, r.db)
	if err := r.ensureAccountsSchema(ctx, q); err != nil {
		return uuid.Nil, err
	}
	query := `
        INSERT INTO Accounts (
            user_id, balance, is_imported, external_account_id, 
            account_type, color_hex, is_archived, name_account, 
            currency, last_synced_at, created_at
        ) 
        VALUES (
            :user_id, :balance, :is_imported, :external_account_id, 
            :account_type, :color_hex, :is_archived, :name_account, 
            :currency, :last_synced_at, :created_at
        )
        RETURNING account_id
    `

	queryStr, args, err := sqlx.Named(query, acc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to process named query: %w", err)
	}
	queryStr = q.Rebind(queryStr)

	var accountID uuid.UUID
	if err := q.QueryRowContext(ctx, queryStr, args...).Scan(&accountID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add account: %w", err)
	}

	return accountID, nil
}

func (r *AccountRepo) ensureAccountsSchema(ctx context.Context, q database.Queryer) error {
	queries := []string{
		`ALTER TABLE Accounts ADD COLUMN IF NOT EXISTS external_account_id VARCHAR(255)`,
		`ALTER TABLE Accounts ADD COLUMN IF NOT EXISTS account_type VARCHAR(50)`,
		`ALTER TABLE Accounts ADD COLUMN IF NOT EXISTS color_hex VARCHAR(7)`,
		`ALTER TABLE Accounts ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE Accounts ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ`,
		`ALTER TABLE Accounts ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'RUB'`,
	}

	for _, query := range queries {
		if _, err := q.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to ensure accounts schema: %w", err)
		}
	}
	return nil
}

func (r *AccountRepo) ArchiveAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error {
	q := database.GetQueryer(ctx, r.db)
	query := `
        UPDATE Accounts 
        SET is_archived = true
        WHERE user_id = $1 AND account_id = $2 AND is_archived = false
    `

	result, err := q.ExecContext(ctx, query, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to archive account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found or already archived")
	}

	return nil
}
