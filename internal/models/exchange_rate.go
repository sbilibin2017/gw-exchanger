package models

import (
	"time"

	"github.com/google/uuid"
)

// ExchangeRateDB describes the model of a currency exchange rate record
// stored in the database.
type ExchangeRateDB struct {
	ExchangeRateID uuid.UUID `json:"exchange_rate_id" db:"exchange_rate_id"` // Unique identifier of the exchange rate (UUID)
	FromCurrency   string    `json:"from_currency" db:"from_currency"`       // Source currency
	ToCurrency     string    `json:"to_currency" db:"to_currency"`           // Target currency
	Rate           float64   `json:"rate" db:"rate"`                         // Exchange rate value (DECIMAL(18,6))
	CreatedAt      time.Time `json:"created_at" db:"created_at"`             // Record creation date and time
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`             // Record last update date and time
}
