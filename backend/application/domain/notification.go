package domain

type NotificationID string

type NotificationType string

const (
	NotificationTypeDonationRequest           NotificationType = "blood_donation_request"
	NotificationTypeDonationCompletion        NotificationType = "blood_donation_completion"
	NotificationTypeDonationRequestAcceptance NotificationType = "blood_donation_request_acceptance"
)

type NotificationForUser struct {
	ID        NotificationID   `json:"id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Content   string           `json:"content"`
	CreatedAt ISOTimestamp     `json:"created_at"`
}

type Notification struct {
	ID        NotificationID
	Type      NotificationType
	Recipient UserID
	Title     string
	Content   string
	CreatedAt ISOTimestamp
}
