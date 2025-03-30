package domains

import (
	"context"
	"time"

	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	ports "github.com/cybrarymin/gRPC/server/internals/domains/ports"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type TransactionService struct {
	ports.TransactionRepositoryPort
	ports.BankAccountRepositoryPort
	*zerolog.Logger
}

func NewTransactionService(repoPort ports.TransactionRepositoryPort, accountPort ports.BankAccountRepositoryPort, logger *zerolog.Logger) *TransactionService {
	return &TransactionService{
		repoPort,
		accountPort,
		logger,
	}
}

func (s *TransactionService) NewTransaction(ctx context.Context, accUUID uuid.UUID, amount float64, TRType string, note string) (*domains.BankTransaction, error) {
	sCtx, nSpan := otel.Tracer("NewTransaction").Start(ctx, "NewTransaction.service.span")
	defer nSpan.End()

	nTransaction := &domains.BankTransaction{
		AccountUUID:          accUUID,
		TransactionTimestamp: time.Now(),
		Amount:               amount,
		TransactionType:      TRType,
		Notes:                note,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	startTime := time.Now()
	nAccountModel, err := s.GetByID(sCtx, accUUID)
	if err != nil {
		s.Logger.Error().
			Err(err).
			Str("account_uuid", accUUID.String()).
			Dur("duration_ms", time.Since(startTime)).
			Msg("failed to retrieve account details")
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to get user information")
		return nil, err
	}

	nAccount := &domains.BankAccount{
		AccountUUID:    nAccountModel.AccountUUID,
		AccountNumber:  nAccountModel.AccountNumber,
		AccountName:    nAccountModel.AccountName,
		Currency:       nAccountModel.Currency,
		CurrentBalance: nAccountModel.CurrentBalance,
		CreatedAt:      nAccountModel.CreatedAt,
		UpdatedAt:      nAccountModel.UpdatedAt,
	}

	startTime = time.Now()
	switch nTransaction.TransactionType {
	case domains.TRDepositType:
		nAccount.CurrentBalance += amount

		_, err := s.Update(sCtx, accUUID, nAccount)
		if err != nil {

			// ERROR log if update fails
			s.Logger.Error().
				Err(err).
				Str("account_uuid", accUUID.String()).
				Str("transaction_type", nTransaction.TransactionType).
				Float64("amount", amount).
				Float64("attempted_balance", nAccount.CurrentBalance).
				Msg("failed to update account balance")
			nSpan.RecordError(err)
			nSpan.SetStatus(codes.Error, "failed to update account balance during transaction")
			return nil, err
		}

	default:
		nAccount.CurrentBalance -= amount

		_, err := s.Update(sCtx, accUUID, nAccount)

		if err != nil {
			// ERROR log if update fails
			s.Logger.Error().
				Err(err).
				Str("account_uuid", accUUID.String()).
				Str("transaction_type", nTransaction.TransactionType).
				Float64("amount", amount).
				Float64("attempted_balance", nAccount.CurrentBalance).
				Msg("failed to update account balance")

			nSpan.RecordError(err)
			nSpan.SetStatus(codes.Error, "failed to update account balance during transaction")
			return nil, err
		}
	}

	txnStartTime := time.Now()

	createdTransaction, err := s.CreateTransaction(sCtx, nTransaction)
	if err != nil {
		s.Logger.Error().
			Err(err).
			Str("account_uuid", nTransaction.AccountUUID.String()).
			Str("transaction_type", nTransaction.TransactionType).
			Float64("amount", nTransaction.Amount).
			Dur("duration_ms", time.Since(txnStartTime)).
			Msg("failed to create transaction record")
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed create new transaction")
		return nil, err
	}

	// INFO log for overall success
	s.Logger.Info().
		Str("transaction_uuid", createdTransaction.TransactionUUID.String()).
		Str("account_uuid", nTransaction.AccountUUID.String()).
		Str("transaction_type", nTransaction.TransactionType).
		Float64("amount", nTransaction.Amount).
		Float64("new_balance", nAccount.CurrentBalance).
		Dur("total_duration_ms", time.Since(startTime)).
		Msg("transaction completed successfully")

	nTransaction.TransactionUUID = createdTransaction.TransactionUUID
	return nTransaction, nil
}
