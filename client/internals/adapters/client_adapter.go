package client_adapters

import (
	"github.com/cybrarymin/gRPC/protogen/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type BankGrpcClientAdapter struct {
	logger         *zerolog.Logger
	client         pb.BankServiceClient
	circuitBreaker *CircuitBreaker
}

func NewBankGrpcClientAdapter(conn *grpc.ClientConn, logger *zerolog.Logger, cb *CircuitBreaker) (*BankGrpcClientAdapter, error) {

	client := pb.NewBankServiceClient(conn)

	return &BankGrpcClientAdapter{
		logger:         logger,
		client:         client,
		circuitBreaker: cb,
	}, nil
}
