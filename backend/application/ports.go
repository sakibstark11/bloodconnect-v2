package application

import (
	"context"

	"bloodconnect/application/domain"
)

// RequestFilters holds optional filters for listing donation requests
type RequestFilters struct {
	BloodType   string
	Status      string
	LocationHex string
}

// UserRepository defines the interface for interacting with User data storage.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUserHealth(ctx context.Context, health *domain.UserHealth) error
	GetUserHealth(ctx context.Context, userID string) ([]domain.UserHealth, error)
	UpdateUserLocation(ctx context.Context, loc *domain.UserPreferredDonationLocation) error
	GetUserLocation(ctx context.Context, userID string) (*domain.UserPreferredDonationLocation, error)
	GetEligibleUsersInHexes(ctx context.Context, hexes []string, bloodType string, count int) ([]domain.User, error)
}

// RequestRepository defines the interface for interacting with DonationRequest data storage.
type RequestRepository interface {
	CreateRequest(ctx context.Context, req *domain.DonationRequest) error
	UpdateRequest(ctx context.Context, req *domain.DonationRequest) error
	GetRequestByID(ctx context.Context, id string) (*domain.DonationRequest, error)
	ListRequests(ctx context.Context, filters RequestFilters, page, pageSize int) ([]domain.DonationRequest, int, error)
	SaveRequestState(ctx context.Context, state *domain.RequestState) error
	GetRequestState(ctx context.Context, requestID, actionedByID string) (*domain.RequestState, error)
	GetActionedUsers(ctx context.Context, requestID string) ([]domain.RequestState, error)
}

// NotificationRepository defines the interface for interacting with Notification data storage.
// Notifications are one-way — only creation and read by recipient are supported.
type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	GetNotificationsForUser(ctx context.Context, userID string) ([]domain.Notification, error)
}

// NotificationSender defines the interface for dispatching notifications via an external channel.
type NotificationSender interface {
	Send(ctx context.Context, notification *domain.Notification) error
}

// JobQueue defines the interface for managing background job queuing.
type JobQueue interface {
	Enqueue(ctx context.Context, job *domain.Job) error
	FetchNextAvailable(ctx context.Context) (*domain.Job, error)
	MarkStatus(ctx context.Context, id string, status domain.JobStatus) error
	Delete(ctx context.Context, id string) error
}
