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
	ID             RequestID     `json:"id"`
	UserID         UserID        `json:"user_id"`
	LocationHex    string        `json:"location_hex"`
	LocationLat    float64       `json:"location_lat"`
	LocationLng    float64       `json:"location_lng"`
	BagCount       int           `json:"bag_count"`
	RequiredByDate ISOTimestamp  `json:"required_by_date"`
	BloodType      BloodType     `json:"blood_type"`
	Description    string        `json:"description"`
	RequesterInfo  string        `json:"requester_info"`
	LocationName   string        `json:"location_name"`
	Status         RequestStatus `json:"status"`
	CreatedAt      ISOTimestamp  `json:"created_at"`
	UpdatedAt      ISOTimestamp  `json:"updated_at"`
}

type RequestActionedUser struct {
	UserID UserID       `json:"user_id"`
	Lat    float64      `json:"lat"`
	Lng    float64      `json:"lng"`
	H3Hex  string       `json:"h3_hex"`
	Action ActionStatus `json:"action"`
}

type ExtendedDonationRequest struct {
	Request       *DonationRequest      `json:"request"`
	NotifiedUsers []RequestActionedUser `json:"notified_users"`
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
