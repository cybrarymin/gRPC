package client_ports

import (
	"context"
)

type GrpcClientPort interface {
	GetCurrentBalance(ctx context.Context, accountID string) (float64, error)
}
