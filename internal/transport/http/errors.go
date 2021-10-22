package http

import (
	"encoding/json"
	"net/http"
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
