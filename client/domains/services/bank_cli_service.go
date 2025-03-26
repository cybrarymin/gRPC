package client_services

import (
	"context"
	"fmt"

	client_ports "github.com/cybrarymin/gRPC/client/domains/ports"
	"github.com/rs/zerolog"
)

type BankCliService struct {
	port   client_ports.GrpcClientPort
	logger *zerolog.Logger
}

func NewBankCliService(port client_ports.GrpcClientPort, logger *zerolog.Logger) *BankCliService {
	return &BankCliService{
		port:   port,
		logger: logger,
	}
}

func (bcs *BankCliService) ShowCurrentBalance(ctx context.Context, accUUID string) {
	balance, err := bcs.port.GetCurrentBalance(ctx, accUUID)
	if err != nil {
		bcs.logger.Error().Err(err).
			Str("account_uuid", accUUID).
			Msg("failed to get current balance for specified account")
	}
	fmt.Printf(`{ "balance": %v}`, balance)
}
