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

func resolveCategoryID(categories []categoryDomain.Category, description string, isIncome bool, mccCode *string) *uuid.UUID {
	categoryName := classifyCategoryName(description, isIncome, mccCode)
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

func classifyCategoryName(description string, isIncome bool, mccCode *string) string {
	d := normalizeKey(description)
	if mccCode != nil {
		if mapped := mapMCCToCategory(*mccCode, isIncome); mapped != "" {
			return mapped
		}
	}

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

func mapMCCToCategory(code string, isIncome bool) string {
	if isIncome {
		return ""
	}
	switch strings.TrimSpace(code) {
	case "5411", "5422", "5441", "5451", "5462", "5499":
		return "Продукты"
	case "5811", "5812", "5813", "5814":
		return "Кафе и рестораны"
	case "4111", "4121", "4131", "4789":
		return "Транспорт"
	case "5912", "8011", "8021", "8041", "8062", "8099":
		return "Здоровье"
	case "4814", "4899", "5732", "5815", "5968":
		return "Подписки"
	case "5311", "5331", "5399", "5651", "5691", "5712", "5734", "5942":
		return "Покупки"
	case "7995", "7832", "7922", "7997":
		return "Развлечения"
	default:
		return ""
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
