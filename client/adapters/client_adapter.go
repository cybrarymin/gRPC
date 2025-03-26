package client_adapters

import (
	"context"

	"github.com/cybrarymin/gRPC/protogen/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type BankGrpcClientAdapter struct {
	logger *zerolog.Logger
	client pb.BankServiceClient
}

func NewBankGrpcClientAdapter(conn *grpc.ClientConn, logger *zerolog.Logger) (*BankGrpcClientAdapter, error) {

	client := pb.NewBankServiceClient(conn)

	return &BankGrpcClientAdapter{
		logger: logger,
		client: client,
	}, nil
}

func (bca *BankGrpcClientAdapter) GetCurrentBalance(ctx context.Context, accountID string) (float64, error) {
	bca.logger.Info().
		Str("account_id", accountID).
		Msg("fetching account current balance...")

	resp, err := bca.client.GetCurrentBalance(ctx, &pb.CurrentBalanceRequest{
		AccountUUID: accountID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return 0, st.Err()
	}
	return resp.CurrentBalance, nil

}
