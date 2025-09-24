package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gw-exchanger/internal/models"
	"go.uber.org/zap"
)

// ExchangeRateReadRepository отвечает за чтение курсов валют из БД.
type ExchangeRateReadRepository struct {
	db  *sqlx.DB
	log *zap.SugaredLogger
}

// NewExchangeRateReadRepository создаёт новый экземпляр репозитория с логгером.
func NewExchangeRateReadRepository(log *zap.SugaredLogger, db *sqlx.DB) *ExchangeRateReadRepository {
	return &ExchangeRateReadRepository{
		db:  db,
		log: log,
	}
}

// Get возвращает курс обмена между двумя валютами.
// Если запись не найдена, возвращается (nil, nil).
func (r *ExchangeRateReadRepository) Get(
	ctx context.Context,
	fromCurrency string,
	toCurrency string,
) (*float64, error) {
	r.log.Infof("Получение курса обмена: %s -> %s", fromCurrency, toCurrency)

	const query = `
		SELECT rate
		FROM exchange_rates
		WHERE from_currency = $1 AND to_currency = $2
	`

	var rate float64
	err := r.db.GetContext(ctx, &rate, query, fromCurrency, toCurrency)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warnf("Курс обмена не найден: %s -> %s", fromCurrency, toCurrency)
			return nil, nil
		}
		r.log.Errorf("Ошибка при получении курса обмена: %v", err)
		return nil, err
	}

	r.log.Infof("Курс обмена получен: %s -> %s = %f", fromCurrency, toCurrency, rate)
	return &rate, nil
}

// List возвращает все записи курсов обмена из таблицы exchange_rates.
func (r *ExchangeRateReadRepository) List(
	ctx context.Context,
) ([]models.ExchangeRate, error) {
	r.log.Info("Получение всех курсов обмена")

	const query = `
		SELECT exchange_rate_id, from_currency, to_currency, rate, created_at, updated_at
		FROM exchange_rates
		ORDER BY created_at DESC
	`

	var rates []models.ExchangeRate
	err := r.db.SelectContext(ctx, &rates, query)
	if err != nil {
		r.log.Errorf("Ошибка при получении всех курсов обмена: %v", err)
		return nil, err
	}

	r.log.Infof("Получено %d курсов обмена", len(rates))
	return rates, nil
}
