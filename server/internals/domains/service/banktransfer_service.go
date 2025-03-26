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

	s.logger.Info().
		Str("source_account", srcAccount.String()).
		Str("destination_account", dstAccount.String()).
		Str("currency", currency).
		Float64("amount", amount).
		Msg("Starting money transfer")

	startTime := time.Now()
	s.logger.Debug().
		Time("timestamp", startTime).
		Msg("Creating bank transfer record")

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

	createdTransfer, err := s.port.CreateTransfer(ctx, nTransfer)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create a tranfer object in database")
		return nil, err
	}

	dstAccountInfo, err := s.accountPort.GetByID(ctx, dstAccount)
	if err != nil {
		return nil, err
	}

	if dstAccountInfo.Currency != currency {
		return nil, domainErrors.InvalidCurrencyError(currency)
	}

	srcAccountInfo, err := s.accountPort.GetByID(ctx, srcAccount)
	if err != nil {
		return nil, err
	}

	exchangeRate, err := s.exchangeRatePort.GetByCurrencies(ctx, srcAccountInfo.Currency, dstAccountInfo.Currency)
	if err != nil {
		return nil, err
	}

	transferAmount := exchangeRate.Rate * amount
	s.logger.Debug().Msg("Creating transaction service instance")
	nTransaction := NewTransactionService(s.transactionPort, s.accountPort, s.logger)

	s.logger.Debug().
		Str("account", srcAccount.String()).
		Float64("amount", amount).
		Msg("Deducting amount from source account")

	sourceTransaction, err := nTransaction.NewTransaction(ctx, srcAccount, transferAmount, "Transfer", fmt.Sprintf("transfer to account %s", dstAccount.String()))
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("account", srcAccount.String()).
			Float64("amount", amount).
			Msg("Failed to deduct amount from source account")
		return nil, err
	}
	s.logger.Debug().
		Str("transaction_id", sourceTransaction.TransactionUUID.String()).
		Msg("Successfully deducted amount from source account")

	s.logger.Debug().
		Str("account", dstAccount.String()).
		Float64("amount", amount).
		Msg("Adding amount to destination account")

	destTransaction, err := nTransaction.NewTransaction(ctx, dstAccount, transferAmount, "Deposit", fmt.Sprintf("transfer from account %s", srcAccount.String()))
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("account", dstAccount.String()).
			Float64("amount", amount).
			Msg("Failed to add amount to destination account")
		return nil, err
	}

	s.logger.Debug().
		Str("transaction_id", destTransaction.TransactionUUID.String()).
		Msg("Successfully added amount to destination account")

	nTransfer.TransferSucceed = true
	nTransfer.TransferUUID = createdTransfer.TransferUUID
	_, err = s.port.UpdateTransfer(ctx, nTransfer.TransferUUID, nTransfer)

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("from_account", srcAccount.String()).
			Str("to_account", dstAccount.String()).
			Float64("amount", amount).
			Msg("Failed to create transfer record, attempting rollback")

		s.logger.Debug().
			Str("account", srcAccount.String()).
			Float64("amount", amount).
			Msg("Rolling back: adding amount back to source account")

		rollbackSource, rollbackErr := nTransaction.NewTransaction(ctx, srcAccount, amount, "Deposit", fmt.Sprintf("rollback transfer to account %s", dstAccount.String()))
		if rollbackErr != nil {
			s.logger.Error().
				Err(rollbackErr).
				Str("account", srcAccount.String()).
				Float64("amount", amount).
				Msg("Failed to rollback source account transaction")
			return nil, fmt.Errorf("transfer failed and rollback failed: %w", rollbackErr)
		}

		s.logger.Debug().
			Str("transaction_id", rollbackSource.TransactionUUID.String()).
			Msg("Successfully rolled back source account transaction")

		s.logger.Debug().
			Str("account", dstAccount.String()).
			Float64("amount", amount).
			Msg("Rolling back: deducting amount from destination account")

		rollbackDest, rollbackErr := nTransaction.NewTransaction(ctx, dstAccount, amount, "Transfer", fmt.Sprintf("rollback transfer from account %s", srcAccount.String()))
		if rollbackErr != nil {
			s.logger.Error().
				Err(rollbackErr).
				Str("account", dstAccount.String()).
				Float64("amount", amount).
				Msg("Failed to rollback destination account transaction")

			return nil, fmt.Errorf("transfer failed and rollback partially failed: %w", rollbackErr)
		}

		s.logger.Debug().
			Str("transaction_id", rollbackDest.TransactionUUID.String()).
			Msg("Successfully rolled back destination account transaction")

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
