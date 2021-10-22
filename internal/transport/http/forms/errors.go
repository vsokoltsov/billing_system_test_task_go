package forms

import (
	"strings"

	"github.com/go-playground/validator"
)

// ValidateForm performs base form validation
func ValidateForm(form interface{}, formErrors map[string][]string) map[string][]string {
	validate := validator.New()
	errValidate := validate.Struct(form)
	if errValidate != nil {
		for _, err := range errValidate.(validator.ValidationErrors) {
			errKey := strings.ToLower(err.Field())
			_, ok := formErrors[errKey]
			if ok {
				formErrors[errKey] = append(formErrors[errKey], errorTagMessage(err.Tag()))
			} else {
				formErrors[errKey] = []string{errorTagMessage(err.Tag())}
			}
		}
	}
	return formErrors
}

func errorTagMessage(tag string) string {
	var result string
	switch tag {
	case "required":
		result = "Field required"
	case "email":
		result = "Invalid email format"
	}
	return result
}
