package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	client_adapters "github.com/cybrarymin/gRPC/client/adapters"
	client_services "github.com/cybrarymin/gRPC/client/domains/services"
	data "github.com/cybrarymin/gRPC/data/migrations"
	dbadapters "github.com/cybrarymin/gRPC/server/internals/adapters/driven_adapters/database"
	adapters "github.com/cybrarymin/gRPC/server/internals/adapters/driving_adapters/grpc"
	domains "github.com/cybrarymin/gRPC/server/internals/domains/service"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	FlagLogLevel             string
	FlagDBMaxConnCount       int
	FlagDBMaxIdleConnCount   int
	FlagDBMaxIdleConnTimeout time.Duration
	FlagDBDSN                string
)

func main() {
	ctx := context.Background()
	var logger zerolog.Logger
	loglvl, err := zerolog.ParseLevel(FlagLogLevel)
	if err != nil {
		log.Panicln("couldn't parse the loglevel") // TODO
	}

	if FlagLogLevel == zerolog.LevelTraceValue {
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Stack().Logger().Level(loglvl)
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(loglvl)
	}

	// Create new bankaccountRepository
	dbCfg := dbadapters.DbConfig{
		DBMaxConnCount:       FlagDBMaxConnCount,
		DBMaxIdleConnCount:   FlagDBMaxIdleConnCount,
		DBMaxIdleConnTimeout: FlagDBMaxIdleConnTimeout,
		DatabaseDSN:          "postgres://postgres:m.amin24242@localhost:5432/bank?sslmode=disable",
		Logger:               &logger,
	}

	db, err := dbadapters.NewBunDB(ctx, &dbCfg)
	if err != nil {
		logger.Panic().Msgf("couldn't establish database connection: %s", err.Error())
	}
	// Create new validator
	v := domains.NewValidator()

	// Create new repository for invoking the CRUD operations on our backend database
	postgresBankAccountRepo := dbadapters.NewBankAccountRepository(db, &logger)
	postgresTransactionRepo := dbadapters.NewBankTransactionRepository(db, &logger)
	postgresExchangeRateRepo := dbadapters.NewBankExchangeRateRepository(db, &logger)
	postgresTransferRepo := dbadapters.NewBankTransferRepository(db, &logger)

	// Create new domain bank account service. This domain service is the type of BankAccountGrpcPort so we will give it to GRPC adapter
	domainBankAccountService := domains.NewBankAccountService(postgresBankAccountRepo, &logger, &v)
	domainTransactionService := domains.NewTransactionService(postgresTransactionRepo, postgresBankAccountRepo, &logger)
	domainExchangeRateService := domains.NewBankExchangeRateService(postgresExchangeRateRepo, &logger)
	domainTransferService := domains.NewBankTransferService(postgresTransferRepo, postgresBankAccountRepo, postgresTransactionRepo, postgresExchangeRateRepo, &logger, &v)

	// Create new grp
	grpcAdapter := adapters.NewGrpcAdapter("0.0.0.0", "9090", &logger, adapters.GrpcPortReference{
		domainBankAccountService,
		domainTransactionService,
		domainExchangeRateService,
		domainTransferService,
		&v,
	})

	// Use dynamic exchange rate updater as a dummy data sampler
	dRateChanger := data.NewDynamicExchangeRate(postgresExchangeRateRepo, &logger)
	BackgroundJob(func() {
		dRateChanger.ChangeExchangeRates(ctx)
	}, "dynamic exchange rate changer paniced", &logger)

	grpcAdapter.Run()
	defer grpcAdapter.Stop()

}

func BackgroundJob(nfunc func(), PanicErrMsg string, logger *zerolog.Logger) {
	go func() {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				pErr := errors.New(fmt.Sprintln(panicErr))
				logger.Error().Stack().Err(pErr).Msg(PanicErrMsg)
			}
		}()
		nfunc()
	}()
}

func client(ctx context.Context) (*client_services.BankCliService, error) {

	var logger zerolog.Logger
	loglvl, err := zerolog.ParseLevel(FlagLogLevel)
	if err != nil {
		logger.Error().Err(err).Msg("unsupported log level type")
		return nil, err
	}

	if FlagLogLevel == zerolog.LevelTraceValue {
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Stack().Logger().Level(loglvl)
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(loglvl)
	}

	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error().Err(err).Msg("couldn't establish connection to the grpc server")
		return nil, err
	}

	// create new client adapter
	grpcAdapter, err := client_adapters.NewBankGrpcClientAdapter(conn, &logger)
	if err != nil {
		logger.Error().Err(err).Msg("couldn't initialize the grpc adapter")
		return nil, err
	}

	// create new client service
	cli_service := client_services.NewBankCliService(grpcAdapter, &logger)

	return cli_service, nil

}
