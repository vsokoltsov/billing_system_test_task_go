package wallets

import (
	"billing_system_test_task/pkg/utils"

	"github.com/shopspring/decimal"
)

type WalletForm struct {
	WalletFrom int             `json:"wallet_from" validate:"required"`
	WalletTo   int             `json:"wallet_to" validate:"required"`
	Amount     decimal.Decimal `json:"amount" validate:"required,gt=0"`
}

func (wf *WalletForm) Submit() *map[string][]string {
	var (
		errors = utils.ValidateForm(wf, make(map[string][]string))
	)
	if !wf.Amount.IsPositive() {
		errors["amount"] = []string{
			"less than a zero",
		}
	}

	// Perform validations by tags
	if len(errors) > 0 {
		return &errors
	}

	return nil
}
