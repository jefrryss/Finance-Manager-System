package repository

import (
	"Finance-Manager-System/internal/infrastructure/modules/transactions/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TransRepository struct {
	db *sqlx.DB
}

func (tr *TransRepository) AddTransactions(ctx context.Context, transactions []*domain.Transaction) error {
	tx, err := tr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка BeginTxx: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO Transactions (user_id, transaction_id, account_id, category_id, name_transaction, is_income, amount, completed_at, is_hidden, is_imported, comment) 
		VALUES (:user_id, :transaction_id, :account_id, :category_id, :name_transaction, :is_income, :amount, :completed_at, :is_hidden, :is_imported, :comment)`

	_, err = tx.NamedExecContext(ctx, query, transactions)
	if err != nil {
		return fmt.Errorf("ошибка NamedExecContext: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка Commit Transactions: %w", err)
	}
	return nil
}

func (tr *TransRepository) ShowTransactions(ctx context.Context, userId uuid.UUID, transactionsIds []int) error {

	if len(transactionsIds) == 0 {
		return nil
	}
	tx, err := tr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error with changing visible in transactions: %w", err)
	}
	defer tx.Rollback()

	qeury := `UPDATE Transactions SET is_hidden = false WHERE user_id = ? AND transaction_id IN (?)`
	query, args, err := sqlx.In(qeury, userId, transactionsIds)

	query = tx.Rebind(query)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка выполнения UPDATE: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}
	return nil

}

func (tr *TransRepository) HideTransactions(ctx context.Context, userId uuid.UUID, transactionIds []int) error {

	if len(transactionIds) == 0 {
		return nil
	}

	tx, err := tr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error with changing visible in transactions: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE Transactions SET is_hidden = true WHERE user_id = ? AND transaction_id in (?)`

	query, args, err := sqlx.In(query, userId, transactionIds)
	if err != nil {
		return fmt.Errorf("ошибка формирования In-запроса: %w", err)
	}

	query = tx.Rebind(query)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка выполнения UPDATE: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}
	return nil
}
