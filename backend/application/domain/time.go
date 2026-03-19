package domain

import (
	"time"
)

type ISOTimestamp string

func Now() ISOTimestamp {
	return ISOTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
}

func (t ISOTimestamp) ToTime() time.Time {
	tt, _ := time.Parse("2006-01-02T15:04:05.000Z", string(t))
	return tt
}

func (t ISOTimestamp) String() string {
	return string(t)
}
