package forms

import (
	"billing_system_test_task/internal/utils"

	"github.com/shopspring/decimal"
)

// WalletForm stores fields for form validation
type WalletForm struct {
	WalletFrom int             `json:"wallet_from" validate:"required"`
	WalletTo   int             `json:"wallet_to" validate:"required"`
	Amount     decimal.Decimal `json:"amount" validate:"required,gt=0"`
}

// Submit validates form attributes
func (wf *WalletForm) Submit() *map[string][]string {
	var (
		errors = utils.ValidateForm(wf, make(map[string][]string))
	)
	if !wf.Amount.IsPositive() {
		errors["amount"] = []string{
			"less than a zero",
		}
	}

	if wf.WalletFrom == wf.WalletTo {
		errors["wallet_from"] = []string{
			"source wallet is equal to destination wallet",
		}
	}

	// Perform validations by tags
	if len(errors) > 0 {
		return &errors
	}

	return nil
}
