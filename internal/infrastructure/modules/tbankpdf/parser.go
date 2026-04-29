package tbankpdf

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

var (
	ErrInvalidPDF            = errors.New("invalid pdf file")
	ErrStatementNotSupported = errors.New("unsupported statement format")
)

type Statement struct {
	ContractNumber string
	AccountNumber  string
	Balance        int64
	Transactions   []TransactionEntry
}

type TransactionEntry struct {
	CompletedAt     time.Time
	Amount          int64
	IsIncome        bool
	Description     string
	CardNumber      string
	BankFee         int64
	SenderAccount   *string
	ReceiverAccount *string
	MCCCode         *string
	ExternalID      *string
}

var (
	reSpaces         = regexp.MustCompile(`\s+`)
	reContractNumber = regexp.MustCompile(`Номер\s+договора:\s*([0-9]+)`)
	reAccountNumber  = regexp.MustCompile(`Номер\s+лицевого\s+счета:\s*([0-9]+)`)
	reBalance        = regexp.MustCompile(`Сумма\s+доступного\s+остатка\s+на\s+[0-9.]+:\s*([+\-]?[0-9\s]+[.,][0-9]{2})\s*₽`)
	reDate           = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)
	reTime           = regexp.MustCompile(`^\d{2}:\d{2}$`)
	reCard           = regexp.MustCompile(`^(—|\d{4})$`)
	reMCC            = regexp.MustCompile(`(?i)\bmcc[:\s]*([0-9]{4})\b`)
)

func ParseStatement(data []byte) (*Statement, error) {
	if len(data) == 0 {
		return nil, ErrInvalidPDF
	}

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, ErrInvalidPDF
	}

	plainText, err := reader.GetPlainText()
	if err != nil {
		return nil, ErrInvalidPDF
	}

	raw, err := io.ReadAll(plainText)
	if err != nil {
		return nil, ErrInvalidPDF
	}

	text := strings.ReplaceAll(string(raw), "\u00a0", " ")
	compact := strings.TrimSpace(reSpaces.ReplaceAllString(text, " "))
	if compact == "" {
		return nil, ErrStatementNotSupported
	}

	statement := &Statement{}

	if m := reContractNumber.FindStringSubmatch(compact); len(m) > 1 {
		statement.ContractNumber = strings.TrimSpace(m[1])
	}
	if m := reAccountNumber.FindStringSubmatch(compact); len(m) > 1 {
		statement.AccountNumber = strings.TrimSpace(m[1])
	}

	balanceMatch := reBalance.FindStringSubmatch(compact)
	if len(balanceMatch) < 2 {
		return nil, ErrStatementNotSupported
	}
	balance, err := parseMoneyToMinor(balanceMatch[1])
	if err != nil {
		return nil, ErrStatementNotSupported
	}
	statement.Balance = balance

	txs, txErr := parseTransactions(text)
	if txErr != nil || len(txs) == 0 {
		return nil, ErrStatementNotSupported
	}

	statement.Transactions = txs
	return statement, nil
}

func parseMoneyToMinor(input string) (int64, error) {
	s := strings.TrimSpace(input)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "₽", "")
	s = strings.ReplaceAll(s, ",", ".")
	if s == "" {
		return 0, errors.New("empty amount")
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(value * 100)), nil
}

func parseTransactions(raw string) ([]TransactionEntry, error) {
	lines := splitCleanLines(raw)
	loc := time.FixedZone("MSK", 3*60*60)
	transactions := make([]TransactionEntry, 0)

	for i := 0; i < len(lines); i++ {
		if i+5 >= len(lines) {
			break
		}

		if !reDate.MatchString(lines[i]) || !reTime.MatchString(lines[i+1]) {
			continue
		}
		if !reDate.MatchString(lines[i+2]) || !reTime.MatchString(lines[i+3]) {
			continue
		}

		amount, err := parseMoneyToMinor(lines[i+4])
		if err != nil {
			continue
		}
		if _, err := parseMoneyToMinor(lines[i+5]); err != nil {
			continue
		}

		j := i + 6
		descParts := make([]string, 0, 2)
		card := ""

		for ; j < len(lines); j++ {
			if reCard.MatchString(lines[j]) {
				card = lines[j]
				j++
				break
			}
			if reDate.MatchString(lines[j]) && len(descParts) > 0 {
				break
			}
			descParts = append(descParts, lines[j])
		}

		description := strings.TrimSpace(strings.Join(descParts, " "))
		if description == "" {
			i = j - 1
			continue
		}
		mccCode := extractMCC(description)

		dt, err := time.ParseInLocation("02.01.2006 15:04", lines[i]+" "+lines[i+1], loc)
		if err != nil {
			i = j - 1
			continue
		}

		isIncome := amount > 0
		absAmount := amount
		if absAmount < 0 {
			absAmount = -absAmount
		}

		externalID := buildExternalID(dt.UTC(), absAmount, isIncome, description, card)
		transactions = append(transactions, TransactionEntry{
			CompletedAt: dt.UTC(),
			Amount:      absAmount,
			IsIncome:    isIncome,
			Description: description,
			CardNumber:  card,
			BankFee:     0,
			MCCCode:     mccCode,
			ExternalID:  &externalID,
		})

		if j <= i {
			i += 6
		} else {
			i = j - 1
		}
	}

	if len(transactions) == 0 {
		return nil, ErrStatementNotSupported
	}
	return transactions, nil
}

func extractMCC(description string) *string {
	match := reMCC.FindStringSubmatch(description)
	if len(match) < 2 {
		return nil
	}
	mcc := strings.TrimSpace(match[1])
	if mcc == "" {
		return nil
	}
	return &mcc
}

func buildExternalID(ts time.Time, amount int64, isIncome bool, description string, card string) string {
	sign := "expense"
	if isIncome {
		sign = "income"
	}
	raw := ts.Format(time.RFC3339Nano) + "|" + strconv.FormatInt(amount, 10) + "|" + sign + "|" + strings.TrimSpace(description) + "|" + strings.TrimSpace(card)
	sum := sha1.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func splitCleanLines(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	raw = strings.ReplaceAll(raw, "\u00a0", " ")
	parts := strings.Split(raw, "\n")
	lines := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(reSpaces.ReplaceAllString(p, " "))
		if s == "" {
			continue
		}
		lines = append(lines, s)
	}
	return lines
}
