package repositories_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/sbilibin2017/gw-exchanger/internal/models"
	"github.com/sbilibin2017/gw-exchanger/internal/repositories"
)

// helper to create logger
func getLogger(t *testing.T) *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	return logger.Sugar()
}

// helper to create sqlx.DB with sqlmock
func getMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxdb := sqlx.NewDb(db, "sqlmock")
	return sqlxdb, mock, func() { sqlxdb.Close() }
}

func TestExchangeRateReadRepository_Get_Success(t *testing.T) {
	db, mock, closeFn := getMockDB(t)
	defer closeFn()
	logger := getLogger(t)

	repo := repositories.NewExchangeRateReadRepository(logger, db)

	from := "USD"
	to := "EUR"
	rate := 1.23

	mock.ExpectQuery(`SELECT rate FROM exchange_rates WHERE from_currency = \$1 AND to_currency = \$2`).
		WithArgs(from, to).
		WillReturnRows(sqlmock.NewRows([]string{"rate"}).AddRow(rate))

	ctx := context.Background()
	got, err := repo.Get(ctx, from, to)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, rate, *got)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExchangeRateReadRepository_Get_NotFound(t *testing.T) {
	db, mock, closeFn := getMockDB(t)
	defer closeFn()
	logger := getLogger(t)

	repo := repositories.NewExchangeRateReadRepository(logger, db)

	from := "USD"
	to := "EUR"

	mock.ExpectQuery(`SELECT rate FROM exchange_rates WHERE from_currency = \$1 AND to_currency = \$2`).
		WithArgs(from, to).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	got, err := repo.Get(ctx, from, to)
	require.NoError(t, err)
	assert.Nil(t, got)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExchangeRateReadRepository_Get_Error(t *testing.T) {
	db, mock, closeFn := getMockDB(t)
	defer closeFn()
	logger := getLogger(t)

	repo := repositories.NewExchangeRateReadRepository(logger, db)

	from := "USD"
	to := "EUR"

	mock.ExpectQuery(`SELECT rate FROM exchange_rates WHERE from_currency = \$1 AND to_currency = \$2`).
		WithArgs(from, to).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	got, err := repo.Get(ctx, from, to)
	assert.Error(t, err)
	assert.Nil(t, got)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExchangeRateReadRepository_List_Success(t *testing.T) {
	db, mock, closeFn := getMockDB(t)
	defer closeFn()
	logger := getLogger(t)

	repo := repositories.NewExchangeRateReadRepository(logger, db)

	rates := []models.ExchangeRate{
		{ExchangeRateID: uuid.New(), FromCurrency: "USD", ToCurrency: "EUR", Rate: 1.23, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ExchangeRateID: uuid.New(), FromCurrency: "EUR", ToCurrency: "USD", Rate: 0.81, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	rows := sqlmock.NewRows([]string{"exchange_rate_id", "from_currency", "to_currency", "rate", "created_at", "updated_at"})
	for _, r := range rates {
		rows.AddRow(r.ExchangeRateID.String(), r.FromCurrency, r.ToCurrency, r.Rate, r.CreatedAt, r.UpdatedAt)
	}

	mock.ExpectQuery(`SELECT exchange_rate_id, from_currency, to_currency, rate, created_at, updated_at FROM exchange_rates ORDER BY created_at DESC`).
		WillReturnRows(rows)

	ctx := context.Background()
	got, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, got, len(rates))
	assert.Equal(t, rates[0].ExchangeRateID, got[0].ExchangeRateID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExchangeRateReadRepository_List_Error(t *testing.T) {
	db, mock, closeFn := getMockDB(t)
	defer closeFn()
	logger := getLogger(t)

	repo := repositories.NewExchangeRateReadRepository(logger, db)

	mock.ExpectQuery(`SELECT exchange_rate_id, from_currency, to_currency, rate, created_at, updated_at FROM exchange_rates ORDER BY created_at DESC`).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	got, err := repo.List(ctx)
	assert.Error(t, err)
	assert.Nil(t, got)

	assert.NoError(t, mock.ExpectationsWereMet())
}
