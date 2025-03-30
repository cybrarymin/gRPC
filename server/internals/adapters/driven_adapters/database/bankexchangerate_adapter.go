package adapters

import (
	"context"
	"database/sql"
	"time"

	domainsErrors "github.com/cybrarymin/gRPC/server/internals/domains/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

type BankExchangeRateRepository struct {
	db     *bun.DB
	logger *zerolog.Logger
}

func NewBankExchangeRateRepository(db *bun.DB, logger *zerolog.Logger) *BankExchangeRateRepository {
	return &BankExchangeRateRepository{
		db:     db,
		logger: logger,
	}
}

func (ad *BankExchangeRateRepository) GetAll(pCtx context.Context) (ExchangeRatesModel, error) {
	exchList := &ExchangeRatesModel{}
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	count, err := ad.db.NewSelect().Model(exchList).ScanAndCount(ctx)
	if err != nil {
		ad.logger.Error().Err(err).Msg("failed to get all exchange rates")
		return nil, domainsErrors.DatabaseError(err, "get all exchange rates")
	}

	if count == 0 {
		ad.logger.Debug().Msg("no exchange rates found")
		return *exchList, nil // Return empty list, not an error
	}

	return *exchList, nil
}

func (ad *BankExchangeRateRepository) GetByID(pCtx context.Context, exchUUID uuid.UUID) (*ExchangeRateModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	nEx := &ExchangeRateModel{}
	err := ad.db.NewSelect().Model(nEx).Where("exchange_rate_uuid = ?", exchUUID).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			ad.logger.Debug().
				Str("exchange_rate_uuid", exchUUID.String()).
				Msg("exchange rate not found")
			return nil, domainsErrors.NotFoundError("exchange rate", exchUUID.String())
		}

		ad.logger.Error().Err(err).
			Str("exchange_rate_uuid", exchUUID.String()).
			Msg("failed to get exchange rate")
		return nil, domainsErrors.DatabaseError(err, "get exchange rate by ID")
	}

	return nEx, nil
}

func (ad *BankExchangeRateRepository) GetByCurrencies(pCtx context.Context, FromCurrency string, ToCurrency string) (*ExchangeRateModel, error) {
	if FromCurrency == ToCurrency {
		return &ExchangeRateModel{
			FromCurrency: FromCurrency,
			ToCurrency:   ToCurrency,
			Rate:         1,
		}, nil
	}

	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	nEx := &ExchangeRateModel{}
	err := ad.db.NewSelect().Model(nEx).Where("from_currency = ? and to_currency = ?", FromCurrency, ToCurrency).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			ad.logger.Debug().
				Str("from_currency", FromCurrency).
				Str("to_currency", ToCurrency).
				Msg("exchange rate not found for currency pair")
			return nil, domainsErrors.NotFoundError("exchange rate", FromCurrency+"/"+ToCurrency)
		}

		ad.logger.Error().Err(err).
			Str("from_currency", FromCurrency).
			Str("to_currency", ToCurrency).
			Msg("failed to get exchange rate for currency pair")
		return nil, domainsErrors.DatabaseError(err, "get exchange rate by currencies")
	}

	return nEx, nil
}

func (ad *BankExchangeRateRepository) Update(pCtx context.Context, exchUUID uuid.UUID, nExchangeRate *ExchangeRateModel) (*ExchangeRateModel, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	nExchangeRate.UpdatedAt = time.Now()

	result, err := ad.db.NewUpdate().
		Model(nExchangeRate).
		Where("exchange_rate_uuid = ? and updated_at < ? ", exchUUID, nExchangeRate.UpdatedAt).
		Returning("*").
		Exec(ctx)

	if err != nil {
		ad.logger.Error().Err(err).
			Str("exchange_rate_uuid", exchUUID.String()).
			Msg("failed to update exchange rate")
		return nil, domainsErrors.DatabaseError(err, "update exchange rate")
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ad.logger.Error().Err(err).
			Str("exchange_rate_uuid", exchUUID.String()).
			Msg("failed to get rows affected after update operation")
		return nil, domainsErrors.DatabaseError(err, "check update result")
	}

	// If no rows were affected, either the exchange rate wasn't found or it was concurrently modified
	if rowsAffected == 0 {
		// Check if the exchange rate exists
		_, err := ad.GetByID(pCtx, exchUUID)
		if err != nil {
			// Propagate the error (which should be NotFoundError if exchange rate doesn't exist)
			return nil, err
		}

		// If we got here, the exchange rate exists but was concurrently modified
		ad.logger.Warn().
			Str("exchange_rate_uuid", exchUUID.String()).
			Time("updated_at", nExchangeRate.UpdatedAt).
			Msg("concurrent modification detected on exchange rate")
		return nil, domainsErrors.ConcurrentModificationError("exchange rate", exchUUID.String())
	}

	return nExchangeRate, nil
}
