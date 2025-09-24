package services

import (
	"context"

	"github.com/sbilibin2017/gw-exchanger/internal/models"
	"go.uber.org/zap"
)

// ExchangeRateReader интерфейс для чтения курсов валют
type ExchangeRateReader interface {
	Get(ctx context.Context, fromCurrency, toCurrency string) (*float64, error)
	List(ctx context.Context) ([]models.ExchangeRate, error)
}

// ExchangeRateService реализует бизнес-логику для работы с курсами валют
type ExchangeRateService struct {
	reader ExchangeRateReader
	log    *zap.SugaredLogger
}

// NewExchangeRateService создаёт новый сервис с логгером
func NewExchangeRateService(log *zap.SugaredLogger, reader ExchangeRateReader) *ExchangeRateService {
	return &ExchangeRateService{
		reader: reader,
		log:    log,
	}
}

// GetExchangeRate возвращает курс обмена между двумя валютами
func (s *ExchangeRateService) GetExchangeRate(
	ctx context.Context,
	fromCurrency, toCurrency string,
) (*float32, error) {
	s.log.Infof("Получение курса обмена: %s -> %s", fromCurrency, toCurrency)

	rate, err := s.reader.Get(ctx, fromCurrency, toCurrency)
	if err != nil {
		s.log.Errorf("Ошибка при получении курса обмена: %v", err)
		return nil, err
	}
	if rate == nil {
		s.log.Warnf("Курс обмена не найден: %s -> %s", fromCurrency, toCurrency)
		return nil, nil
	}

	rateF32 := float32(*rate)

	s.log.Infof("Курс обмена получен: %s -> %s = %f", fromCurrency, toCurrency, *rate)
	return &rateF32, nil
}

// GetAllExchangeRates возвращает карту всех курсов валют
func (s *ExchangeRateService) GetAllExchangeRates(
	ctx context.Context,
) ([]models.ExchangeRate, error) {
	s.log.Info("Получение всех курсов обмена")

	rates, err := s.reader.List(ctx)
	if err != nil {
		s.log.Errorf("Ошибка при получении всех курсов обмена: %v", err)
		return nil, err
	}

	s.log.Infof("Всего получено курсов обмена: %d", len(rates))
	return rates, nil
}
