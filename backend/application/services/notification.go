package services

import (
	"context"

	"bloodconnect/application"
	"bloodconnect/application/domain"

	"github.com/oklog/ulid/v2"
)

type NotificationService interface {
	Submit(ctx context.Context, notifType domain.NotificationType, recipient domain.UserID, title, content string, metadata domain.NotificationMetadata) (domain.NotificationID, error)
	GetForUser(ctx context.Context, userID domain.UserID, lastNotificationID domain.NotificationID, pageSize int) ([]domain.Notification, error)
}

type notificationService struct {
	repo  application.NotificationRepository
	queue application.JobQueue
}

func NewNotificationService(repo application.NotificationRepository, queue application.JobQueue) NotificationService {
	return &notificationService{
		repo:  repo,
		queue: queue,
	}
}

func (s *notificationService) Submit(ctx context.Context, notifType domain.NotificationType, recipient domain.UserID, title, content string, metadata domain.NotificationMetadata) (domain.NotificationID, error) {
	id := domain.NotificationID("notification_" + ulid.Make().String())

	notification := &domain.Notification{
		ID:        id,
		Type:      notifType,
		Recipient: recipient,
		Title:     title,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: domain.Now(),
	}

	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return "", err
	}

	_ = s.queue.Enqueue(ctx, &domain.Job{
		ID:      domain.JobID("job_" + ulid.Make().String()),
		Type:    domain.JobTypeNotification,
		Payload: map[string]interface{}{"notification_id": string(id)},
		RunAt:   domain.Now(),
	})

	return id, nil
}

func (s *notificationService) GetForUser(ctx context.Context, userID domain.UserID, lastNotificationID domain.NotificationID, pageSize int) ([]domain.Notification, error) {
	return s.repo.GetNotificationsForUser(ctx, userID, lastNotificationID, pageSize)
}
