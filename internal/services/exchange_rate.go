package services

import (
	"context"
	"fmt"

	"github.com/sbilibin2017/gw-exchanger/internal/models"
	pb "github.com/sbilibin2017/proto-exchange/exchange"
	"go.uber.org/zap"
)

// Constants for supported currencies
const (
	usd = "USD"
	rub = "RUB"
	eur = "EUR"
)

// supportedCurrencies is a map of valid currencies for quick lookup
var supportedCurrencies = map[string]struct{}{
	usd: {},
	rub: {},
	eur: {},
}

// ExchangeRateReader is an interface for reading currency exchange rates.
type ExchangeRateReader interface {
	Get(ctx context.Context, fromCurrency, toCurrency string) (*float64, error)
	List(ctx context.Context) ([]models.ExchangeRateDB, error)
}

// ExchangeRateService implements the gRPC server for currency exchange rates.
type ExchangeRateService struct {
	pb.UnimplementedExchangeServiceServer
	reader ExchangeRateReader
	log    *zap.SugaredLogger
}

// NewExchangeRateService creates a new instance of ExchangeRateService.
func NewExchangeRateService(
	log *zap.SugaredLogger,
	reader ExchangeRateReader,
) *ExchangeRateService {
	return &ExchangeRateService{
		reader: reader,
		log:    log,
	}
}

// GetExchangeRateForCurrency returns the exchange rate for a specific currency pair.
func (s *ExchangeRateService) GetExchangeRateForCurrency(
	ctx context.Context,
	req *pb.CurrencyRequest,
) (*pb.ExchangeRateResponse, error) {

	if _, ok := supportedCurrencies[req.FromCurrency]; !ok {
		err := fmt.Errorf("unsupported from currency: %s", req.FromCurrency)
		s.log.Errorf("op: get exchange rate, err: %v", err)
		return nil, err
	}
	if _, ok := supportedCurrencies[req.ToCurrency]; !ok {
		err := fmt.Errorf("unsupported to currency: %s", req.ToCurrency)
		s.log.Errorf("op: get exchange rate, err: %v", err)
		return nil, err
	}

	ratePtr, err := s.reader.Get(ctx, req.FromCurrency, req.ToCurrency)
	if err != nil {
		s.log.Errorf("op: get exchange rate, err: %v", err)
		return nil, err
	}

	if ratePtr == nil {
		s.log.Warnf("op: get exchange rate, rate not found: %s -> %s", req.FromCurrency, req.ToCurrency)
		return nil, nil
	}

	return &pb.ExchangeRateResponse{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         float32(*ratePtr),
	}, nil
}

// GetExchangeRates returns all available exchange rates.
func (s *ExchangeRateService) GetExchangeRates(
	ctx context.Context,
	req *pb.Empty,
) (*pb.ExchangeRatesResponse, error) {

	rows, err := s.reader.List(ctx)
	if err != nil {
		s.log.Errorf("op: list exchange rates, err: %v", err)
		return nil, err
	}

	rates := make(map[string]float32, len(rows))
	for _, r := range rows {
		rates[r.ToCurrency] = float32(r.Rate)
	}

	return &pb.ExchangeRatesResponse{
		Rates: rates,
	}, nil
}
