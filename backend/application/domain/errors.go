package domain

import "errors"

var (
	ErrDonationWaitPeriodNotMet = errors.New("you cannot accept another request within the minimum donation wait period")
	ErrBloodTypeUpdateDenied    = errors.New("blood type cannot be updated once it is set")
	ErrUserNotFound             = errors.New("user not found")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrPendingRequestExists     = errors.New("you already have a pending request")
	ErrIncompatibleBloodType    = errors.New("incompatible blood type")
	ErrCannotActOnOwnRequest    = errors.New("you cannot act on your own request")
	ErrRequestAlreadyClosed     = errors.New("request is already closed")
	ErrRequestNotFound          = errors.New("request not found")
	ErrUnknownError             = errors.New("unknown error")
)
