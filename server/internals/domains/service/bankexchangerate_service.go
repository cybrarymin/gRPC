package domains

import (
	"context"

	domains "github.com/cybrarymin/gRPC/server/internals/domains/ports"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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
	sCtx, nSpan := otel.Tracer("CalculateRate").Start(ctx, "CalculateRate.service.span")
	defer nSpan.End()

	exRate, err := s.port.GetByCurrencies(sCtx, fromCurrency, toCurrency)
	if err != nil {
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to get exchange rate")
		return 0, err
	}
	return exRate.Rate * amount, nil
}
