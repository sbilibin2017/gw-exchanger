package models

import (
	"time"

	"github.com/google/uuid"
)

// Supported currencies
const (
	USD = "USD"
	RUB = "RUB"
	EUR = "EUR"
)

// ExchangeRate описывает модель записи курса обмена валют,
// которая хранится в базе данных.
type ExchangeRate struct {
	ExchangeRateID uuid.UUID `json:"exchange_rate_id" db:"exchange_rate_id"` // Уникальный идентификатор курса (UUID)
	FromCurrency   string    `json:"from_currency" db:"from_currency"`       // Исходная валюта
	ToCurrency     string    `json:"to_currency" db:"to_currency"`           // Целевая валюта
	Rate           float64   `json:"rate" db:"rate"`                         // Курс обмена (DECIMAL(18,6))
	CreatedAt      time.Time `json:"created_at" db:"created_at"`             // Дата и время создания записи
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`             // Дата и время последнего обновления записи
}
