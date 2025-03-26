package adapters

import (
	"time"

	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type BankTransactionsModel []BankTransactionModel

type BankTransactionModel struct {
	bun.BaseModel        `bun:"table:bank_transactions"`
	TransactionUUID      uuid.UUID         `bun:",pk,type:uuid,nullzero,notnull"`
	BankAccount          *BankAccountModel `bun:"rel:belongs-to,join:account_uuid=account_uuid"`
	AccountUUID          uuid.UUID
	TransactionTimestamp time.Time `bun:",type:timestamptz,notnull"`
	Amount               float64   `bun:",type:numeric(15,2),notnull"`
	TransactionType      string    `bun:",type:varchar(25),notnull"`
	Notes                string    `bun:",type:text"`
	CreatedAt            time.Time `bun:",type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt            time.Time `bun:",type:timestamptz,nullzero,notnull"`
}

type BankAccountsModel []BankAccountModel

type BankAccountModel struct {
	bun.BaseModel   `bun:"table:bank_accounts"`
	AccountUUID     uuid.UUID               `bun:",pk,type:uuid,nullzero,notnull"`
	AccountNumber   string                  `bun:",type:varchar(20),unique,notnull"`
	AccountName     string                  `bun:",type:varchar(100),notnull"`
	Currency        string                  `bun:",type:varchar(5),notnull"`
	CurrentBalance  float64                 `bun:",type:numeric(15,2),notnull"`
	CreatedAt       time.Time               `bun:",type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt       time.Time               `bun:",type:timestsamptz,nullzero,notnull"`
	BankTransaction []*BankTransactionModel `bun:"rel:has-many,join:account_uuid=account_uuid"`
	BankTransfer    []*BankTransferModel    `bun:"rel:has-many,join:account_uuid=from_account_uuid"`
}

type ExchangeRatesModel []ExchangeRateModel

type ExchangeRateModel struct {
	bun.BaseModel      `bun:"table:bank_exchange_rates"`
	ExchangeRateUUID   uuid.UUID `bun:",type:uuid,unique,notnull"`
	FromCurrency       string    `bun:",type:varchar(5),notnull"`
	ToCurrency         string    `bun:",type:varchar(5),notnull"`
	Rate               float64   `bun:",type:numeric(20,5),notnull,nullzero"`
	ValidFromTimestamp time.Time `bun:",type:timestamptz,nullzero,notnull"`
	ValidToTimestamp   time.Time `bun:",type:timestamptz,nullzero,notnull"`
	CreatedAt          time.Time `bun:",type:timestamptz,nullzero,notnull"`
	UpdatedAt          time.Time `bun:",type:timestamptz,nullzero,notnull"`
}

type BankTransfersModel []BankTransferModel

type BankTransferModel struct {
	bun.BaseModel     `bun:"table:bank_transfers"`
	TransferUUID      uuid.UUID         `bun:",type:uuid,notnull,nullzero,unique"`
	FromBankAccount   *BankAccountModel `bun:"rel:belongs-to,join:from_account_uuid=account_uuid"`
	ToBankAccount     *BankAccountModel `bun:"rel:belongs-to,join:to_account_uuid=account_uuid"`
	FromAccountUUID   uuid.UUID         `bun:",type:uuid,notnull"`
	ToAccountUUID     uuid.UUID         `bun:",type:uuid,notnull"`
	Currency          string            `bun:",type:varchar(20),notnull"`
	Amount            float64           `bun:",type:numeric(15,2),notnull,unique"`
	TransferTimestamp time.Time         `bun:",type:timestamptz,notnull,nullzero"`
	TransferSucceed   bool              `bun:",type:boolean,notnull"`
	CreatedAt         time.Time         `bun:",type:timestamptz,notnull,nullzero"`
	UpdatedAt         time.Time         `bun:",type:timestamptz,notnull,nullzero"`
}

func NewBankAccountModel(ba *domains.BankAccount) *BankAccountModel {
	return &BankAccountModel{
		AccountUUID:    ba.AccountUUID,
		AccountNumber:  ba.AccountNumber,
		AccountName:    ba.AccountName,
		Currency:       ba.Currency,
		CurrentBalance: ba.CurrentBalance,
		UpdatedAt:      ba.UpdatedAt,
	}
}

func NewBankTransactionModel(bt *domains.BankTransaction) *BankTransactionModel {
	return &BankTransactionModel{
		TransactionUUID:      bt.TransactionUUID,
		AccountUUID:          bt.AccountUUID,
		Amount:               bt.Amount,
		TransactionTimestamp: bt.TransactionTimestamp,
		TransactionType:      bt.TransactionType,
		Notes:                bt.Notes,
		CreatedAt:            bt.CreatedAt,
		UpdatedAt:            bt.UpdatedAt,
	}
}

func NewExchangeRateModel(srcCurrency string, dstCurrency string, rate float64) *ExchangeRateModel {
	startTime := time.Now()
	return &ExchangeRateModel{
		FromCurrency:       srcCurrency,
		ToCurrency:         dstCurrency,
		Rate:               rate,
		ValidFromTimestamp: startTime,
		ValidToTimestamp:   startTime.Add(time.Minute * 1),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func NewTransferModel(nt *domains.BankTransfer) *BankTransferModel {
	return &BankTransferModel{
		TransferUUID:      nt.TransferUUID,
		FromAccountUUID:   nt.FromAccountUUID,
		ToAccountUUID:     nt.ToAccountUUID,
		Currency:          nt.Currency,
		Amount:            nt.Amount,
		TransferTimestamp: nt.TransferTimestamp,
		CreatedAt:         nt.CreatedAt,
		UpdatedAt:         nt.UpdatedAt,
		TransferSucceed:   nt.TransferSucceed,
	}
}
