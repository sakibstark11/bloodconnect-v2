package domain

type NotificationID string

type NotificationType string

const (
	NotificationTypeDonationRequest           NotificationType = "blood_donation_request"
	NotificationTypeDonationCompletion        NotificationType = "blood_donation_completion"
	NotificationTypeDonationRequestAcceptance NotificationType = "blood_donation_request_acceptance"
)

type NotificationMetadata interface{}

type DonationRequestMetadata struct {
	RequestID string `json:"request_id"`
}

type DonationCompletionMetadata struct {
	RequestID string `json:"request_id"`
}

type DonationAcceptanceMetadata struct {
	RequestID string `json:"request_id"`
	DonorID   string `json:"donor_id"`
}

type Notification struct {
	ID        NotificationID
	Type      NotificationType
	Recipient UserID
	Title     string
	Content   string
	Metadata  NotificationMetadata
	CreatedAt ISOTimestamp
}
