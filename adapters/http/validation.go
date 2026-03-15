package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// Validate is the global validator instance used across HTTP handlers
var Validate = validator.New()

// JSONError represents a structured API error response
type JSONError struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// RespondJSONError sends a structured JSON error response
func RespondJSONError(w http.ResponseWriter, statusCode int, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	resp := JSONError{
		Error:   message,
		Details: details,
	}
	json.NewEncoder(w).Encode(resp)
}

// FormatValidationErrors converts validator.ValidationErrors into a simple map for JSON responses
func FormatValidationErrors(err error) map[string]string {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errMap := make(map[string]string)
		for _, e := range validationErrs {
			// Instead of a giant string, this creates a clean mapping like {"bag_count": "required"}
			errMap[e.Field()] = "failed validation on tag '" + e.Tag() + "'"
		}
		return errMap
	}
	return nil
}
