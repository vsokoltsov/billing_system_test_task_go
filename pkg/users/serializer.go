package users

import "github.com/shopspring/decimal"

type UserSerializer struct {
	ID       int             `json:"id"`
	Email    string          `json:"email"`
	Balance  decimal.Decimal `json:"balance"`
	Currency string          `json:"currency"`
}
