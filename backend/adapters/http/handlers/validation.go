package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

type JSONError struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func RespondJSONError(w http.ResponseWriter, statusCode int, message string, details any) {
	RespondJSON(w, statusCode, JSONError{
		Error:   message,
		Details: details,
	})
}

func RespondJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

func decodeJSONBody(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

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

func parseDateTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05.000Z", s)
}
