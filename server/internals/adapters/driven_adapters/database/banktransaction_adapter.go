package adapters

import (
	"context"
	"database/sql"
	"time"

	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	domainsErrors "github.com/cybrarymin/gRPC/server/internals/domains/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

type BankTransactionRepository struct {
	db     *bun.DB
	logger *zerolog.Logger
}

func NewBankTransactionRepository(db *bun.DB, logger *zerolog.Logger) *BankTransactionRepository {
	return &BankTransactionRepository{
		db:     db,
		logger: logger,
	}
}

func (br *BankTransactionRepository) CreateTransaction(pCtx context.Context, bt *domains.BankTransaction) (*BankTransactionModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	nTransactionModel := NewBankTransactionModel(bt)

	_, err := br.db.NewInsert().Model(nTransactionModel).Exec(ctx, nTransactionModel)
	if err != nil {
		br.logger.Error().Err(err).
			Str("transaction_uuid", bt.TransactionUUID.String()).
			Str("account_uuid", bt.AccountUUID.String()).
			Float64("amount", bt.Amount).
			Str("transaction_type", bt.TransactionType).
			Msg("failed to create transaction")
		return nil, domainsErrors.DatabaseError(err, "create bank transaction")
	}
	return nTransactionModel, nil
}

func (br *BankTransactionRepository) GetTransactionByID(pCtx context.Context, transactionUUID uuid.UUID) (*BankTransactionModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	transaction := &BankTransactionModel{}
	err := br.db.NewSelect().Model(transaction).Where("transaction_uuid = ?", transactionUUID).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			br.logger.Debug().
				Str("transaction_uuid", transactionUUID.String()).
				Msg("transaction not found")
			return nil, domainsErrors.NotFoundError("bank transaction", transactionUUID.String())
		}

		br.logger.Error().Err(err).
			Str("transaction_uuid", transactionUUID.String()).
			Msg("failed to get transaction")
		return nil, domainsErrors.DatabaseError(err, "get bank transaction")
	}

	return transaction, nil
}

func (br *BankTransactionRepository) GetTransactionsByAccount(pCtx context.Context, accountUUID uuid.UUID) (BankTransactionsModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	transactions := make(BankTransactionsModel, 0)
	err := br.db.NewSelect().Model(&transactions).Where("account_uuid = ?", accountUUID).Scan(ctx)
	if err != nil {
		br.logger.Error().Err(err).
			Str("account_uuid", accountUUID.String()).
			Msg("failed to get transactions for account")
		return nil, domainsErrors.DatabaseError(err, "get transactions by account")
	}

	if len(transactions) == 0 {
		br.logger.Debug().
			Str("account_uuid", accountUUID.String()).
			Msg("no transactions found for account")
	}

	return transactions, nil
}
