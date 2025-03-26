package domains

import (
	"time"

	"github.com/google/uuid"
)

const (
	TRRefundType   = "Refund"
	TRPaymentType  = "Payment"
	TRTransferType = "Transfer"
	TRDepositType  = "Deposit"
	TRWithDrawType = "Withdraw"
)

type BankTransactions []BankTransaction

type BankTransaction struct {
	TransactionUUID      uuid.UUID
	AccountUUID          uuid.UUID
	TransactionTimestamp time.Time
	Amount               float64
	TransactionType      string
	Notes                string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
