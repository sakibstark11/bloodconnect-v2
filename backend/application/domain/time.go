package domain

import (
	"time"
)

// ISOTimestamp is a string representation of time in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ).
type ISOTimestamp string

// Now returns the current UTC time as an ISOTimestamp.
func Now() ISOTimestamp {
	return ISOTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
}

// ToTime converts an ISOTimestamp back to time.Time for calculations.
func (t ISOTimestamp) ToTime() time.Time {
	tt, _ := time.Parse("2006-01-02T15:04:05.000Z", string(t))
	return tt
}

// String returns the string value.
func (t ISOTimestamp) String() string {
	return string(t)
}
