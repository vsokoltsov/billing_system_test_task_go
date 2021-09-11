package users

import "github.com/shopspring/decimal"

// UserSerializer serializes user information
type UserSerializer struct {
	ID       int             `json:"id"`
	Email    string          `json:"email"`
	Balance  decimal.Decimal `json:"balance"`
	Currency string          `json:"currency"`
}
