package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gw-exchanger/internal/models"
	pb "github.com/sbilibin2017/proto-exchange/exchange"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// floatPtr helper
func floatPtr(f float64) *float64 {
	return &f
}

func TestGetExchangeRateForCurrency(t *testing.T) {
	testCases := []struct {
		name          string
		fromCurrency  string
		toCurrency    string
		mockSetup     func(t *testing.T) (*ExchangeRateService, *gomock.Controller)
		expectError   bool
		expectNilResp bool
		expectedRate  float32
	}{
		{
			name:         "valid rate found",
			fromCurrency: "USD",
			toCurrency:   "RUB",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				mockReader := NewMockExchangeRateReader(ctrl)
				mockReader.EXPECT().
					Get(gomock.Any(), "USD", "RUB").
					Return(floatPtr(75.5), nil)
				svc := NewExchangeRateService(zap.NewNop().Sugar(), mockReader)
				return svc, ctrl
			},
			expectError:   false,
			expectNilResp: false,
			expectedRate:  75.5,
		},
		{
			name:         "rate not found",
			fromCurrency: "USD",
			toCurrency:   "EUR",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				mockReader := NewMockExchangeRateReader(ctrl)
				mockReader.EXPECT().
					Get(gomock.Any(), "USD", "EUR").
					Return(nil, nil)
				svc := NewExchangeRateService(zap.NewNop().Sugar(), mockReader)
				return svc, ctrl
			},
			expectError:   false,
			expectNilResp: true,
		},
		{
			name:         "reader returns error",
			fromCurrency: "USD",
			toCurrency:   "RUB",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				mockReader := NewMockExchangeRateReader(ctrl)
				mockReader.EXPECT().
					Get(gomock.Any(), "USD", "RUB").
					Return(nil, errors.New("db error"))
				svc := NewExchangeRateService(zap.NewNop().Sugar(), mockReader)
				return svc, ctrl
			},
			expectError:   true,
			expectNilResp: true,
		},
		{
			name:         "unsupported from currency",
			fromCurrency: "GBP",
			toCurrency:   "USD",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				svc := NewExchangeRateService(zap.NewNop().Sugar(), nil)
				return svc, nil
			},
			expectError:   true,
			expectNilResp: true,
		},
		{
			name:         "unsupported to currency",
			fromCurrency: "USD",
			toCurrency:   "JPY",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				svc := NewExchangeRateService(zap.NewNop().Sugar(), nil)
				return svc, nil
			},
			expectError:   true,
			expectNilResp: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctrl := tc.mockSetup(t)
			if ctrl != nil {
				defer ctrl.Finish()
			}

			resp, err := svc.GetExchangeRateForCurrency(context.Background(), &pb.CurrencyRequest{
				FromCurrency: tc.fromCurrency,
				ToCurrency:   tc.toCurrency,
			})

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expectNilResp {
				assert.Nil(t, resp)
			} else {
				assert.NotNil(t, resp)
				assert.Equal(t, tc.expectedRate, resp.Rate)
			}
		})
	}
}

func TestGetExchangeRates(t *testing.T) {
	testCases := []struct {
		name          string
		mockSetup     func(t *testing.T) (*ExchangeRateService, *gomock.Controller)
		expectError   bool
		expectedRates map[string]float32
	}{
		{
			name: "returns multiple rates",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				mockReader := NewMockExchangeRateReader(ctrl)
				mockReader.EXPECT().
					List(gomock.Any()).
					Return([]models.ExchangeRateDB{
						{ToCurrency: "RUB", Rate: 75.5},
						{ToCurrency: "EUR", Rate: 0.92},
					}, nil)
				svc := NewExchangeRateService(zap.NewNop().Sugar(), mockReader)
				return svc, ctrl
			},
			expectError: false,
			expectedRates: map[string]float32{
				"RUB": 75.5,
				"EUR": 0.92,
			},
		},
		{
			name: "no rates found",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				mockReader := NewMockExchangeRateReader(ctrl)
				mockReader.EXPECT().
					List(gomock.Any()).
					Return([]models.ExchangeRateDB{}, nil)
				svc := NewExchangeRateService(zap.NewNop().Sugar(), mockReader)
				return svc, ctrl
			},
			expectError:   false,
			expectedRates: map[string]float32{},
		},
		{
			name: "reader returns error",
			mockSetup: func(t *testing.T) (*ExchangeRateService, *gomock.Controller) {
				ctrl := gomock.NewController(t)
				mockReader := NewMockExchangeRateReader(ctrl)
				mockReader.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("db error"))
				svc := NewExchangeRateService(zap.NewNop().Sugar(), mockReader)
				return svc, ctrl
			},
			expectError:   true,
			expectedRates: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctrl := tc.mockSetup(t)
			defer func() {
				if ctrl != nil {
					ctrl.Finish()
				}
			}()

			resp, err := svc.GetExchangeRates(context.Background(), &pb.Empty{})

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tc.expectedRates, resp.Rates)
			}
		})
	}
}
