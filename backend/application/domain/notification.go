package domain

type NotificationID string

type NotificationType string

const (
	NotificationTypeDonationRequest           NotificationType = "blood_donation_request"
	NotificationTypeDonationCompletion        NotificationType = "blood_donation_completion"
	NotificationTypeDonationRequestAcceptance NotificationType = "blood_donation_request_acceptance"
)

// NotificationMetadata is a sealed interface — only concrete types in this package implement it.
type NotificationMetadata interface {
	notificationMetadata()
}

// DonationRequestMetadata is the metadata for a blood_donation_request notification.
type DonationRequestMetadata struct {
	RequestID string `json:"request_id"`
}

func (DonationRequestMetadata) notificationMetadata() {}

// DonationCompletionMetadata is the metadata for a blood_donation_completion notification.
type DonationCompletionMetadata struct {
	RequestID string `json:"request_id"`
}

func (DonationCompletionMetadata) notificationMetadata() {}

// DonationAcceptanceMetadata is the metadata for a blood_donation_request_acceptance notification.
type DonationAcceptanceMetadata struct {
	RequestID string `json:"request_id"`
	DonorID   string `json:"donor_id"`
}

func (DonationAcceptanceMetadata) notificationMetadata() {}

type Notification struct {
	ID        NotificationID
	Type      NotificationType
	Recipient UserID
	Title     string
	Content   string
	Metadata  NotificationMetadata
	CreatedAt ISOTimestamp
}
