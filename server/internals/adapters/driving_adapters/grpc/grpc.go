package adapters

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/cybrarymin/gRPC/protogen/pb"
	domains "github.com/cybrarymin/gRPC/server/internals/domains/ports"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcPortReference struct {
	domains.BankAccountGrpcPort
	domains.TransactionGrpcPort
	domains.BankExchangeRateGrpcPort
	domains.BankTransferGrpcPort
	domains.ValidatorGrpcPort
}

type GrpcAdapter struct {
	port     GrpcPortReference
	Srv      *grpc.Server
	grpcPort string
	grpcHost string
	logger   *zerolog.Logger
	pb.BankServiceServer
}

func (ad *GrpcAdapter) OpenAccount(ctx context.Context, req *pb.BankAccountCreateRequest) (*pb.BankAccountCreateResponse, error) {
	ad.logger.Info().
		Str("account_name", req.AccountName).
		Str("account_number", req.AccountNumber).
		Str("currency", req.Currency.String()).
		Float64("balance", req.CurrentBalance).
		Msg("received open account request")

	ad.port.Validate(len(req.AccountNumber) == 10, "account_number", "account number length should be 10 digit")
	ad.port.Validate(req.CurrentBalance >= 0, "account_balance", "account balance shouldn't be a negative number")
	if _, exists := pb.Currency_value[req.Currency.String()]; !exists {
		ad.port.AddError("currency", "unsupported currency")
	}

	if !ad.port.Valid() {
		ad.logger.Error().
			Interface("validation_errors", ad.port.ValidatorErrors()).
			Msg("account creation validation failed")
		return nil, StatusCheck(ad.port.ValidatorErrors())
	}

	createdAcc, err := ad.port.OpenAccount(ctx, req.AccountName, req.AccountNumber, req.Currency.String(), req.CurrentBalance)
	if err != nil {
		ad.logger.Error().Err(err).
			Str("account_name", req.AccountName).
			Str("account_number", req.AccountNumber).
			Msg("failed to create account")
		return nil, StatusCheck(err)
	}

	ad.logger.Info().
		Str("account_uuid", createdAcc.AccountUUID.String()).
		Str("account_number", createdAcc.AccountNumber).
		Msg("account created successfully")
	return &pb.BankAccountCreateResponse{
		AccountUUID:    createdAcc.AccountUUID.String(),
		AccountNumber:  createdAcc.AccountNumber,
		AccountName:    createdAcc.AccountName,
		Currency:       pb.Currency(pb.Currency_value[createdAcc.Currency]),
		CurrentBalance: createdAcc.CurrentBalance,
		CreatedAt:      timestamppb.New(createdAcc.CreatedAt),
		UpdatedAt:      timestamppb.New(createdAcc.UpdatedAt),
	}, nil
}

func (ad *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *pb.CurrentBalanceRequest) (*pb.CurrentBalanceResponse, error) {
	ad.logger.Info().
		Str("account_uuid", req.AccountUUID).
		Msg("received balance check request")

	accUUID, err := uuid.Parse(req.AccountUUID)
	if err != nil {
		ad.logger.Error().Err(err).
			Str("account_uuid", req.AccountUUID).
			Msg("invalid account UUID format")
		ad.port.AddError("account_uuid", err.Error())
	}
	if !ad.port.Valid() {
		ad.logger.Error().
			Interface("validation_errors", ad.port.ValidatorErrors()).
			Msg("balance check validation failed")
		return nil, StatusCheck(ad.port.ValidatorErrors())
	}

	balance, currency, err := ad.port.GetCurrentBalance(ctx, accUUID)
	if err != nil {
		ad.logger.Error().Err(err).
			Str("account_uuid", req.AccountUUID).
			Msg("failed to get account balance")
		return nil, StatusCheck(err)
	}

	ad.logger.Info().
		Str("account_uuid", req.AccountUUID).
		Float64("balance", balance).
		Str("currency", currency).
		Msg("balance retrieved successfully")
	return &pb.CurrentBalanceResponse{
		AccountUUID:    req.AccountUUID,
		Currency:       pb.Currency(pb.Currency_value[currency]),
		CurrentBalance: balance,
	}, nil
}

func (ad *GrpcAdapter) CreateTransaction(ctx context.Context, req *pb.BankTransactionCreateRequest) (*pb.BankTransactionCreateResponse, error) {
	ad.logger.Info().
		Str("account_uuid", req.AccountUUID).
		Float64("amount", req.Amount).
		Str("type", req.TransactionType.String()).
		Msg("received create transaction request")

	ad.port.Validate(req.Amount >= 0, "transaction_amount", "transaction amount can't be negative")
	if _, exists := pb.TransactionType_value[req.TransactionType.String()]; !exists {
		ad.port.Validate(exists, "transaction_type", "unsupported transaction type")
	}

	acUUID, err := uuid.Parse(req.AccountUUID)
	if err != nil {
		ad.port.AddError("account_uuid", err.Error())
	}

	if !ad.port.Valid() {
		ad.logger.Error().
			Interface("validation_errors", ad.port.ValidatorErrors()).
			Msg("transaction creation validation failed")
		return nil, StatusCheck(ad.port.ValidatorErrors())
	}

	nTransaction, err := ad.port.NewTransaction(ctx, acUUID, req.Amount, req.TransactionType.String(), req.Notes)
	if err != nil {
		ad.logger.Error().Err(err).
			Str("account_uuid", req.AccountUUID).
			Float64("amount", req.Amount).
			Str("type", req.TransactionType.String()).
			Msg("failed to create transaction")
		return nil, StatusCheck(err)
	}

	ad.logger.Info().
		Str("transaction_uuid", nTransaction.TransactionUUID.String()).
		Str("account_uuid", nTransaction.AccountUUID.String()).
		Float64("amount", nTransaction.Amount).
		Msg("transaction created successfully")
	return &pb.BankTransactionCreateResponse{
		TransactionUUID:      nTransaction.TransactionUUID.String(),
		AccountUUID:          nTransaction.AccountUUID.String(),
		Amount:               nTransaction.Amount,
		TransactionType:      pb.TransactionType(pb.TransactionType_value[nTransaction.TransactionType]),
		Notes:                nTransaction.Notes,
		TransactionTimestamp: timestamppb.New(nTransaction.TransactionTimestamp),
		CreatedAt:            timestamppb.New(nTransaction.CreatedAt),
		UpdatedAt:            timestamppb.New(nTransaction.UpdatedAt),
	}, nil
}

func (ad *GrpcAdapter) GetExchangeRate(req *pb.ExchangeRateRequest, stream grpc.ServerStreamingServer[pb.ExchangeRateResponse]) error {

	ad.logger.Info().
		Str("from_currency", req.FromCurrency.String()).
		Str("to_currency", req.ToCurrency.String()).
		Float64("amount", req.Amount).
		Msg("started exchange rate stream")

	ad.port.Validate(req.Amount >= 0, "amount", "exchange amount shouldn't be negative")
	if _, exists := pb.Currency_value[req.FromCurrency.String()]; !exists {
		ad.port.Validate(!exists, "from_currency", "unsupported currency type")
	}
	if _, exists := pb.Currency_value[req.ToCurrency.String()]; !exists {
		ad.port.Validate(!exists, "to_currency", "unsupported currency type")
	}

	if !ad.port.Valid() {
		ad.logger.Error().
			Interface("validation_errors", ad.port.ValidatorErrors()).
			Msg("exchange rate validation failed")
		return StatusCheck(ad.port.ValidatorErrors())
	}

	total, err := ad.port.CalculateRate(stream.Context(), req.FromCurrency.String(), req.ToCurrency.String(), req.Amount)
	if err != nil {
		ad.logger.Error().Err(err).
			Str("from_currency", req.FromCurrency.String()).
			Str("to_currency", req.ToCurrency.String()).
			Float64("amount", req.Amount).
			Msg("failed to calculate exchange rate")
		return StatusCheck(err)
	}
	resp := &pb.ExchangeRateResponse{
		Currency: req.ToCurrency.String(),
		Amount:   total,
	}

	for {
		select {
		case <-stream.Context().Done():
			ad.logger.Info().Msg("client canceled exchange rate stream")
			return stream.Context().Err()
		default:
			err := stream.Send(resp)
			if err != nil {
				ad.logger.Error().Err(err).Msg("failed to send exchange rate response")
				return StatusCheck(err)
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (ad *GrpcAdapter) CreateTransfers(stream grpc.BidiStreamingServer[pb.BankTransferRequest, pb.BankTransferResponse]) error {
	ctx := stream.Context()
	ad.logger.Info().Msg("started bidirectional transfer stream")

	for {
		select {
		case <-ctx.Done():
			ad.logger.Info().Msg("client canceled transfer stream")
			return ctx.Err()
		default:
			req, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					ad.logger.Info().Msg("transfer stream completed")
					return nil
				}
				ad.logger.Error().Err(err).Msg("failed to read data from client stream")
				return err
			}

			ad.logger.Info().
				Str("from_account", req.FromAccount).
				Str("to_account", req.ToAccount).
				Float64("amount", req.Amount).
				Str("currency", req.Currency.String()).
				Msg("received transfer request")

			ad.port.Validate(req.Amount >= 0, "amount", "transfer amount shouldn't be a negative number")
			ad.port.Validate(req.Currency != pb.Currency_Currency_UNSPECEFIED, "currency", "unsupported currency type")

			fromAccountUUID, err := uuid.Parse(req.FromAccount)
			if err != nil {
				ad.port.AddError("from_account", err.Error())
			}
			toAccountUUID, err := uuid.Parse(req.ToAccount)
			if err != nil {
				ad.port.AddError("to_account", err.Error())
			}

			if !ad.port.Valid() {
				ad.logger.Error().
					Interface("validation_errors", ad.port.ValidatorErrors()).
					Msg("transfer validation failed")
				return StatusCheck(ad.port.ValidatorErrors())
			}

			nTransfer, err := ad.port.TransferMoney(stream.Context(), fromAccountUUID, toAccountUUID, req.Currency.String(), req.Amount)
			if err != nil {
				ad.logger.Error().Err(err).
					Str("from_account", req.FromAccount).
					Str("to_account", req.ToAccount).
					Float64("amount", req.Amount).
					Msg("money transfer failed")
				return StatusCheck(err)
			}

			ad.logger.Info().
				Str("from_account", nTransfer.FromAccountUUID.String()).
				Str("to_account", nTransfer.ToAccountUUID.String()).
				Float64("amount", nTransfer.Amount).
				Msg("transfer completed successfully")

			resp := &pb.BankTransferResponse{
				FromAccount:    nTransfer.FromAccountUUID.String(),
				ToAccount:      nTransfer.ToAccountUUID.String(),
				Currency:       pb.Currency(pb.Currency_value[nTransfer.Currency]),
				Amount:         nTransfer.Amount,
				Time:           timestamppb.New(nTransfer.TransferTimestamp),
				TransferStatus: pb.TransferStatus_Succes,
			}
			err = stream.Send(resp)
			if err != nil {
				ad.logger.Error().Err(err).Msg("failed to send the grpc response to the client")
				return StatusCheck(err)
			}
		}
	}
}

func NewGrpcAdapter(grpcHost string, grpcPort string, logger *zerolog.Logger, port GrpcPortReference) *GrpcAdapter {
	srv := grpc.NewServer()
	ad := &GrpcAdapter{
		port:     port,
		grpcPort: grpcPort,
		grpcHost: grpcHost,
		logger:   logger,
		Srv:      srv,
	}
	pb.RegisterBankServiceServer(srv, ad)
	reflection.Register(srv)
	return ad
}

func (ad *GrpcAdapter) Run() error {
	listenAddr, err := net.Listen("tcp4", ad.grpcHost+":"+ad.grpcPort)
	if err != nil {
		ad.logger.Error().Err(err).Msg("failed to listen on address")
		return err
	}

	ad.logger.Info().Msgf("starting grpc server on %s:%s", ad.grpcHost, ad.grpcPort)
	err = ad.Srv.Serve(listenAddr)
	if err != nil {
		ad.logger.Error().Err(err).Msg("server failed to serve")
		return err
	}

	return nil
}

func (ad *GrpcAdapter) Stop() {
	ad.logger.Info().Msg("gracefully stopping gRPC server")

	// Give ongoing requests a chance to complete
	stopped := make(chan struct{})
	go func() {
		ad.Srv.GracefulStop()
		close(stopped)
	}()

	// Set a timeout for graceful shutdown
	t := time.NewTimer(10 * time.Second)
	select {
	case <-stopped:
		ad.logger.Info().Msg("gRPC server stopped gracefully")
	case <-t.C:
		ad.logger.Warn().Msg("forcing gRPC server to stop")
		ad.Srv.Stop()
	}
}
