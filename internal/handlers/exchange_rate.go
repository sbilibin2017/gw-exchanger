package handlers

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/sbilibin2017/gw-exchanger/internal/models"
	pb "github.com/sbilibin2017/proto-exchange/exchange"
)

// ExchangeRateService интерфейс для работы с сервисом курсов валют (simple args)
type ExchangeRateService interface {
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*float32, error)
	GetAllExchangeRates(ctx context.Context) ([]models.ExchangeRate, error)
}

// ExchangeRateHandler реализует gRPC сервер для работы с курсами валют.
type ExchangeRateHandler struct {
	pb.UnimplementedExchangeServiceServer
	svc ExchangeRateService
	log *zap.SugaredLogger
}

// NewExchangeRateHandler создаёт новый экземпляр gRPC обработчика с логгером.
func NewExchangeRateHandler(log *zap.SugaredLogger, svc ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{
		svc: svc,
		log: log,
	}
}

// GetExchangeRateForCurrency возвращает курс обмена для конкретной валюты.
func (a *ExchangeRateHandler) GetExchangeRateForCurrency(
	ctx context.Context,
	req *pb.CurrencyRequest,
) (*pb.ExchangeRateResponse, error) {
	a.log.Infof("Получен запрос GetExchangeRateForCurrency: from=%s, to=%s", req.FromCurrency, req.ToCurrency)

	if err := validateCurrencyRequest(req.FromCurrency, req.ToCurrency); err != nil {
		a.log.Warnf("Ошибка валидации: %v", err)
		return nil, err
	}

	ratePtr, err := a.svc.GetExchangeRate(ctx, req.FromCurrency, req.ToCurrency)
	if err != nil {
		a.log.Errorf("Ошибка при получении курса обмена: %v", err)
		return nil, err
	}
	if ratePtr == nil {
		a.log.Warnf("Курс обмена не найден: %s -> %s", req.FromCurrency, req.ToCurrency)
		return nil, nil
	}

	rate := *ratePtr
	a.log.Infof("Возвращаем курс обмена: %s -> %s = %f", req.FromCurrency, req.ToCurrency, rate)
	return &pb.ExchangeRateResponse{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         rate,
	}, nil
}

// validateCurrencyRequest проверяет корректность запроса валют
func validateCurrencyRequest(fromCurrency, toCurrency string) error {
	if fromCurrency == "" {
		return fmt.Errorf("from_currency не может быть пустым")
	}
	if toCurrency == "" {
		return fmt.Errorf("to_currency не может быть пустым")
	}
	if fromCurrency == toCurrency {
		return fmt.Errorf("from_currency и to_currency не могут совпадать")
	}
	if !isSupportedCurrency(fromCurrency) {
		return fmt.Errorf("неподдерживаемая валюта from_currency: %s", fromCurrency)
	}
	if !isSupportedCurrency(toCurrency) {
		return fmt.Errorf("неподдерживаемая валюта to_currency: %s", toCurrency)
	}
	return nil
}

// isSupportedCurrency проверяет, поддерживается ли валюта
func isSupportedCurrency(currency string) bool {
	switch currency {
	case models.USD, models.RUB, models.EUR:
		return true
	default:
		return false
	}
}

// GetExchangeRates возвращает все доступные курсы валют.
func (a *ExchangeRateHandler) GetExchangeRates(
	ctx context.Context,
	req *pb.Empty,
) (*pb.ExchangeRatesResponse, error) {
	a.log.Info("Получен запрос GetExchangeRates")

	rows, err := a.svc.GetAllExchangeRates(ctx)
	if err != nil {
		a.log.Errorf("Ошибка при получении всех курсов обмена: %v", err)
		return nil, err
	}

	rates := convertExchangeRates(rows)

	a.log.Infof("Возвращаем %d курсов обмена", len(rates))
	return &pb.ExchangeRatesResponse{
		Rates: rates,
	}, nil
}

// convertExchangeRates преобразует срез ExchangeRate в map[string]float32
func convertExchangeRates(rows []models.ExchangeRate) map[string]float32 {
	rates := make(map[string]float32, len(rows))
	for _, r := range rows {
		rates[r.ToCurrency] = float32(r.Rate)
	}
	return rates
}
