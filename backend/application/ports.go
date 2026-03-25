package application

import (
	"context"

	"bloodconnect/application/domain"
)

type RequestFilters struct {
	BloodType   domain.BloodType
	Status      string
	LocationHex string
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserAuthByEmail(ctx context.Context, email string) (*domain.UserAuth, error)
	GetUserHealth(ctx context.Context, userID domain.UserID) ([]*domain.UserHealth, error)
	GetUserHealthByType(ctx context.Context, userID domain.UserID, infoType domain.InfoType) (*domain.UserHealth, error)
	GetUserLocation(ctx context.Context, userID domain.UserID) ([]*domain.UserPreferredDonationLocation, error)
	GetEligibleUsersInHexes(ctx context.Context, hexes []string, bloodType domain.BloodType, count int, excludedUserIDs []domain.UserID) ([]*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User, hashedPassword string) error
	UpdateUserHealth(ctx context.Context, health *domain.UserHealth) error
	UpdateUserLocation(ctx context.Context, loc *domain.UserPreferredDonationLocation) error
	DeleteUserLocation(ctx context.Context, userID domain.UserID, h3hex string) error
}

type RequestRepository interface {
	CreateRequest(ctx context.Context, req *domain.DonationRequest) error
	UpdateRequest(ctx context.Context, req *domain.DonationRequest) error
	GetRequestByID(ctx context.Context, id domain.RequestID) (*domain.DonationRequest, error)
	ListRequests(ctx context.Context, filters RequestFilters, lastRequestID domain.RequestID, pageSize int) ([]*domain.DonationRequest, error)
	SaveRequestState(ctx context.Context, state *domain.RequestState) error
	GetRequestState(ctx context.Context, requestID domain.RequestID, actionedByID domain.UserID) (*domain.RequestState, error)
	GetActionedUsers(ctx context.Context, requestID domain.RequestID) ([]*domain.RequestState, error)
	GetUserRecentActions(ctx context.Context, userID domain.UserID, actions []domain.ActionStatus, since domain.ISOTimestamp) ([]*domain.RequestState, error)
}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	GetNotificationByID(ctx context.Context, id domain.NotificationID) (*domain.Notification, error)
	GetNotificationsForUser(ctx context.Context, userID domain.UserID, lastNotificationID domain.NotificationID, pageSize int) ([]*domain.Notification, error)
}

type NotificationSender interface {
	Send(ctx context.Context, notification *domain.Notification) error
}

type JobQueue interface {
	Enqueue(ctx context.Context, job *domain.Job) error
	Consume(ctx context.Context, jobType domain.JobType) (<-chan *domain.Job, error)
	Close() error
}
