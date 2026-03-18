package domain

import "time"

type ContextKey string

const (
	TraceIDKey ContextKey = "trace_id"
)

// RequestStatus represents the state of the overall donation request
type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "Pending"
	RequestStatusCompleted RequestStatus = "Completed"
	RequestStatusCancelled RequestStatus = "Cancelled"
	RequestStatusFailed    RequestStatus = "Failed"
)

// BloodType represents the standard A/B/O blood types
type BloodType string

const (
	BloodTypeAPos  BloodType = "A+"
	BloodTypeANeg  BloodType = "A-"
	BloodTypeBPos  BloodType = "B+"
	BloodTypeBNeg  BloodType = "B-"
	BloodTypeABPos BloodType = "AB+"
	BloodTypeABNeg BloodType = "AB-"
	BloodTypeOPos  BloodType = "O+"
	BloodTypeONeg  BloodType = "O-"
)

// DonationRequest represents a user asking for a blood donation
type DonationRequest struct {
	ID             string
	UserID         string
	LocationHex    string
	LocationLat    float64
	LocationLng    float64
	BagCount       int
	RequiredByDate time.Time
	BloodType      BloodType
	ContactPhone   string
	Description    string
	RequesterInfo  string
	LocationName   string
	Status         RequestStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// RequestActionedUser represents an individual tracking response
type RequestActionedUser struct {
	UserID string       `json:"user_id"`
	Lat    float64      `json:"lat"`
	Lng    float64      `json:"lng"`
	H3Hex  string       `json:"h3_hex"`
	Action ActionStatus `json:"action"`
}

// ExtendedDonationRequest provides the request details alongside the locations of all people notified
type ExtendedDonationRequest struct {
	Request       *DonationRequest       `json:"request"`
	NotifiedUsers []RequestActionedUser  `json:"notified_users"`
}

// ActionStatus represents the individual donor's response to a request
type ActionStatus string

const (
	ActionStatusPending  ActionStatus = "Pending"
	ActionStatusAccepted ActionStatus = "Accepted"
	ActionStatusDeclined ActionStatus = "Declined"
	ActionStatusDonated  ActionStatus = "Donated"
)

// RequestState tracks which users have been pinged and their decisions
type RequestState struct {
	RequestID    string
	ActionedByID string // UserID
	Action       ActionStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
