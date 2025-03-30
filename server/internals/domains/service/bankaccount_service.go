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

type BankAccountService struct {
	port      ports.BankAccountRepositoryPort
	logger    *zerolog.Logger
	validator *Validator
}

func NewBankAccountService(repoPort ports.BankAccountRepositoryPort, logger *zerolog.Logger, validator *Validator) *BankAccountService {
	return &BankAccountService{
		port:      repoPort,
		logger:    logger,
		validator: validator,
	}
}

func (s *BankAccountService) OpenAccount(ctx context.Context, accName string, accNumber string, currency string, balance float64) (*domains.BankAccount, error) {
	sCtx, nSpan := otel.Tracer("OpenAccount").Start(ctx, "OpenAccount.service.span")
	defer nSpan.End()

	nAccount := &domains.BankAccount{
		AccountNumber:  accNumber,
		AccountName:    accName,
		Currency:       currency,
		CurrentBalance: balance,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	s.logger.Info().
		Str("account_name", nAccount.AccountName).
		Str("account_number", nAccount.AccountNumber).
		Msg("staring bank account creation process....")

	createdAccount, err := s.port.Create(sCtx, nAccount)
	if err != nil {
		s.logger.Error().Err(err).
			Str("account_id", nAccount.AccountUUID.String()).
			Str("account_number", nAccount.AccountNumber).
			Str("account_name", nAccount.AccountName).
			Msg("failed to open a new bank account")
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to create account")
		return nil, err
	}

	s.logger.Info().
		Msg("ending bank account creation process....")
	nAccount.AccountUUID = createdAccount.AccountUUID
	return nAccount, nil
}

func (s *BankAccountService) GetCurrentBalance(ctx context.Context, accUUID uuid.UUID) (float64, string, error) {
	sCtx, nSpan := otel.Tracer("GetCurrentBalance").Start(ctx, "GetCurrentBalance.service.span")
	defer nSpan.End()

	s.logger.Info().
		Str("account_uuid", accUUID.String()).
		Msg("fetching account uuid information to get its current balance...")

	bankAccountModel, err := s.port.GetByID(sCtx, accUUID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("account_uuid", accUUID.String()).
			Msg("couldn't get requested account information")
		nSpan.RecordError(err)
		nSpan.SetStatus(codes.Error, "failed to retrieve user")
		return 0, "", err
	}

	s.logger.Info().
		Str("account_uuid", accUUID.String()).
		Float64("balance", bankAccountModel.CurrentBalance).
		Msg("finished getting account current balance...")

	return bankAccountModel.CurrentBalance, bankAccountModel.Currency, nil
}
