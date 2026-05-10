package tbankpdf

import (
	"bytes"
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
	CompletedAt time.Time
	Amount      int64
	IsIncome    bool
	Description string
	CardNumber  string
}

var (
	reSpaces         = regexp.MustCompile(`\s+`)
	reContractNumber = regexp.MustCompile(`Номер\s+договора:\s*([0-9]+)`)
	reAccountNumber  = regexp.MustCompile(`Номер\s+лицевого\s+счета:\s*([0-9]+)`)
	reBalance        = regexp.MustCompile(`Сумма\s+доступного\s+остатка\s+на\s+[0-9.]+:\s*([+\-]?[0-9\s]+[.,][0-9]{2})\s*₽`)
	reTx             = regexp.MustCompile(`(\d{2}\.\d{2}\.\d{4})\s+(\d{2}:\d{2})\s+\d{2}\.\d{2}\.\d{4}\s+\d{2}:\d{2}\s+([+\-][0-9\s]+[.,][0-9]{2})\s*₽\s+[+\-][0-9\s]+[.,][0-9]{2}\s*₽\s+(.+?)\s+(—|\d{4})(?=\s+\d{2}\.\d{2}\.\d{4}\s+\d{2}:\d{2}\s+\d{2}\.\d{2}\.\d{4}\s+\d{2}:\d{2}\s+[+\-]|\s+АО\s+«ТБанк|\s+Пополнения:|$)`)
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

	matches := reTx.FindAllStringSubmatch(compact, -1)
	if len(matches) == 0 {
		return nil, ErrStatementNotSupported
	}

	loc := time.FixedZone("MSK", 3*60*60)
	txs := make([]TransactionEntry, 0, len(matches))

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		dt, err := time.ParseInLocation("02.01.2006 15:04", strings.TrimSpace(match[1])+" "+strings.TrimSpace(match[2]), loc)
		if err != nil {
			continue
		}

		amount, err := parseMoneyToMinor(match[3])
		if err != nil {
			continue
		}

		isIncome := amount > 0
		absAmount := amount
		if absAmount < 0 {
			absAmount = -absAmount
		}

		description := strings.TrimSpace(match[4])
		card := strings.TrimSpace(match[5])

		txs = append(txs, TransactionEntry{
			CompletedAt: dt.UTC(),
			Amount:      absAmount,
			IsIncome:    isIncome,
			Description: description,
			CardNumber:  card,
		})
	}

	if len(txs) == 0 {
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
