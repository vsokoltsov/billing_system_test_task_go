package entities

import "github.com/shopspring/decimal"

// Wallet represents internal information about users' wallet structure
type Wallet struct {
	ID       int
	UserID   int
	Balance  decimal.Decimal
	Currency string
}
