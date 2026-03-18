package domain

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeDonationRequest           NotificationType = "blood_donation_request"
	NotificationTypeDonationCompletion        NotificationType = "blood_donation_completion"
	NotificationTypeDonationRequestAcceptance NotificationType = "blood_donation_request_acceptance"
)

// Notification represents a message sent to a user.
// Notifications are one-way — they are created and sent; no status is tracked.
type Notification struct {
	ID        string           `json:"id"`
	Type      NotificationType `json:"type"`
	Recipient string           `json:"recipient"` // UserID
	Title     string           `json:"title"`
	Content   string           `json:"content"`
	CreatedAt ISOTimestamp     `json:"created_at"`
}
