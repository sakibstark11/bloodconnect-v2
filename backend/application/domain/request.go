package domain

type RequestID string

type ContextKey string

const (
	TraceIDKey ContextKey = "trace_id"
	UserIDKey  ContextKey = "user_id"
)

type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "Pending"
	RequestStatusCompleted RequestStatus = "Completed"
	RequestStatusCancelled RequestStatus = "Cancelled"
	RequestStatusFailed    RequestStatus = "Failed"
)

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

type DonationRequest struct {
	ID             RequestID
	UserID         UserID
	LocationHex    string
	LocationLat    float64
	LocationLng    float64
	BagCount       int
	RequiredByDate ISOTimestamp
	BloodType      BloodType
	Description    string
	RequesterInfo  string
	LocationName   string
	Status         RequestStatus
	CreatedAt      ISOTimestamp
	UpdatedAt      ISOTimestamp
}

type RequestActionedUser struct {
	UserID UserID
	Lat    float64
	Lng    float64
	H3Hex  string
	Action ActionStatus
}

type ExtendedDonationRequest struct {
	Request       *DonationRequest
	NotifiedUsers []RequestActionedUser
}

type ActionStatus string

const (
	ActionStatusPending  ActionStatus = "Pending"
	ActionStatusAccepted ActionStatus = "Accepted"
	ActionStatusDeclined ActionStatus = "Declined"
	ActionStatusDonated  ActionStatus = "Donated"
)

type RequestState struct {
	RequestID    RequestID
	ActionedByID UserID
	Action       ActionStatus
	CreatedAt    ISOTimestamp
	UpdatedAt    ISOTimestamp
}
