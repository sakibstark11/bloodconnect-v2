package dummy

import (
	"context"
	"log"

	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
)

type dummyNotificationSender struct{}

// NewNotificationSender creates a dummy sender that just logs to stdout.
func NewNotificationSender() ports.NotificationSender {
	return &dummyNotificationSender{}
}

func (s *dummyNotificationSender) Send(ctx context.Context, notification *domain.Notification) error {
	log.Printf("DUMMY SENDER: Sending [%s] to user %s: %s\n", notification.Type, notification.Recipient, notification.Content)
	return nil
}
