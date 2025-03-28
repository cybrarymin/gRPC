package client_ports

import (
	"context"
)

type ExchangeRateStreamResponsePort interface {
	Next() ([]byte, error)
	Close() error
}

type GrpcClientPort interface {
	GetCurrentBalance(ctx context.Context, accountID string) (float64, error)
	ShowExchangeRate(ctx context.Context, fromCurrency string, toCurrency string, amount float64) (ExchangeRateStreamResponsePort, error)
}
