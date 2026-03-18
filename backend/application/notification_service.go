package application

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
)

type NotificationService interface {
	Submit(ctx context.Context, notifType domain.NotificationType, recipient, title, content string) (string, error)
	UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error
	Delete(ctx context.Context, id string) error
	GetForUser(ctx context.Context, userID string) ([]domain.Notification, error)
}

type notificationService struct {
	repo   ports.NotificationRepository
	sender ports.NotificationSender
}

func NewNotificationService(repo ports.NotificationRepository, sender ports.NotificationSender) NotificationService {
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
		Status:    domain.NotificationStatusPending,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return "", err
	}

	// Trigger sender (for now inline)
	if err := s.sender.Send(ctx, notification); err == nil {
		notification.Status = domain.NotificationStatusSent
		_ = s.repo.UpdateNotification(ctx, notification)
	}

	return id, nil
}

func (s *notificationService) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error {
	notification, err := s.repo.GetNotificationByID(ctx, id)
	if err != nil {
		return err
	}

	notification.Status = status
	notification.UpdatedAt = time.Now()

	return s.repo.UpdateNotification(ctx, notification)
}

func (s *notificationService) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteNotification(ctx, id)
}

func (s *notificationService) GetForUser(ctx context.Context, userID string) ([]domain.Notification, error) {
	return s.repo.GetNotificationsForUser(ctx, userID)
}
