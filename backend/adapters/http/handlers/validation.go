package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"bloodconnect/application/domain"
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

func RespondWithError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	statusCode := http.StatusInternalServerError
	message := "Internal server error"
	var details any

	switch err {
	case domain.ErrUnauthorized:
		statusCode = http.StatusUnauthorized
		message = "Unauthorized"
	case domain.ErrUserNotFound, domain.ErrRequestNotFound:
		statusCode = http.StatusNotFound
		message = "Resource not found"
		details = err.Error()
	case domain.ErrIncompatibleBloodType, domain.ErrPendingRequestExists, domain.ErrDonationWaitPeriodNotMet, domain.ErrBloodTypeUpdateDenied:
		statusCode = http.StatusBadRequest
		message = "Eligibility check failed"
		details = err.Error()
	case domain.ErrCannotActOnOwnRequest:
		statusCode = http.StatusForbidden
		message = "Forbidden action"
		details = err.Error()
	case domain.ErrRequestAlreadyClosed:
		statusCode = http.StatusGone
		message = "Request already closed"
	case domain.ErrEmailAlreadyInUse:
		statusCode = http.StatusConflict
		message = "Conflict"
		details = err.Error()
	default:
		if err.Error() == "unauthorized to cancel this request" {
			statusCode = http.StatusForbidden
			message = "Permission denied"
			details = err.Error()
		} else if err.Error() == "email already in use" {
			statusCode = http.StatusConflict
			message = "Conflict"
			details = err.Error()
		}
	}

	RespondJSONError(w, statusCode, message, details)
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
