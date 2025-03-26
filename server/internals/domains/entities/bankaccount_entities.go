package domains

import (
	"time"

	"github.com/google/uuid"
)

type BankAccounts []BankAccount

type BankAccount struct {
	AccountUUID    uuid.UUID
	AccountNumber  string
	AccountName    string
	Currency       string
	CurrentBalance float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
