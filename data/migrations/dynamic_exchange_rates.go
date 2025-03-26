package data

import (
	"context"
	"math/rand"
	"time"

	adapters "github.com/cybrarymin/gRPC/server/internals/adapters/driven_adapters/database"
	"github.com/rs/zerolog"
)

type DynamicExchangeRate struct {
	ad     *adapters.BankExchangeRateRepository
	logger *zerolog.Logger
}

func NewDynamicExchangeRate(ad *adapters.BankExchangeRateRepository, logger *zerolog.Logger) *DynamicExchangeRate {
	return &DynamicExchangeRate{
		ad:     ad,
		logger: logger,
	}
}

func (d *DynamicExchangeRate) ChangeExchangeRates(ctx context.Context) error {
	d.logger.Info().Msg("started dynamic exchange rate sampler...")
	for {
		exRates, err := d.ad.GetAll(ctx)
		if err != nil {
			return err
		}

		for _, exRate := range exRates {
			fluc := -2 + rand.Float64()*4
			exRate.Rate += fluc
			startTime := time.Now()
			exRate.ValidFromTimestamp = startTime
			exRate.ValidToTimestamp = startTime.Add(time.Second * 30)
			_, err := d.ad.Update(ctx, exRate.ExchangeRateUUID, &exRate)
			if err != nil {
				return err
			}
			d.logger.Debug().
				Str("exchange_rate_uuid", exRate.ExchangeRateUUID.String()).
				Str("from_currency", exRate.FromCurrency).
				Str("to_currency", exRate.ToCurrency).
				Float64("new_rate", exRate.Rate).
				Time("rate_validity_period", exRate.ValidToTimestamp)
		}
		time.Sleep(30 * time.Second)
	}
}
