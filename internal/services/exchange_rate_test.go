package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/sbilibin2017/gw-exchanger/internal/models"
	"github.com/sbilibin2017/gw-exchanger/internal/services"
)

func getLogger(t *testing.T) *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	return logger.Sugar()
}

func TestExchangeRateService_GetExchangeRate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := services.NewMockExchangeRateReader(ctrl)
	logger := getLogger(t)
	service := services.NewExchangeRateService(logger, mockReader)
	ctx := context.Background()

	tests := []struct {
		name       string
		from       string
		to         string
		mockReturn *float64
		mockError  error
		wantRate   *float32
		wantErr    bool
	}{
		{
			name:       "rate found",
			from:       "USD",
			to:         "EUR",
			mockReturn: float64Ptr(1.23),
			wantRate:   float32Ptr(1.23),
		},
		{
			name:       "rate not found",
			from:       "USD",
			to:         "JPY",
			mockReturn: nil,
			wantRate:   nil,
		},
		{
			name:      "error from reader",
			from:      "USD",
			to:        "GBP",
			mockError: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader.EXPECT().
				Get(ctx, tt.from, tt.to).
				Return(tt.mockReturn, tt.mockError)

			got, err := service.GetExchangeRate(ctx, tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRate, got)
			}
		})
	}
}

func TestExchangeRateService_GetAllExchangeRates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := services.NewMockExchangeRateReader(ctrl)
	logger := getLogger(t)
	service := services.NewExchangeRateService(logger, mockReader)
	ctx := context.Background()

	exampleRates := []models.ExchangeRate{
		{FromCurrency: "USD", ToCurrency: "EUR", Rate: 1.23},
		{FromCurrency: "EUR", ToCurrency: "USD", Rate: 0.81},
	}

	tests := []struct {
		name       string
		mockReturn []models.ExchangeRate
		mockError  error
		want       []models.ExchangeRate
		wantErr    bool
	}{
		{
			name:       "success",
			mockReturn: exampleRates,
			want:       exampleRates,
		},
		{
			name:      "error from reader",
			mockError: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader.EXPECT().
				List(ctx).
				Return(tt.mockReturn, tt.mockError)

			got, err := service.GetAllExchangeRates(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// helper functions
func float64Ptr(f float64) *float64 { return &f }
func float32Ptr(f float64) *float32 { v := float32(f); return &v }
