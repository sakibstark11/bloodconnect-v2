package ports

import (
	"context"

	"github.com/sakibalam/bloodconnect/domain"
)

// NotificationRepository defines the interface for interacting with Notification storage
type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	UpdateNotification(ctx context.Context, notification *domain.Notification) error
	GetNotificationByID(ctx context.Context, id string) (*domain.Notification, error)
	DeleteNotification(ctx context.Context, id string) error
	GetNotificationsForUser(ctx context.Context, userID string) ([]domain.Notification, error)
}
