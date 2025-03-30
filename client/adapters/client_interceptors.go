package client_adapters

import (
	"context"

	"github.com/cybrarymin/gRPC/protogen/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func BasicClientUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if method == "GetCurrentBalance" {
			// do something here
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

type WrappedClientStream struct {
	grpc.ClientStream
}

func BasicStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// adding new metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "new-metadata-key", "new metadata value")
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		interceptedClientStream := &WrappedClientStream{
			clientStream,
		}

		return interceptedClientStream, nil
	}
}

func (w *WrappedClientStream) RecvMsg(m interface{}) error {
	err := w.RecvMsg(m)
	if err != nil {
		return err
	}
	switch m.(type) {
	case pb.BankTransferRequest:
		// do something

	}
	return nil
}

func (w *WrappedClientStream) SendMsg(m interface{}) error {
	switch m.(type) {
	case pb.BankTransferRequest:
		// do something
	}
	err := w.SendMsg(m)
	if err != nil {
		return err
	}
	return nil
}
