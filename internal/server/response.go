package server

import (
	"encoding/json"
	"net/http"
)

// apiError is the error object returned inside the standard envelope.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// envelope is the standard API response shape used across all Dojah services.
type envelope struct {
	Entity any       `json:"entity"`
	Error  *apiError `json:"error"`
	Status bool      `json:"status"`
}

func writeJSON(w http.ResponseWriter, status int, payload envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// Best-effort encode; the response headers are already committed.
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSuccess(w http.ResponseWriter, status int, entity any) {
	writeJSON(w, status, envelope{Entity: entity, Error: nil, Status: true})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, envelope{
		Entity: nil,
		Error:  &apiError{Code: code, Message: message},
		Status: false,
	})
}
