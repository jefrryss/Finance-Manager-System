package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"Finance-Manager-System/internal/infrastructure/database"
	"Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
)

type TransRepository struct {
	db *sqlx.DB
}

func NewTransRepository(db *sqlx.DB) *TransRepository {
	return &TransRepository{db: db}
}

func (tr *TransRepository) MoveTransactionsCategory(ctx context.Context, userID uuid.UUID, oldCategoryID uuid.UUID, newCategoryID uuid.UUID) error {
	q := database.GetQueryer(ctx, tr.db)
	query := `UPDATE Transactions SET category_id = $1 WHERE category_id = $2 AND user_id = $3`
	_, err := q.ExecContext(ctx, query, newCategoryID, oldCategoryID, userID)
	return err
}

func (tr *TransRepository) AddTransactions(ctx context.Context, transactions []*domain.Transaction) error {
	q := database.GetQueryer(ctx, tr.db)

	query := `
        INSERT INTO Transactions (
            user_id, account_id, category_id, name_transaction, 
            is_income, amount, completed_at, is_hidden, is_imported, comment
        ) 
        VALUES (
            :user_id, :account_id, :category_id, :name_transaction, 
            :is_income, :amount, :completed_at, :is_hidden, :is_imported, :comment
        )
    `

	_, err := q.NamedExecContext(ctx, query, transactions)
	if err != nil {
		return fmt.Errorf("ошибка NamedExecContext: %w", err)
	}
	return nil
}

func (tr *TransRepository) ShowTransactions(ctx context.Context, userId uuid.UUID, transactionIds []uuid.UUID) error {
	if len(transactionIds) == 0 {
		return nil
	}
	q := database.GetQueryer(ctx, tr.db)

	query := `UPDATE Transactions SET is_hidden = false WHERE user_id = ? AND transaction_id IN (?)`
	query, args, err := sqlx.In(query, userId, transactionIds)
	if err != nil {
		return fmt.Errorf("ошибка формирования In-запроса: %w", err)
	}

	query = q.Rebind(query)

	_, err = q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка выполнения UPDATE: %w", err)
	}
	return nil
}

func (tr *TransRepository) HideTransactions(ctx context.Context, userId uuid.UUID, transactionIds []uuid.UUID) error {
	if len(transactionIds) == 0 {
		return nil
	}
	q := database.GetQueryer(ctx, tr.db)

	query := `UPDATE Transactions SET is_hidden = true WHERE user_id = ? AND transaction_id IN (?)`
	query, args, err := sqlx.In(query, userId, transactionIds)
	if err != nil {
		return fmt.Errorf("ошибка формирования In-запроса: %w", err)
	}

	query = q.Rebind(query)

	_, err = q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка выполнения UPDATE: %w", err)
	}
	return nil
}

func (tr *TransRepository) GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*domain.Transaction, error) {
	q := database.GetQueryer(ctx, tr.db)
	var trans domain.Transaction
	query := `SELECT * FROM Transactions WHERE user_id = $1 AND transaction_id = $2`

	err := q.GetContext(ctx, &trans, query, userID, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &trans, nil
}

func (tr *TransRepository) AddTransaction(ctx context.Context, trans *domain.Transaction) error {
	q := database.GetQueryer(ctx, tr.db)
	query := `
        INSERT INTO Transactions (
            user_id, account_id, category_id, name_transaction, 
            is_income, amount, completed_at, is_hidden, is_imported, comment
        ) 
        VALUES (
            :user_id, :account_id, :category_id, :name_transaction, 
            :is_income, :amount, :completed_at, :is_hidden, :is_imported, :comment
        )
    `

	_, err := q.NamedExecContext(ctx, query, trans)
	if err != nil {
		return fmt.Errorf("failed to add transaction: %w", err)
	}

	return nil
}

func (tr *TransRepository) DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	q := database.GetQueryer(ctx, tr.db)
	query := `DELETE FROM Transactions WHERE user_id = $1 AND transaction_id = $2`

	result, err := q.ExecContext(ctx, query, userID, transactionID)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

func (tr *TransRepository) GetAllTransactions(ctx context.Context, userID uuid.UUID) ([]domain.Transaction, error) {
	q := database.GetQueryer(ctx, tr.db)
	transactions := make([]domain.Transaction, 0)

	query := `
        SELECT * FROM Transactions 
        WHERE user_id = $1 
        ORDER BY completed_at DESC
    `

	err := q.SelectContext(ctx, &transactions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all transactions: %w", err)
	}

	return transactions, nil
}

func (tr *TransRepository) GetTransactionsByIDs(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID) ([]domain.Transaction, error) {
	if len(transactionIDs) == 0 {
		return nil, nil
	}
	q := database.GetQueryer(ctx, tr.db)

	transactions := make([]domain.Transaction, 0)
	query := `SELECT * FROM Transactions WHERE user_id = ? AND transaction_id IN (?)`

	query, args, err := sqlx.In(query, userID, transactionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build IN query: %w", err)
	}

	query = q.Rebind(query)

	err = q.SelectContext(ctx, &transactions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by IDs: %w", err)
	}

	return transactions, nil
}
