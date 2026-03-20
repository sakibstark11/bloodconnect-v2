package domain

type UserID string

type User struct {
	ID        UserID
	Name      string
	Email     string
	Phone     string
	CreatedAt ISOTimestamp
	UpdatedAt ISOTimestamp
}

type UserAuth struct {
	UserID   UserID
	Password string
}

type InfoType string

const (
	InfoTypeBloodType    InfoType = "blood_type"
	InfoTypeWeight       InfoType = "weight"
	InfoTypeHeight       InfoType = "height"
	InfoTypeLastDonation InfoType = "last_donation_date"
	InfoTypeLastVaccine  InfoType = "last_vaccination_date"
	InfoTypeMedicalCond  InfoType = "medical_condition"
)

type UserHealth struct {
	UserID    UserID
	InfoType  InfoType
	Details   string
	CreatedAt ISOTimestamp
	UpdatedAt ISOTimestamp
}

type UserPreferredDonationLocation struct {
	UserID    UserID
	Lat       float64
	Lng       float64
	H3Hex     string
	CreatedAt ISOTimestamp
	UpdatedAt ISOTimestamp
}
