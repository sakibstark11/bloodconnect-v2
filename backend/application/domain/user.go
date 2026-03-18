package domain

import (
	"time"
)

// User represents a registered donor or requester
type User struct {
	ID        string
	Name      string
	Email     string
	Password  string // Hashed
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
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
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserPreferredDonationLocation represents a user's location preference
type UserPreferredDonationLocation struct {
	UserID    string
	Lat       float64
	Lng       float64
	H3Hex     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
