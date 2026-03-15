package ports

import (
	"context"

	"github.com/sakibalam/bloodconnect/domain"
)

// NotificationSender defines the port for sending notifications to users
type NotificationSender interface {
	Send(ctx context.Context, notification *domain.Notification) error
}
