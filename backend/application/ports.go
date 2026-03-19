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
	CreateUser(ctx context.Context, user *domain.User, hashedPassword string) error
	GetUserByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserAuthByEmail(ctx context.Context, email string) (*domain.UserAuth, error)
	UpdateUserHealth(ctx context.Context, health *domain.UserHealth) error
	GetUserHealth(ctx context.Context, userID domain.UserID) ([]domain.UserHealth, error)
	UpdateUserLocation(ctx context.Context, loc *domain.UserPreferredDonationLocation) error
	GetUserLocation(ctx context.Context, userID domain.UserID) (*domain.UserPreferredDonationLocation, error)
	GetEligibleUsersInHexes(ctx context.Context, hexes []string, bloodType domain.BloodType, count int, excludedUserIDs []domain.UserID) ([]domain.User, error)
}

type RequestRepository interface {
	CreateRequest(ctx context.Context, req *domain.DonationRequest) error
	UpdateRequest(ctx context.Context, req *domain.DonationRequest) error
	GetRequestByID(ctx context.Context, id domain.RequestID) (*domain.DonationRequest, error)
	ListRequests(ctx context.Context, filters RequestFilters, page, pageSize int) ([]domain.DonationRequest, int, error)
	SaveRequestState(ctx context.Context, state *domain.RequestState) error
	GetRequestState(ctx context.Context, requestID domain.RequestID, actionedByID domain.UserID) (*domain.RequestState, error)
	GetActionedUsers(ctx context.Context, requestID domain.RequestID) ([]domain.RequestState, error)
	GetUserRecentActions(ctx context.Context, userID domain.UserID, action domain.ActionStatus, since domain.ISOTimestamp) ([]domain.RequestState, error)
}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	GetNotificationsForUser(ctx context.Context, userID domain.UserID, page, pageSize int) ([]domain.Notification, int, error)
}

type NotificationSender interface {
	Send(ctx context.Context, notification *domain.Notification) error
}

type JobQueue interface {
	Enqueue(ctx context.Context, job *domain.Job) error
	FetchNextAvailable(ctx context.Context) (*domain.Job, error)
	MarkStatus(ctx context.Context, id domain.JobID, status domain.JobStatus) error
	Delete(ctx context.Context, id domain.JobID) error
}
