package wallets

import (
	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID       int
	UserID   int
	Balance  decimal.Decimal
	Currency string
}
