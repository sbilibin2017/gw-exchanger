package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gw-exchanger/internal/models"
	"go.uber.org/zap"
)

// ExchangeRateReadRepository reads currency exchange rates from the DB.
type ExchangeRateReadRepository struct {
	db  *sqlx.DB
	log *zap.SugaredLogger
}

// NewExchangeRateReadRepository creates a new repository with a logger.
func NewExchangeRateReadRepository(log *zap.SugaredLogger, db *sqlx.DB) *ExchangeRateReadRepository {
	return &ExchangeRateReadRepository{
		db:  db,
		log: log,
	}
}

// Get returns the exchange rate for a currency pair.
func (r *ExchangeRateReadRepository) Get(
	ctx context.Context,
	fromCurrency string,
	toCurrency string,
) (*float64, error) {

	query, args := buildGetExchangeRateQuery(fromCurrency, toCurrency)
	var rate float64
	err := r.db.GetContext(ctx, &rate, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.log.Errorf("op: get exchange rate, err: %v", err)
		return nil, err
	}

	return &rate, nil
}

// List returns all exchange rate records.
func (r *ExchangeRateReadRepository) List(
	ctx context.Context,
) ([]models.ExchangeRateDB, error) {

	query, args := buildListExchangeRateQuery()
	var rates []models.ExchangeRateDB
	err := r.db.SelectContext(ctx, &rates, query, args...)
	if err != nil {
		r.log.Errorf("op: list exchange rates, err: %v", err)
		return nil, err
	}

	return rates, nil
}

// buildGetExchangeRateQuery returns the SQL query and arguments for a single exchange rate.
func buildGetExchangeRateQuery(fromCurrency, toCurrency string) (string, []any) {
	query := `
		SELECT rate
		FROM exchange_rates
		WHERE from_currency = $1 AND to_currency = $2
	`
	args := []any{fromCurrency, toCurrency}
	return query, args
}

// buildListExchangeRateQuery returns the SQL query and empty arguments for all exchange rates.
func buildListExchangeRateQuery() (string, []any) {
	query := `
		SELECT exchange_rate_id, from_currency, to_currency, rate, created_at, updated_at
		FROM exchange_rates
		ORDER BY created_at DESC
	`
	return query, nil
}
