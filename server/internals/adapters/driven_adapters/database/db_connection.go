package adapters

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bunzerolog"
)

type DbConfig struct {
	DBMaxConnCount       int
	DBMaxIdleConnCount   int
	DBMaxIdleConnTimeout time.Duration
	DatabaseDSN          string
	Logger               *zerolog.Logger
}

func NewBunDB(ctx context.Context, cfg *DbConfig) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DatabaseDSN)))
	db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())
	db.AddQueryHook(bunzerolog.NewQueryHook(
		bunzerolog.WithLogger(cfg.Logger),
		bunzerolog.WithQueryLogLevel(zerolog.DebugLevel),      // Show database interaction logs by debug tag
		bunzerolog.WithSlowQueryLogLevel(zerolog.WarnLevel),   // Show database slow queries as warnings tag
		bunzerolog.WithErrorQueryLogLevel(zerolog.DebugLevel), // Show database queries errors as error tag
		bunzerolog.WithSlowQueryThreshold(3*time.Second),
	))
	db.SetMaxOpenConns(cfg.DBMaxConnCount)
	db.SetMaxIdleConns(cfg.DBMaxIdleConnCount)
	db.SetConnMaxIdleTime(cfg.DBMaxIdleConnTimeout)

	err := db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
