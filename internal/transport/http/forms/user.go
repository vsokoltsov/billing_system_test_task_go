package forms

import (
	"billing_system_test_task/internal/utils"

	"github.com/shopspring/decimal"
)

// UserForm represents user form for parameters validation
type UserForm struct {
	Email string `json:"email" validate:"required,email"`
}

// Submit validates given parameter for user
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

// EnrollForm represents form for wallet's enroll
type EnrollForm struct {
	Amount decimal.Decimal `json:"amount" validate:"required,gt=0"`
}

// Submit validates given parameter for wallet's enroll
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
