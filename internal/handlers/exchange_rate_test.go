package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/sbilibin2017/gw-exchanger/internal/models"
	pb "github.com/sbilibin2017/proto-exchange/exchange"
)

// helper logger
func getLogger(t *testing.T) *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	return logger.Sugar()
}

func float32Ptr(f float32) *float32 { return &f }

// -------------------------
// Test GetExchangeRateForCurrency
// -------------------------
func TestExchangeRateHandler_GetExchangeRateForCurrency(t *testing.T) {
	ctx := context.Background()
	logger := getLogger(t)

	tests := []struct {
		name      string
		req       *pb.CurrencyRequest
		wantResp  *pb.ExchangeRateResponse
		wantErr   bool
		mockSetup func(ctrl *gomock.Controller, handler *ExchangeRateHandler)
	}{
		{
			name:     "success",
			req:      &pb.CurrencyRequest{FromCurrency: models.USD, ToCurrency: models.EUR},
			wantResp: &pb.ExchangeRateResponse{FromCurrency: models.USD, ToCurrency: models.EUR, Rate: 1.23},
			mockSetup: func(ctrl *gomock.Controller, handler *ExchangeRateHandler) {
				handlerSvc := handler.svc.(*MockExchangeRateService)
				handlerSvc.EXPECT().
					GetExchangeRate(ctx, models.USD, models.EUR).
					Return(float32Ptr(1.23), nil)
			},
		},
		{
			name: "rate not found",
			req:  &pb.CurrencyRequest{FromCurrency: models.USD, ToCurrency: models.EUR},
			mockSetup: func(ctrl *gomock.Controller, handler *ExchangeRateHandler) {
				handlerSvc := handler.svc.(*MockExchangeRateService)
				handlerSvc.EXPECT().
					GetExchangeRate(ctx, models.USD, models.EUR).
					Return(nil, nil)
			},
		},
		{
			name:    "service error",
			req:     &pb.CurrencyRequest{FromCurrency: models.USD, ToCurrency: models.EUR},
			wantErr: true,
			mockSetup: func(ctrl *gomock.Controller, handler *ExchangeRateHandler) {
				handlerSvc := handler.svc.(*MockExchangeRateService)
				handlerSvc.EXPECT().
					GetExchangeRate(ctx, models.USD, models.EUR).
					Return(nil, errors.New("service error"))
			},
		},
		{
			name:    "invalid request same currency",
			req:     &pb.CurrencyRequest{FromCurrency: models.USD, ToCurrency: models.USD},
			wantErr: true,
		},
		{
			name:    "invalid request unsupported currency",
			req:     &pb.CurrencyRequest{FromCurrency: "XXX", ToCurrency: models.EUR},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := NewMockExchangeRateService(ctrl)
			handler := NewExchangeRateHandler(logger, mockSvc)

			if tt.mockSetup != nil {
				tt.mockSetup(ctrl, handler)
			}

			got, err := handler.GetExchangeRateForCurrency(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tt.wantResp != nil {
					assert.Equal(t, tt.wantResp, got)
				}
			}
		})
	}
}

// -------------------------
// Test GetExchangeRates
// -------------------------
func TestExchangeRateHandler_GetExchangeRates(t *testing.T) {
	ctx := context.Background()
	logger := getLogger(t)

	tests := []struct {
		name      string
		wantResp  *pb.ExchangeRatesResponse
		wantErr   bool
		mockSetup func(ctrl *gomock.Controller, handler *ExchangeRateHandler)
	}{
		{
			name: "success",
			wantResp: &pb.ExchangeRatesResponse{
				Rates: map[string]float32{
					models.EUR: 1.23,
					models.USD: 0.81,
				},
			},
			mockSetup: func(ctrl *gomock.Controller, handler *ExchangeRateHandler) {
				handlerSvc := handler.svc.(*MockExchangeRateService)
				handlerSvc.EXPECT().GetAllExchangeRates(ctx).Return([]models.ExchangeRate{
					{FromCurrency: models.USD, ToCurrency: models.EUR, Rate: 1.23},
					{FromCurrency: models.EUR, ToCurrency: models.USD, Rate: 0.81},
				}, nil)
			},
		},
		{
			name:    "service error",
			wantErr: true,
			mockSetup: func(ctrl *gomock.Controller, handler *ExchangeRateHandler) {
				handlerSvc := handler.svc.(*MockExchangeRateService)
				handlerSvc.EXPECT().GetAllExchangeRates(ctx).Return(nil, errors.New("service error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := NewMockExchangeRateService(ctrl)
			handler := NewExchangeRateHandler(logger, mockSvc)

			if tt.mockSetup != nil {
				tt.mockSetup(ctrl, handler)
			}

			got, err := handler.GetExchangeRates(ctx, &pb.Empty{})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, got)
			}
		})
	}
}

// -------------------------
// Test private functions
// -------------------------
func TestValidateCurrencyRequest(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		{"valid USD->EUR", models.USD, models.EUR, false},
		{"same currency", models.USD, models.USD, true},
		{"unsupported from", "XXX", models.EUR, true},
		{"unsupported to", models.USD, "YYY", true},
		{"empty from", "", models.EUR, true},
		{"empty to", models.USD, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCurrencyRequest(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsSupportedCurrency(t *testing.T) {
	tests := []struct {
		currency  string
		supported bool
	}{
		{models.USD, true},
		{models.RUB, true},
		{models.EUR, true},
		{"XXX", false},
		{"YYY", false},
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			assert.Equal(t, tt.supported, isSupportedCurrency(tt.currency))
		})
	}
}
