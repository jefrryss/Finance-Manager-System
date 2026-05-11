package repository

import (
	"context"
	"fmt"
	"strings"
	"unicode"

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
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return err
	}
	query := `UPDATE Transactions SET category_id = $1 WHERE category_id = $2 AND user_id = $3`
	_, err := q.ExecContext(ctx, query, newCategoryID, oldCategoryID, userID)
	return err
}

func (tr *TransRepository) AddTransactions(ctx context.Context, transactions []*domain.Transaction) (int, error) {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return 0, err
	}
	query := `
        INSERT INTO Transactions (
            user_id, account_id, category_id, name_transaction, 
            is_income, amount, completed_at, is_hidden, is_imported, comment,
            sender_account, receiver_account, currency, bank_fee, status, external_transaction_id, mcc_code
        ) 
        VALUES (
            :user_id, :account_id, :category_id, :name_transaction, 
            :is_income, :amount, :completed_at, :is_hidden, :is_imported, :comment,
            :sender_account, :receiver_account, :currency, :bank_fee, :status, :external_transaction_id, :mcc_code
        )
        ON CONFLICT (user_id, account_id, external_transaction_id)
        WHERE external_transaction_id IS NOT NULL
        DO NOTHING
    `
	res, err := q.NamedExecContext(ctx, query, transactions)
	if err != nil {
		return 0, fmt.Errorf("ошибка NamedExecContext: %w", err)
	}
	rowsAffected, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", rowsErr)
	}
	return int(rowsAffected), nil
}

func (tr *TransRepository) ShowTransactions(ctx context.Context, userId uuid.UUID, transactionIds []uuid.UUID) error {
	if len(transactionIds) == 0 {
		return nil
	}
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return err
	}
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
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return err
	}
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
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return nil, err
	}
	var trans domain.Transaction
	query := `SELECT * FROM Transactions WHERE user_id = $1 AND transaction_id = $2`

	err := q.GetContext(ctx, &trans, query, userID, transactionID)
	if err != nil {
		return nil, domain.ErrTransNotFound
	}
	return &trans, nil
}

func (tr *TransRepository) AddTransaction(ctx context.Context, trans *domain.Transaction) error {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return err
	}
	query := `
        INSERT INTO Transactions (
            user_id, account_id, category_id, name_transaction, 
            is_income, amount, completed_at, is_hidden, is_imported, comment,
            sender_account, receiver_account, currency, bank_fee, status, external_transaction_id, mcc_code
        ) 
        VALUES (
            :user_id, :account_id, :category_id, :name_transaction, 
            :is_income, :amount, :completed_at, :is_hidden, :is_imported, :comment,
            :sender_account, :receiver_account, :currency, :bank_fee, :status, :external_transaction_id, :mcc_code
        )
        ON CONFLICT (user_id, account_id, external_transaction_id)
        WHERE external_transaction_id IS NOT NULL
        DO NOTHING
    `
	_, err := q.NamedExecContext(ctx, query, trans)
	if err != nil {
		return fmt.Errorf("failed to add transaction: %w", err)
	}
	return nil
}

func (tr *TransRepository) UpdateTransaction(ctx context.Context, trans *domain.Transaction) error {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return err
	}
	query := `
        UPDATE Transactions SET 
            category_id = :category_id, 
            name_transaction = :name_transaction, 
            is_income = :is_income,
            amount = :amount, 
            completed_at = :completed_at, 
            comment = :comment,
            sender_account = :sender_account,
            receiver_account = :receiver_account,
            currency = :currency,
            bank_fee = :bank_fee,
            status = :status,
            external_transaction_id = :external_transaction_id,
            mcc_code = :mcc_code
        WHERE transaction_id = :transaction_id AND user_id = :user_id
    `
	_, err := q.NamedExecContext(ctx, query, trans)
	return err
}

func (tr *TransRepository) DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return err
	}
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
		return domain.ErrTransNotFound
	}

	return nil
}

func (tr *TransRepository) GetAllTransactions(ctx context.Context, userID uuid.UUID) ([]domain.Transaction, error) {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return nil, err
	}
	transactions := make([]domain.Transaction, 0)

	query := `
        SELECT * FROM Transactions 
        WHERE user_id = $1 
        ORDER BY completed_at DESC, transaction_id DESC
    `

	err := q.SelectContext(ctx, &transactions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all transactions: %w", err)
	}
	return transactions, nil
}

func (tr *TransRepository) GetTransactionsWithFilter(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return nil, err
	}
	transactions := make([]domain.Transaction, 0)

	query := `SELECT * FROM Transactions WHERE user_id = $1`
	args := []interface{}{userID}
	argID := 2

	if filter.AccountID != nil {
		query += fmt.Sprintf(` AND account_id = $%d`, argID)
		args = append(args, *filter.AccountID)
		argID++
	}
	if filter.CategoryID != nil {
		query += fmt.Sprintf(` AND category_id = $%d`, argID)
		args = append(args, *filter.CategoryID)
		argID++
	}
	if filter.IsIncome != nil {
		query += fmt.Sprintf(` AND is_income = $%d`, argID)
		args = append(args, *filter.IsIncome)
		argID++
	}
	if filter.StartDate != nil {
		query += fmt.Sprintf(` AND completed_at >= $%d`, argID)
		args = append(args, *filter.StartDate)
		argID++
	}
	if filter.EndDate != nil {
		query += fmt.Sprintf(` AND completed_at <= $%d`, argID)
		args = append(args, *filter.EndDate)
		argID++
	}
	if filter.IsHidden != nil {
		query += fmt.Sprintf(` AND is_hidden = $%d`, argID)
		args = append(args, *filter.IsHidden)
		argID++
	} else if !filter.IncludeHidden {
		query += ` AND is_hidden = false`
	}
	if len(filter.AccountIDs) > 0 {
		placeholders := make([]string, len(filter.AccountIDs))
		for i, id := range filter.AccountIDs {
			placeholders[i] = fmt.Sprintf("$%d", argID)
			args = append(args, id)
			argID++
		}
		query += ` AND account_id IN (` + strings.Join(placeholders, ", ") + `)`
	}

	query += ` ORDER BY completed_at DESC, transaction_id DESC`

	err := q.SelectContext(ctx, &transactions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered transactions: %w", err)
	}
	return transactions, nil
}

func (tr *TransRepository) GetTransactionsByIDs(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID) ([]domain.Transaction, error) {
	if len(transactionIDs) == 0 {
		return nil, nil
	}
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return nil, err
	}

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

func (tr *TransRepository) ResolveAutoCategoryID(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string) (*uuid.UUID, error) {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return nil, err
	}

	merchantKey := normalizeRuleKey(description)
	if mccCode != nil {
		var categoryID uuid.UUID
		err := q.GetContext(
			ctx,
			&categoryID,
			`SELECT category_id
			 FROM AutoCategoryRules
			 WHERE user_id = $1 AND is_income = $2 AND mcc_code = $3
			 ORDER BY updated_at DESC
			 LIMIT 1`,
			userID,
			isIncome,
			strings.TrimSpace(*mccCode),
		)
		if err == nil {
			return &categoryID, nil
		}
	}

	if merchantKey == "" {
		return nil, nil
	}

	var categoryID uuid.UUID
	err := q.GetContext(
		ctx,
		&categoryID,
		`SELECT category_id
		 FROM AutoCategoryRules
		 WHERE user_id = $1 AND is_income = $2 AND merchant_key = $3
		 ORDER BY updated_at DESC
		 LIMIT 1`,
		userID,
		isIncome,
		merchantKey,
	)
	if err != nil {
		return nil, nil
	}
	return &categoryID, nil
}

func (tr *TransRepository) UpsertAutoCategoryRule(ctx context.Context, userID uuid.UUID, isIncome bool, mccCode *string, description string, categoryID uuid.UUID) error {
	q := database.GetQueryer(ctx, tr.db)
	if err := tr.ensureTransactionsSchema(ctx, q); err != nil {
		return err
	}
	merchantKey := normalizeRuleKey(description)
	normalizedMCC := ""
	if mccCode != nil {
		normalizedMCC = strings.TrimSpace(*mccCode)
	}

	if normalizedMCC == "" && merchantKey == "" {
		return nil
	}

	if normalizedMCC != "" {
		if _, err := q.ExecContext(
			ctx,
			`INSERT INTO AutoCategoryRules (user_id, is_income, mcc_code, merchant_key, category_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 ON CONFLICT (user_id, is_income, mcc_code)
			 DO UPDATE SET category_id = EXCLUDED.category_id, merchant_key = EXCLUDED.merchant_key, updated_at = CURRENT_TIMESTAMP`,
			userID,
			isIncome,
			normalizedMCC,
			merchantKey,
			categoryID,
		); err != nil {
			return fmt.Errorf("failed to upsert mcc auto-category rule: %w", err)
		}
	}

	if merchantKey != "" {
		if _, err := q.ExecContext(
			ctx,
			`INSERT INTO AutoCategoryRules (user_id, is_income, mcc_code, merchant_key, category_id, created_at, updated_at)
			 VALUES ($1, $2, NULL, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 ON CONFLICT (user_id, is_income, merchant_key)
			 DO UPDATE SET category_id = EXCLUDED.category_id, updated_at = CURRENT_TIMESTAMP`,
			userID,
			isIncome,
			merchantKey,
			categoryID,
		); err != nil {
			return fmt.Errorf("failed to upsert merchant auto-category rule: %w", err)
		}
	}

	return nil
}

func (tr *TransRepository) ensureTransactionsSchema(ctx context.Context, q database.Queryer) error {
	queries := []string{
		`ALTER TABLE Transactions ADD COLUMN IF NOT EXISTS sender_account TEXT`,
		`ALTER TABLE Transactions ADD COLUMN IF NOT EXISTS receiver_account TEXT`,
		`ALTER TABLE Transactions ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'RUB'`,
		`ALTER TABLE Transactions ADD COLUMN IF NOT EXISTS bank_fee BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE Transactions ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'completed'`,
		`ALTER TABLE Transactions ADD COLUMN IF NOT EXISTS external_transaction_id TEXT`,
		`ALTER TABLE Transactions ADD COLUMN IF NOT EXISTS mcc_code VARCHAR(4)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_external_uid ON Transactions(user_id, account_id, external_transaction_id) WHERE external_transaction_id IS NOT NULL`,
		`CREATE TABLE IF NOT EXISTS AutoCategoryRules (
			rule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			is_income BOOLEAN NOT NULL,
			mcc_code VARCHAR(4),
			merchant_key TEXT NOT NULL DEFAULT '',
			category_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_user_auto_category_rule FOREIGN KEY (user_id) REFERENCES Users(user_id) ON DELETE CASCADE,
			CONSTRAINT fk_category_auto_category_rule FOREIGN KEY (category_id) REFERENCES Category(category_id) ON DELETE CASCADE
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_auto_category_rule_mcc ON AutoCategoryRules (user_id, is_income, mcc_code) WHERE mcc_code IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_auto_category_rule_merchant ON AutoCategoryRules (user_id, is_income, merchant_key) WHERE merchant_key <> ''`,
	}
	for _, query := range queries {
		if _, err := q.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to ensure transactions schema: %w", err)
		}
	}
	return nil
}

func normalizeRuleKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '.' || r == '*' {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
