package repos

import (
	"context"

	"bloodconnect/adapters/sqlite/models"
	"bloodconnect/application"
	"bloodconnect/application/domain"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) application.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(models.NotificationFromDomain(notification)).Error
}

func (r *notificationRepository) GetNotificationsForUser(ctx context.Context, userID string) ([]domain.Notification, error) {
	var ms []models.Notification
	if err := r.db.WithContext(ctx).Where("recipient = ?", userID).Order("created_at desc").Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Notification, len(ms))
	for i, m := range ms {
		result[i] = *m.ToDomain()
	}
	return result, nil
}
