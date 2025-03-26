package domains

import (
	"time"

	"github.com/google/uuid"
)

type BankTransfers []BankTransfer

type BankTransfer struct {
	TransferUUID      uuid.UUID
	FromAccountUUID   uuid.UUID
	ToAccountUUID     uuid.UUID
	Currency          string
	Amount            float64
	TransferTimestamp time.Time
	TransferSucceed   bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
