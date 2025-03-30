package client_adapters

import (
	"context"
	"fmt"

	"github.com/cybrarymin/gRPC/protogen/pb"
)

func (bca *BankGrpcClientAdapter) GetCurrentBalance(ctx context.Context, accountID string) (float64, error) {
	// calling function using our circuit breaker
	resp, err := bca.circuitBreaker.Call(func() (any, error) {
		return bca.client.GetCurrentBalance(ctx, &pb.CurrentBalanceRequest{
			AccountUUID: accountID,
		})
	})

	if err != nil {
		return 0, err
	}

	// Safe type assertion with ok check
	balanceResp, ok := resp.(*pb.CurrentBalanceResponse)
	if !ok {
		bca.logger.Error().
			Str("account_id", accountID).
			Str("type", fmt.Sprintf("%T", resp)).
			Msg("unexpected response type from circuit breaker")
		return 0, fmt.Errorf("unexpected response type: %T", resp)
	}

	return balanceResp.CurrentBalance, nil
}
