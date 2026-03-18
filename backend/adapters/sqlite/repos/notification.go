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

func (r *notificationRepository) GetNotificationsForUser(ctx context.Context, userID string, page, pageSize int) ([]domain.Notification, int, error) {
	var ms []models.Notification
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Notification{}).Where("recipient = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&ms).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domain.Notification, len(ms))
	for i, m := range ms {
		result[i] = *m.ToDomain()
	}
	return result, int(total), nil
}
