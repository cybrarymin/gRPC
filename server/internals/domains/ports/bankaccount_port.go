package domains

import (
	"context"

	adapters "github.com/cybrarymin/gRPC/server/internals/adapters/driven_adapters/database"
	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	"github.com/google/uuid"
)

type BankAccountRepositoryPort interface {
	Create(context.Context, *domains.BankAccount) (*adapters.BankAccountModel, error)
	DeleteByID(context.Context, uuid.UUID) error
	Update(context.Context, uuid.UUID, *domains.BankAccount) (*adapters.BankAccountModel, error)
	GetByID(context.Context, uuid.UUID) (*adapters.BankAccountModel, error)
}

type BankAccountGrpcPort interface {
	OpenAccount(ctx context.Context, accName string, accNum string, currency string, balance float64) (*domains.BankAccount, error)
	GetCurrentBalance(ctx context.Context, accUUID uuid.UUID) (float64, string, error)
}
