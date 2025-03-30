package domains

import (
	"context"
	"errors"
	"fmt"
	"time"

	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	domainErrors "github.com/cybrarymin/gRPC/server/internals/domains/errors"
	ports "github.com/cybrarymin/gRPC/server/internals/domains/ports"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type BankTransferService struct {
	port             ports.BankTransferRepositoryPort
	accountPort      ports.BankAccountRepositoryPort
	transactionPort  ports.TransactionRepositoryPort
	exchangeRatePort ports.BankExchangeRateRepositoryPort
	logger           *zerolog.Logger
	validator        *Validator
}

func NewBankTransferService(port ports.BankTransferRepositoryPort, accountPort ports.BankAccountRepositoryPort, transactionPort ports.TransactionRepositoryPort, exchangeRatePort ports.BankExchangeRateRepositoryPort, logger *zerolog.Logger, validator *Validator) *BankTransferService {
	logger.Debug().Msg("Initializing BankTransferService")
	return &BankTransferService{
		port,
		accountPort,
		transactionPort,
		exchangeRatePort,
		logger,
		validator,
	}
}

func (s *BankTransferService) TransferMoney(ctx context.Context, srcAccount uuid.UUID, dstAccount uuid.UUID, currency string, amount float64) (*domains.BankTransfer, error) {
	sCtx, nSpan := otel.Tracer("TransferMoney").Start(ctx, "TransferMoney.service.span")
	defer nSpan.End()

	s.logger.Info().
		Str("source_account", srcAccount.String()).
		Str("destination_account", dstAccount.String()).
		Str("currency", currency).
		Float64("amount", amount).
		Msg("Starting money transfer")

	startTime := time.Now()

	nTransfer := &domains.BankTransfer{
		FromAccountUUID:   srcAccount,
		ToAccountUUID:     dstAccount,
		Currency:          currency,
		Amount:            amount,
		TransferTimestamp: startTime,
		TransferSucceed:   false,
		CreatedAt:         startTime,
		UpdatedAt:         startTime,
	}

	createdTransfer, err := s.port.CreateTransfer(sCtx, nTransfer)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create a tranfer object in database")
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to create new transfer object in database")
		return nil, err
	}

	dstAccountInfo, err := s.accountPort.GetByID(sCtx, dstAccount)
	if err != nil {
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to get destination account information of the money transfer request")
		return nil, err
	}

	if dstAccountInfo.Currency != currency {
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "non-compliant destination account currency with tranfer request currency")
		return nil, domainErrors.InvalidCurrencyError(currency)
	}

	srcAccountInfo, err := s.accountPort.GetByID(sCtx, srcAccount)
	if err != nil {
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to get source account information of the money transfer request")
		return nil, err
	}

	exchangeRate, err := s.exchangeRatePort.GetByCurrencies(sCtx, srcAccountInfo.Currency, dstAccountInfo.Currency)
	if err != nil {
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to get currencies exchange rate to convert source account curreny to destination account currency")
		return nil, err
	}

	transferAmount := exchangeRate.Rate * amount
	nTransaction := NewTransactionService(s.transactionPort, s.accountPort, s.logger)

	_, err = nTransaction.NewTransaction(sCtx, srcAccount, transferAmount, "Transfer", fmt.Sprintf("transfer to account %s", dstAccount.String()))
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("account", srcAccount.String()).
			Float64("amount", amount).
			Msg("Failed to deduct amount from source account")
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to deduct amount from source account during money transfer")
		return nil, err
	}

	_, err = nTransaction.NewTransaction(ctx, dstAccount, transferAmount, "Deposit", fmt.Sprintf("transfer from account %s", srcAccount.String()))
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("account", dstAccount.String()).
			Float64("amount", amount).
			Msg("Failed to add amount to destination account")
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to add amount from destination account during money transfer")
		return nil, err
	}

	nTransfer.TransferSucceed = true
	nTransfer.TransferUUID = createdTransfer.TransferUUID
	_, err = s.port.UpdateTransfer(sCtx, nTransfer.TransferUUID, nTransfer)

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("from_account", srcAccount.String()).
			Str("to_account", dstAccount.String()).
			Float64("amount", amount).
			Msg("Failed to create transfer record, attempting rollback")

		_, rollbackErr := nTransaction.NewTransaction(ctx, srcAccount, amount, "Deposit", fmt.Sprintf("rollback transfer to account %s", dstAccount.String()))
		if rollbackErr != nil {
			s.logger.Error().
				Err(rollbackErr).
				Str("account", srcAccount.String()).
				Float64("amount", amount).
				Msg("Failed to rollback source account transaction")
			nSpan.RecordError(err)
			nSpan.SetStatus(codes.Error, "failed to rollback source account transaction after money transfer failure")
			return nil, fmt.Errorf("transfer failed and rollback failed: %w", rollbackErr)
		}

		_, rollbackErr = nTransaction.NewTransaction(ctx, dstAccount, amount, "Transfer", fmt.Sprintf("rollback transfer from account %s", srcAccount.String()))
		if rollbackErr != nil {
			s.logger.Error().
				Err(rollbackErr).
				Str("account", dstAccount.String()).
				Float64("amount", amount).
				Msg("Failed to rollback destination account transaction")
			nSpan.RecordError(err)
			nSpan.SetStatus(codes.Error, "failed to rollback destination account transaction after money transfer failure")
			return nil, fmt.Errorf("transfer failed and rollback partially failed: %w", rollbackErr)
		}

		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to change transfer status successful")
		return nil, err
	}

	s.logger.Info().
		Str("transfer_id", nTransfer.TransferUUID.String()).
		Str("from_account", srcAccount.String()).
		Str("to_account", dstAccount.String()).
		Float64("amount", amount).
		Str("currency", currency).
		Msg("Money transfer completed successfully")

	return nTransfer, nil
}
