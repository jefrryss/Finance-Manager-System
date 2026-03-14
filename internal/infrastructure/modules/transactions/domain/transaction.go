package transactions

import (
	"time"

	"github.com/google/uuid"
)

type Transactions struct {
	user_id          uuid.UUID
	transaction_id   int
	account_id       int
	category_id      uuid.UUID
	name_transaction string
	is_income        bool
	amount           int64
	completed_at     time.Time
	is_hidden        bool
	is_imported      bool
	comment          string
}
