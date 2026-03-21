package domain

import "errors"

var (
	ErrDonationWaitPeriodNotMet = errors.New("you cannot accept another request within the minimum donation wait period")
	ErrBloodTypeUpdateDenied    = errors.New("blood type cannot be updated once it is set")
	ErrUserNotFound             = errors.New("user not found")
	ErrUnauthorized             = errors.New("unauthorized")
)
