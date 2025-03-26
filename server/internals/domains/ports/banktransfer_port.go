package domains

import (
	"context"

	adapters "github.com/cybrarymin/gRPC/server/internals/adapters/driven_adapters/database"
	domains "github.com/cybrarymin/gRPC/server/internals/domains/entities"
	"github.com/google/uuid"
)

type BankTransferRepositoryPort interface {
	CreateTransfer(pCtx context.Context, ntransfer *domains.BankTransfer) (*adapters.BankTransferModel, error)
	UpdateTransfer(pCtx context.Context, transferUUID uuid.UUID, ntransfer *domains.BankTransfer) (*adapters.BankTransferModel, error)
	GetTransferByID(pCtx context.Context, transferUUID uuid.UUID) (*adapters.BankTransferModel, error)
}

type BankTransferGrpcPort interface {
	TransferMoney(ctx context.Context, srcAccount uuid.UUID, dstAccount uuid.UUID, currency string, amount float64) (*domains.BankTransfer, error)
}
