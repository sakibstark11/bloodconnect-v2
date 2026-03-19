package dummy

import (
	"context"
	"log"

	"bloodconnect/application"
	"bloodconnect/application/domain"
)

type dummyNotificationSender struct{}

func NewNotificationSender() application.NotificationSender {
	return &dummyNotificationSender{}
}

func (s *dummyNotificationSender) Send(ctx context.Context, notification *domain.Notification) error {
	log.Printf("DUMMY SENDER: Sending [%s] to user %s: %s\n", notification.Type, notification.Recipient, notification.Content)
	return nil
}
