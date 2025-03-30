package adapters

import (
	"context"
	"strings"

	"github.com/cybrarymin/gRPC/protogen/pb"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RpcReqID string

var (
	RpcCtxRequestIDKey RpcReqID = "request_id"
)

func requestIDGenerator() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		reqID := uuid.New()
		nCtx := context.WithValue(ctx, RpcCtxRequestIDKey, reqID.String())
		return handler(nCtx, req)
	}
}

func logReqUnaryInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		names := strings.Split(info.FullMethod, "/")
		logger.Info().
			Interface("request_id", ctx.Value(RpcCtxRequestIDKey)).
			Str("grpc_service", names[1]).
			Str("grpc_method", names[2]).
			Interface("request_info", req).
			Msg("grpc request")

		return handler(ctx, req)
	}
}

func ExampleBaiscUnaryServerInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// re: is the request object in gRPC
		// info: has some information about gRPC like FullMethod name which is /package.servicename/methodname
		// handler: is the actual gRPC method which is going to get run and provide a response for the client
		logger.Info().
			Str("grpc_method_name", info.FullMethod).
			Interface("request_info", req).
			Msg("grpc request")

		// fetch metadata from the request context
		reqMeta, exists := metadata.FromIncomingContext(ctx)
		if exists {
			//.... some process on reqMeta if we want
			reqMeta.Set("new-request-metadata", "version1")
		}

		// handler will run gRPC method and prepare the response. When we return the response gRPC sends the response back to the gRPC client
		resp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}

		// create some new outgoing metadata
		respMeta, exists := metadata.FromOutgoingContext(ctx)
		if exists {
			respMeta = metadata.New(nil) // delete all the outgoing metadatas and set new metadatas
		}
		respMeta.Set("test-metadata-for-response", "version1")
		respMeta.Set("test-metadata-for-response", "version2")

		// Set the response metadat in ctx
		grpc.SetHeader(ctx, respMeta)

		// In case you want to change anything from within the response u should check for the resp type first
		switch response := resp.(type) {
		case *pb.CurrentBalanceResponse:
			response.CurrentBalance += 10
		}

		return resp, nil
	}
}

func BasicStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Create a wrapper around the server stream to intercept messages
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
		}

		// handler will use two function of the stream which is stream.RecvMsg and stream.SendMsg.
		// If we use ss.RecvMsg() directly we recieve a message from client and we are fully responsible for processing of that message and we should map that message to specific gRPC method and so on.
		// moreover when we use ss.RecvMsg() directly handler won't have that message anymore.
		// because of this we use a new type and create two new methods of RecvMsg and SendMsg then our handler will be able to use those calls and we are also able to process request streams or respons streams
		return handler(srv, wrappedStream)
	}
}

// wrappedServerStream wraps grpc.ServerStream to allow intercepting stream messages
type wrappedServerStream struct {
	grpc.ServerStream
}

// RecvMsg intercepts incoming messages
func (w *wrappedServerStream) RecvMsg(msg interface{}) error {
	if err := w.ServerStream.RecvMsg(msg); err != nil {
		return err
	}

	// Here you can modify the received message
	// You need to type assert to the specific message type
	switch m := msg.(type) {
	case *pb.BankTransferRequest:
		// Modify request fields as needed
		reqMeta, exists := metadata.FromIncomingContext(w.Context())
		if exists {
			reqMeta.Set("test-stream-interceptor", m.Currency.String())
		}
		w.SetHeader(reqMeta)
	}

	return nil
}

// SendMsg intercepts outgoing messages
func (w *wrappedServerStream) SendMsg(msg interface{}) error {
	// Here you can modify the response message before sending
	switch m := msg.(type) {
	case *pb.BankTransferResponse:
		// Modify response fields as needed
		m.Amount += 100
	}

	return w.ServerStream.SendMsg(msg)
}
