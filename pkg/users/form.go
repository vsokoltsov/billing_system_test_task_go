package users

import (
	"billing_system_test_task/pkg/utils"

	"github.com/shopspring/decimal"
)

type UserForm struct {
	Email string `json:"email" validate:"required,email"`
}

func (uf *UserForm) Submit() *map[string][]string {
	var (
		errors = utils.ValidateForm(uf, make(map[string][]string))
	)

	// Perform validations by tags
	if len(errors) > 0 {
		return &errors
	}

	return nil
}

type EnrollForm struct {
	Amount decimal.Decimal `json:"amount" validate:"required,gt=0"`
}

func (ef *EnrollForm) Submit() *map[string][]string {
	errors := make(map[string][]string)
	if !ef.Amount.IsPositive() {
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
