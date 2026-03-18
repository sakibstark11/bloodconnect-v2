package services

import (
	"context"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"

	"github.com/oklog/ulid/v2"
)

type NotificationService interface {
	Submit(ctx context.Context, notifType domain.NotificationType, recipient, title, content string) (string, error)
	GetForUser(ctx context.Context, userID string) ([]domain.Notification, error)
}

type notificationService struct {
	repo   application.NotificationRepository
	sender application.NotificationSender
}

func NewNotificationService(repo application.NotificationRepository, sender application.NotificationSender) NotificationService {
	return &notificationService{
		repo:   repo,
		sender: sender,
	}
}

func (s *notificationService) Submit(ctx context.Context, notifType domain.NotificationType, recipient, title, content string) (string, error) {
	id := "notification_" + ulid.Make().String()

	notification := &domain.Notification{
		ID:        id,
		Type:      notifType,
		Recipient: recipient,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return "", err
	}

	// Best-effort send — notifications are fire-and-forget
	_ = s.sender.Send(ctx, notification)

	return id, nil
}

func (s *notificationService) GetForUser(ctx context.Context, userID string) ([]domain.Notification, error) {
	return s.repo.GetNotificationsForUser(ctx, userID)
}
