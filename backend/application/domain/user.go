package domain

import ()

// User represents a registered donor or requester
type User struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Phone     string       `json:"phone"`
	CreatedAt ISOTimestamp `json:"created_at"`
	UpdatedAt ISOTimestamp `json:"updated_at"`
}

// UserAuth represents security credentials for a user
type UserAuth struct {
	UserID   string
	Password string // Hashed
}

// InfoType represents the type of health information
type InfoType string

const (
	InfoTypeBloodType    InfoType = "blood_type"
	InfoTypeWeight       InfoType = "weight"
	InfoTypeHeight       InfoType = "height"
	InfoTypeLastDonation InfoType = "last_donation_date"
	InfoTypeLastVaccine  InfoType = "last_vaccination_date"
	InfoTypeMedicalCond  InfoType = "medical_condition"
)

// UserHealth represents specific health details for a user
type UserHealth struct {
	UserID    string
	InfoType  InfoType
	Details   string
	CreatedAt ISOTimestamp
	UpdatedAt ISOTimestamp
}

// UserPreferredDonationLocation represents a user's location preference
type UserPreferredDonationLocation struct {
	UserID    string
	Lat       float64
	Lng       float64
	H3Hex     string
	CreatedAt ISOTimestamp
	UpdatedAt ISOTimestamp
}
