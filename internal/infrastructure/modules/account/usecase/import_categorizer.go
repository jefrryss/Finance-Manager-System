package usecase

import (
	"strings"
	"unicode"

	"github.com/google/uuid"

	categoryDomain "Finance-Manager-System/internal/infrastructure/modules/category/domain"
	"Finance-Manager-System/internal/infrastructure/modules/tbankpdf"
)

func buildImportedAccountName(statement *tbankpdf.Statement) string {
	base := "Т-Банк счет"
	if statement.AccountNumber == "" {
		return base
	}
	if len(statement.AccountNumber) <= 4 {
		return base + " " + statement.AccountNumber
	}
	return base + " *" + statement.AccountNumber[len(statement.AccountNumber)-4:]
}

func resolveCategoryID(categories []categoryDomain.Category, description string, isIncome bool) *uuid.UUID {
	categoryName := classifyCategoryName(description, isIncome)
	normalizedTarget := normalizeKey(categoryName)
	normalizedFallback := normalizeKey("Другое")

	var targetID *uuid.UUID
	var fallbackID *uuid.UUID
	var anyTypeID *uuid.UUID

	for i := range categories {
		c := categories[i]
		if c.IsIncome != isIncome {
			continue
		}

		if anyTypeID == nil {
			id := c.CategoryID
			anyTypeID = &id
		}

		n := normalizeKey(c.NameCategory)
		if n == normalizedTarget && targetID == nil {
			id := c.CategoryID
			targetID = &id
		}
		if n == normalizedFallback && fallbackID == nil {
			id := c.CategoryID
			fallbackID = &id
		}
	}

	if targetID != nil {
		return targetID
	}
	if fallbackID != nil {
		return fallbackID
	}
	return anyTypeID
}

func classifyCategoryName(description string, isIncome bool) string {
	d := normalizeKey(description)

	if isIncome {
		switch {
		case hasAny(d, "кэшбэк", "cashback"):
			return "Кэшбэк"
		case hasAny(d, "зарплат", "salary"):
			return "Зарплата"
		case hasAny(d, "процент", "interest"):
			return "Проценты"
		case hasAny(d, "подар"):
			return "Подарки"
		case hasAny(d, "перевод", "пополнение", "вывод средств"):
			return "Переводы"
		default:
			return "Другое"
		}
	}

	switch {
	case hasAny(d, "перевод", "сбп", "пополнение"):
		return "Переводы"
	case hasAny(d, "bundle", "yandex", "plus", "подписк"):
		return "Подписки"
	case hasAny(d, "transport", "mos.transport", "трансп", "билет"):
		return "Транспорт"
	case hasAny(d, "оптика", "apteka", "аптек", "clinic", "health"):
		return "Здоровье"
	case hasAny(d, "burger", "jpan", "stolovaya", "кафе", "ресторан", "naprilavke", "arkadiya", "qsr", "flowwow"):
		return "Кафе и рестораны"
	case hasAny(d, "perekrestok", "auchan", "winelab", "продукт", "перекресток"):
		return "Продукты"
	case hasAny(d, "ozon", "dns", "gold", "apple", "beeline", "shop", "barbershop", "оплата в"):
		return "Покупки"
	case hasAny(d, "ggs", "ggsel", "club", "klub", "onlipay", "fincom"):
		return "Развлечения"
	default:
		return "Другое"
	}
}

func hasAny(text string, variants ...string) bool {
	for _, v := range variants {
		if strings.Contains(text, normalizeKey(v)) {
			return true
		}
	}
	return false
}

func normalizeKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '*' || r == '.' || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
