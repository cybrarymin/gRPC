package domains

import (
	"context"

	adapters "github.com/cybrarymin/gRPC/server/internals/adapters/driven_adapters/database"
)

type BankExchangeRateRepositoryPort interface {
	GetByCurrencies(pCtx context.Context, FromCurrency string, ToCurrency string) (*adapters.ExchangeRateModel, error)
}

type BankExchangeRateGrpcPort interface {
	CalculateRate(ctx context.Context, fromCurrency string, toCurrency string, amount float64) (float64, error)
}
