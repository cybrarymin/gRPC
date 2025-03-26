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

type BankAccountRepository struct {
	db     *bun.DB
	logger *zerolog.Logger
}

func NewBankAccountRepository(db *bun.DB, logger *zerolog.Logger) *BankAccountRepository {
	return &BankAccountRepository{
		db:     db,
		logger: logger,
	}
}

func (br *BankAccountRepository) Create(pCtx context.Context, ba *domains.BankAccount) (*BankAccountModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	nBankAccountModel := NewBankAccountModel(ba)
	_, err := br.db.NewInsert().Model(nBankAccountModel).Exec(ctx, nBankAccountModel)
	if err != nil {
		br.logger.Error().Err(err).
			Str("account_uuid", ba.AccountUUID.String()).
			Str("account_number", ba.AccountNumber).
			Msg("failed to create bank account")
		return nil, domainsErrors.DatabaseError(err, "create bank account")
	}
	return nBankAccountModel, nil
}

func (br *BankAccountRepository) DeleteByID(pCtx context.Context, accID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	result, err := br.db.NewDelete().Model((*BankAccountModel)(nil)).Where("account_uuid = ?", accID).Exec(ctx)
	if err != nil {
		br.logger.Error().Err(err).
			Str("account_uuid", accID.String()).
			Msg("failed to delete bank account")
		return domainsErrors.DatabaseError(err, "delete bank account")
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		br.logger.Error().Err(err).
			Str("account_uuid", accID.String()).
			Msg("failed to get rows affected after delete operation")
		return domainsErrors.DatabaseError(err, "check delete result")
	}

	// If no rows were affected, the account wasn't found
	if rowsAffected == 0 {
		br.logger.Warn().
			Str("account_uuid", accID.String()).
			Msg("no bank account found to delete")
		return domainsErrors.NotFoundError("bank account", accID.String())
	}

	return nil
}

func (br *BankAccountRepository) GetByID(pCtx context.Context, accID uuid.UUID) (*BankAccountModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	nAccount := &BankAccountModel{}
	err := br.db.NewSelect().Model(nAccount).Where("account_uuid = ?", accID).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			br.logger.Debug().
				Str("account_uuid", accID.String()).
				Msg("bank account not found")
			return nil, domainsErrors.NotFoundError("bank account", accID.String())
		}

		br.logger.Error().Err(err).
			Str("account_uuid", accID.String()).
			Msg("failed to get bank account")
		return nil, domainsErrors.DatabaseError(err, "get bank account")
	}
	return nAccount, nil
}

func (br *BankAccountRepository) Update(pCtx context.Context, accUUID uuid.UUID, nAccount *domains.BankAccount) (*BankAccountModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	nAccount.UpdatedAt = time.Now()
	nBankAccountModel := NewBankAccountModel(nAccount)

	result, err := br.db.NewUpdate().
		Model(nBankAccountModel).
		Where("account_uuid = ? AND updated_at < ?", accUUID, nAccount.UpdatedAt).
		Returning("*").
		Exec(ctx)

	if err != nil {
		br.logger.Error().Err(err).
			Str("account_uuid", accUUID.String()).
			Msg("failed to update the bank account information")
		return nil, domainsErrors.DatabaseError(err, "update bank account")
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		br.logger.Error().Err(err).
			Str("account_uuid", accUUID.String()).
			Msg("failed to get rows affected after update operation")
		return nil, domainsErrors.DatabaseError(err, "check update result")
	}

	// If no rows were affected, either the account wasn't found or it was concurrently modified
	if rowsAffected == 0 {
		// Check if the account exists
		_, err := br.GetByID(pCtx, accUUID)
		if err != nil {
			// Propagate the error (which should be NotFoundError if account doesn't exist)
			return nil, err
		}

		// If we got here, the account exists but was concurrently modified
		br.logger.Warn().
			Str("account_uuid", accUUID.String()).
			Time("updated_at", nAccount.UpdatedAt).
			Msg("concurrent modification detected on bank account")
		return nil, domainsErrors.ConcurrentModificationError("bank account", accUUID.String())
	}

	return nBankAccountModel, nil
}
