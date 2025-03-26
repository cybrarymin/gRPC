package domains

import (
	"context"
	"time"

	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	ports "github.com/cybrarymin/gRPC/server/internals/domains/ports"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
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
	// DEBUG log for new bank transaction process with provided input
	s.Logger.Debug().
		Str("account_uuid", accUUID.String()).
		Float64("amount", amount).
		Str("transaction_type", TRType).
		Str("note", note).
		Msg("starting new bank account transaction process with provided inputs")

	nTransaction := &domains.BankTransaction{
		AccountUUID:          accUUID,
		TransactionTimestamp: time.Now(),
		Amount:               amount,
		TransactionType:      TRType,
		Notes:                note,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// DEBUG log for new bank transaction object
	s.Logger.Debug().
		Str("account_uuid", nTransaction.AccountUUID.String()).
		Float64("amount", nTransaction.Amount).
		Str("transaction_type", nTransaction.TransactionType).
		Str("note", nTransaction.Notes).
		Time("transaction_timestamp", nTransaction.TransactionTimestamp).
		Time("created_at", nTransaction.CreatedAt).
		Time("updated_at", nTransaction.UpdatedAt).
		Msg("create  bank account transaction object for transaction creation process")

	// DEBUG log before account retrieval
	s.Logger.Debug().
		Str("account_uuid", accUUID.String()).
		Msg("retrieving account details")

	startTime := time.Now()
	nAccountModel, err := s.GetByID(ctx, accUUID)
	if err != nil {
		s.Logger.Error().
			Err(err).
			Str("account_uuid", accUUID.String()).
			Dur("duration_ms", time.Since(startTime)).
			Msg("failed to retrieve account details")
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

	// DEBUG log for account retrieval success
	s.Logger.Debug().
		Str("account_uuid", nAccountModel.AccountUUID.String()).
		Str("account_number", nAccountModel.AccountNumber).
		Str("account_name", nAccountModel.AccountName).
		Float64("current_balance", nAccountModel.CurrentBalance).
		Dur("duration_ms", time.Since(startTime)).
		Msg("account details retrieved successfully")

	// DEBUG log for balance calculation process
	s.Logger.Info().
		Str("account_uuid", nAccountModel.AccountUUID.String()).
		Str("account_number", nAccountModel.AccountNumber).
		Str("account_name", nAccountModel.AccountName).
		Float64("current_balance", nAccountModel.CurrentBalance).
		Str("transaction_type", nTransaction.TransactionType).
		Float64("transaction_amount", nTransaction.Amount).
		Msg("starting balance calculation process..")

	startTime = time.Now()
	switch nTransaction.TransactionType {
	case domains.TRDepositType:
		nAccount.CurrentBalance += amount
		// DEBUG log for transaction calculation
		s.Logger.Debug().
			Float64("previous_balance", nAccount.CurrentBalance-amount).
			Float64("deposit_amount", amount).
			Float64("new_balance", nAccount.CurrentBalance).
			Msg("deposit calculation completed")

		_, err := s.Update(ctx, accUUID, nAccount)
		if err != nil {

			// ERROR log if update fails
			s.Logger.Error().
				Err(err).
				Str("account_uuid", accUUID.String()).
				Str("transaction_type", nTransaction.TransactionType).
				Float64("amount", amount).
				Float64("attempted_balance", nAccount.CurrentBalance).
				Msg("failed to update account balance")
			return nil, err
		}

	default:
		nAccount.CurrentBalance -= amount
		// DEBUG log for transaction calculation
		s.Logger.Debug().
			Float64("previous_balance", nAccount.CurrentBalance+amount).
			Float64("withdrawal_amount", amount).
			Float64("new_balance", nAccount.CurrentBalance).
			Msg("deposit calculation completed")
		_, err := s.Update(ctx, accUUID, nAccount)

		if err != nil {
			// ERROR log if update fails
			s.Logger.Error().
				Err(err).
				Str("account_uuid", accUUID.String()).
				Str("transaction_type", nTransaction.TransactionType).
				Float64("amount", amount).
				Float64("attempted_balance", nAccount.CurrentBalance).
				Msg("failed to update account balance")
			return nil, err
		}
	}
	// DEBUG log for balance calculation process
	s.Logger.Info().
		Str("account_uuid", nAccountModel.AccountUUID.String()).
		Str("account_number", nAccountModel.AccountNumber).
		Str("account_name", nAccountModel.AccountName).
		Float64("current_balance", nAccountModel.CurrentBalance).
		Str("transaction_type", nTransaction.TransactionType).
		Float64("transaction_amount", nTransaction.Amount).
		Dur("duration_ms", time.Since(startTime)).
		Msg("finished balance calculation and update process..")

	// DEBUG log before transaction creation
	txnStartTime := time.Now()
	s.Logger.Debug().
		Str("account_uuid", nTransaction.AccountUUID.String()).
		Float64("amount", nTransaction.Amount).
		Str("transaction_type", nTransaction.TransactionType).
		Msg("creating transaction record")

	createdTransaction, err := s.CreateTransaction(ctx, nTransaction)
	if err != nil {
		s.Logger.Error().
			Err(err).
			Str("account_uuid", nTransaction.AccountUUID.String()).
			Str("transaction_type", nTransaction.TransactionType).
			Float64("amount", nTransaction.Amount).
			Dur("duration_ms", time.Since(txnStartTime)).
			Msg("failed to create transaction record")
		return nil, err
	}

	// DEBUG log for transaction creation success
	s.Logger.Debug().
		Str("transaction_uuid", nTransaction.TransactionUUID.String()).
		Str("account_uuid", nTransaction.AccountUUID.String()).
		Dur("duration_ms", time.Since(txnStartTime)).
		Msg("transaction record created successfully")

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
