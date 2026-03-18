package domain

import "time"

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
	ID        string
	Type      NotificationType
	Recipient string // UserID
	Title     string
	Content   string
	CreatedAt time.Time
}
