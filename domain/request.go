package domain

import "time"

// RequestStatus represents the state of the overall donation request
type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "Pending"
	RequestStatusCompleted RequestStatus = "Completed"
	RequestStatusCancelled RequestStatus = "Cancelled"
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
