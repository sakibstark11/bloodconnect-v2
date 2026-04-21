package handlers

import (
	"encoding/json"
	"errors"
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
	details := err.Error()

	if errors.Is(err, domain.ErrUnauthorized) {
		statusCode = http.StatusUnauthorized
		message = "Unauthorized"
	} else if errors.Is(err, domain.ErrForbidden) {
		statusCode = http.StatusForbidden
		message = "Forbidden"
	} else if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrRequestNotFound) {
		statusCode = http.StatusNotFound
		message = "Resource not found"
	} else if errors.Is(err, domain.ErrIncompatibleBloodType) || 
	          errors.Is(err, domain.ErrPendingRequestExists) || 
			  errors.Is(err, domain.ErrDonationWaitPeriodNotMet) || 
			  errors.Is(err, domain.ErrBloodTypeUpdateDenied) ||
			  errors.Is(err, domain.ErrLastLocationDeleteDenied) {
		statusCode = http.StatusBadRequest
		message = "Action denied"
	} else if errors.Is(err, domain.ErrCannotActOnOwnRequest) {
		statusCode = http.StatusForbidden
		message = "Forbidden action"
	} else if errors.Is(err, domain.ErrRequestAlreadyClosed) {
		statusCode = http.StatusGone
		message = "Request already closed"
	} else if errors.Is(err, domain.ErrEmailAlreadyInUse) {
		statusCode = http.StatusConflict
		message = "Email already in use"
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
