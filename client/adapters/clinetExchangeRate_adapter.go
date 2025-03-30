package client_adapters

import (
	"context"
	"fmt"

	client_ports "github.com/cybrarymin/gRPC/client/internals/domains/ports"
	"github.com/cybrarymin/gRPC/protogen/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type ExchangeRateStreamResponse struct {
	stream grpc.ServerStreamingClient[pb.ExchangeRateResponse]
}

func NewExchangeRateStreamResponse(stream grpc.ServerStreamingClient[pb.ExchangeRateResponse]) *ExchangeRateStreamResponse {
	return &ExchangeRateStreamResponse{
		stream: stream,
	}
}

func (exr *ExchangeRateStreamResponse) Next() ([]byte, error) {

	resp, err := exr.stream.Recv()
	if err != nil {
		return nil, err
	}
	jsonRes, err := protojson.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return jsonRes, nil
}

func (exr *ExchangeRateStreamResponse) Close() error {
	return nil
}

func (bca *BankGrpcClientAdapter) ShowExchangeRate(ctx context.Context, fromCurrency string, toCurrency string, amount float64) (client_ports.ExchangeRateStreamResponsePort, error) {

	req := &pb.ExchangeRateRequest{
		FromCurrency: pb.Currency(pb.Currency_value[fromCurrency]),
		ToCurrency:   pb.Currency(pb.Currency_value[toCurrency]),
		Amount:       amount,
	}

	streamResp, err := bca.circuitBreaker.Call(func() (any, error) {
		return bca.client.GetExchangeRate(ctx, req)
	})

	if err != nil {
		return nil, err
	}

	resp, ok := streamResp.(grpc.ServerStreamingClient[pb.ExchangeRateResponse])
	if !ok {
		if !ok {
			bca.logger.Error().
				Str("type", fmt.Sprintf("%T", resp)).
				Msg("unexpected response type from circuit breaker")
			return nil, fmt.Errorf("unexpected response type: %T", resp)
		}
	}

	return NewExchangeRateStreamResponse(resp), nil
}
