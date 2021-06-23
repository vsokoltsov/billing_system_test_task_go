package operations

import (
	"time"

	"github.com/shopspring/decimal"
)

type WalletOperation struct {
	ID         int
	Operation  string
	WalletFrom int
	WalletTo   int
	Amount     decimal.Decimal
	CreatedAt  time.Time
}
