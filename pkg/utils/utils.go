package utils

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
)

type FormErrorSerializer struct {
	Messages map[string][]string `json:"messages"`
}

type ErrorMsg struct {
	Message string `json:"message"`
}

func JsonResponseError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorMsg{Message: message})
}

// ValidateForm performs base form validation
func ValidateForm(form interface{}, formErrors map[string][]string) map[string][]string {
	validate := validator.New()
	errValidate := validate.Struct(form)
	if errValidate != nil {
		for _, err := range errValidate.(validator.ValidationErrors) {
			errKey := strings.ToLower(err.Field())
			_, ok := formErrors[errKey]
			if ok {
				formErrors[errKey] = append(formErrors[errKey], ErrorTagMessage(err.Tag()))
			} else {
				formErrors[errKey] = []string{ErrorTagMessage(err.Tag())}
			}
		}
	}
	return formErrors
}

func ErrorTagMessage(tag string) string {
	var result string
	switch tag {
	case "required":
		result = "Field required"
	case "email":
		result = "Invalid email format"
	}
	return result
}
