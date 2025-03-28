package client_services

import (
	"context"
	"fmt"
	"io"

	client_ports "github.com/cybrarymin/gRPC/client/internals/domains/ports"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/status"
)

type ReferencePorts struct {
	client_ports.GrpcClientPort
}

// Bank client service which uses client_ports to have access to gRPC adapter client functions
type BankCliService struct {
	port   ReferencePorts
	logger *zerolog.Logger
}

// Create new bank client service
func NewBankCliService(grpcPort client_ports.GrpcClientPort, logger *zerolog.Logger) *BankCliService {
	return &BankCliService{
		port: ReferencePorts{
			grpcPort,
		},
		logger: logger,
	}
}

// shows the current balance of the specified account with its uuid identifier.
func (bcs *BankCliService) ShowCurrentBalance(pCtx context.Context, accUUID string) {
	ctx, cancel := context.WithCancel(pCtx)
	defer cancel()

	balance, err := bcs.port.GetCurrentBalance(ctx, accUUID)

	if err != nil {
		// err is coming from gRPC server. Which we have coded in gRPC server to use status.Error().
		// we use status.Convert() to conver thte error to status.
		st := status.Convert(err)

		bcs.logger.Error().Err(fmt.Errorf("%s", st.Message())).
			Str("status", st.Code().String()).
			Str("account_uuid", accUUID).
			Send()
		//return
	}
	fmt.Printf(`{ "balance": %v}`, balance)

}

func (bcs *BankCliService) ShowExchangeRate(pCtx context.Context, fromCurrency string, toCurrency string, amount float64) {
	ctx, cancel := context.WithCancel(pCtx)
	defer cancel()
	streamResp, err := bcs.port.ShowExchangeRate(ctx, fromCurrency, toCurrency, amount)
	if err != nil {
		st := status.Convert(err)

		bcs.logger.Error().Err(fmt.Errorf("%s", st.Message())).
			Str("status", st.Code().String()).
			Str("fromCurrency", fromCurrency).
			Str("toCurrency", toCurrency).
			Send()
	}
	defer streamResp.Close()
	for {
		jsonResp, err := streamResp.Next()
		if err != nil {
			st := status.Convert(err)
			bcs.logger.Error().Err(fmt.Errorf("%s", st.Message())).
				Str("status", st.Code().String()).
				Str("fromCurrency", fromCurrency).
				Str("toCurrency", toCurrency).
				Send()

			if err == io.EOF {
				bcs.logger.Error().Err(fmt.Errorf("received unexpted response from the server"))
				break
			}
		}
		fmt.Println(string(jsonResp))
	}
}
