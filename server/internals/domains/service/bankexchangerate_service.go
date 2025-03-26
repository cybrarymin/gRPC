package domains

import (
	"context"

	domains "github.com/cybrarymin/gRPC/server/internals/domains/ports"
	"github.com/rs/zerolog"
)

type BankExchangeRateService struct {
	port   domains.BankExchangeRateRepositoryPort
	logger *zerolog.Logger
}

func NewBankExchangeRateService(port domains.BankExchangeRateRepositoryPort, logger *zerolog.Logger) *BankExchangeRateService {
	return &BankExchangeRateService{
		port:   port,
		logger: logger,
	}
}

func (s *BankExchangeRateService) CalculateRate(ctx context.Context, fromCurrency string, toCurrency string, amount float64) (float64, error) {
	exRate, err := s.port.GetByCurrencies(ctx, fromCurrency, toCurrency)
	if err != nil {
		s.logger.Error().Err(err).Msgf("couldn't fetch exchange rate from database for %s to %s", fromCurrency, toCurrency)
		return 0, err
	}
	return exRate.Rate * amount, nil
}
