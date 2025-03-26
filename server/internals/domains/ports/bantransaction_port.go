package domains

import (
	"context"

	adapters "github.com/cybrarymin/gRPC/server/internals/adapters/driven_adapters/database"
	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	"github.com/google/uuid"
)

type TransactionRepositoryPort interface {
	CreateTransaction(context.Context, *domains.BankTransaction) (*adapters.BankTransactionModel, error)
	GetTransactionByID(context.Context, uuid.UUID) (*adapters.BankTransactionModel, error)
	GetTransactionsByAccount(context.Context, uuid.UUID) (adapters.BankTransactionsModel, error)
}

type TransactionGrpcPort interface {
	NewTransaction(ctx context.Context, accUUID uuid.UUID, amount float64, TRType string, note string) (*domains.BankTransaction, error)
}
