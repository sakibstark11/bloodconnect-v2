package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

// Validate is the global validator instance used across handlers
var Validate = validator.New()

// JSONError represents a structured API error response
type JSONError struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// RespondJSONError sends a structured JSON error response
func RespondJSONError(w http.ResponseWriter, statusCode int, message string, details any) {
	RespondJSON(w, statusCode, JSONError{
		Error:   message,
		Details: details,
	})
}

// RespondJSON writes a JSON-encoded response with the given status code and body.
func RespondJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

// decodeJSONBody decodes r.Body into dst and returns an error for invalid JSON.
func decodeJSONBody(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// FormatValidationErrors converts validator.ValidationErrors into a simple map for JSON responses
func FormatValidationErrors(err error) map[string]string {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errMap := make(map[string]string)
		for _, e := range validationErrs {
			errMap[e.Field()] = "failed validation on tag '" + e.Tag() + "'"
		}
		return errMap
	}
	return nil
}

// parseDateTime parses a RFC3339 datetime string
func parseDateTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05.000Z", s)
}
