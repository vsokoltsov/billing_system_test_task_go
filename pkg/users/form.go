package users

import "billing_system_test_task/pkg/utils"

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
