package services

import (
	"context"

	"bloodconnect/application"
	"bloodconnect/application/domain"

	"github.com/oklog/ulid/v2"
)

type NotificationService interface {
	Submit(ctx context.Context, notifType domain.NotificationType, recipient domain.UserID, title, content string) (domain.NotificationID, error)
	GetForUser(ctx context.Context, userID domain.UserID, page, pageSize int) ([]domain.Notification, int, error)
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

func (s *notificationService) Submit(ctx context.Context, notifType domain.NotificationType, recipient domain.UserID, title, content string) (domain.NotificationID, error) {
	id := domain.NotificationID("notification_" + ulid.Make().String())

	notification := &domain.Notification{
		ID:        id,
		Type:      notifType,
		Recipient: recipient,
		Title:     title,
		Content:   content,
		CreatedAt: domain.Now(),
	}

	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return "", err
	}

	_ = s.sender.Send(ctx, notification)

	return id, nil
}

func (s *notificationService) GetForUser(ctx context.Context, userID domain.UserID, page, pageSize int) ([]domain.Notification, int, error) {
	return s.repo.GetNotificationsForUser(ctx, userID, page, pageSize)
}
