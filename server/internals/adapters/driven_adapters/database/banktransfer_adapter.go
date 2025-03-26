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

type BankTransferRepository struct {
	db     *bun.DB
	logger *zerolog.Logger
}

func NewBankTransferRepository(db *bun.DB, logger *zerolog.Logger) *BankTransferRepository {
	return &BankTransferRepository{
		db:     db,
		logger: logger,
	}
}

func (ad *BankTransferRepository) CreateTransfer(pCtx context.Context, ntransfer *domains.BankTransfer) (*BankTransferModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	ntransferModel := NewTransferModel(ntransfer)
	_, err := ad.db.NewInsert().Model(ntransferModel).Exec(ctx, ntransferModel)
	if err != nil {
		ad.logger.Error().Err(err).
			Str("src_account", ntransfer.FromAccountUUID.String()).
			Str("to_account", ntransfer.ToAccountUUID.String()).
			Float64("amount", ntransfer.Amount).
			Time("transfer_timestamp", ntransfer.TransferTimestamp).
			Msg("failed creating new transfer object in database")

		return nil, domainsErrors.DatabaseError(err, "create bank transfer")
	}
	return ntransferModel, nil
}

func (ad *BankTransferRepository) GetTransferByID(pCtx context.Context, transferUUID uuid.UUID) (*BankTransferModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	transfer := &BankTransferModel{}
	err := ad.db.NewSelect().Model(transfer).Where("transfer_uuid = ?", transferUUID).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			ad.logger.Debug().
				Str("transfer_uuid", transferUUID.String()).
				Msg("bank transfer not found")
			return nil, domainsErrors.NotFoundError("bank transfer", transferUUID.String())
		}

		ad.logger.Error().Err(err).
			Str("transfer_uuid", transferUUID.String()).
			Msg("failed to get bank transfer")
		return nil, domainsErrors.DatabaseError(err, "get bank transfer")
	}

	return transfer, nil
}

func (ad *BankTransferRepository) UpdateTransfer(pCtx context.Context, transferUUID uuid.UUID, ntransfer *domains.BankTransfer) (*BankTransferModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	ntransferModel := NewTransferModel(ntransfer)
	ntransferModel.UpdatedAt = time.Now()

	result, err := ad.db.NewUpdate().
		Model(ntransferModel).
		Where("transfer_uuid = ? AND updated_at < ?", transferUUID, ntransferModel.UpdatedAt).
		Returning("*").
		Exec(ctx)

	if err != nil {
		ad.logger.Error().Err(err).
			Str("src_account", ntransfer.FromAccountUUID.String()).
			Str("to_account", ntransfer.ToAccountUUID.String()).
			Float64("amount", ntransfer.Amount).
			Time("transfer_timestamp", ntransfer.TransferTimestamp).
			Msg("failed updating transfer object in database")

		return nil, domainsErrors.DatabaseError(err, "update bank transfer")
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ad.logger.Error().Err(err).
			Str("transfer_uuid", transferUUID.String()).
			Msg("failed to get rows affected after update operation")
		return nil, domainsErrors.DatabaseError(err, "check update result")
	}

	// If no rows were affected, either the transfer wasn't found or it was concurrently modified
	if rowsAffected == 0 {
		// Check if the transfer exists
		_, err := ad.GetTransferByID(pCtx, transferUUID)
		if err != nil {
			// Propagate the error (which should be NotFoundError if transfer doesn't exist)
			return nil, err
		}

		// If we got here, the transfer exists but was concurrently modified
		ad.logger.Warn().
			Str("transfer_uuid", transferUUID.String()).
			Time("updated_at", ntransfer.UpdatedAt).
			Msg("concurrent modification detected on bank transfer")
		return nil, domainsErrors.ConcurrentModificationError("bank transfer", transferUUID.String())
	}

	return ntransferModel, nil
}
