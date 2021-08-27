package operations

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

//easyjson:json
type WalletOperation struct {
	ID         int             `json:"id"`
	Operation  string          `json:"operation"`
	WalletFrom sql.NullInt32   `json:"wallet_from"`
	WalletTo   int             `json:"wallet_to"`
	Amount     decimal.Decimal `json:"amount"`
	CreatedAt  time.Time       `json:"created_at"`
}
