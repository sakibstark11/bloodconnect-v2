package domain

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeDonationRequest          NotificationType = "blood_donation_request"
	NotificationTypeDonationCompletion       NotificationType = "blood_donation_completion"
	NotificationTypeDonationRequestAcceptance NotificationType = "blood_donation_request_acceptance"
)

// NotificationStatus represents the delivery status
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "Pending"
	NotificationStatusSent      NotificationStatus = "Sent"
	NotificationStatusFailed    NotificationStatus = "Failed"
	NotificationStatusCompleted NotificationStatus = "Completed"
)

// Notification represents a message sent to a user
type Notification struct {
	ID        string
	Type      NotificationType
	Recipient string // UserID
	Status    NotificationStatus
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
