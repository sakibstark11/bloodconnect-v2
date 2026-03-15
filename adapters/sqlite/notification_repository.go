package sqlite

import (
	"context"

	"gorm.io/gorm"

	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new SQLite notification repository
func NewNotificationRepository(db *gorm.DB) ports.NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

func (r *notificationRepository) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	m := fromDomainNotification(notification)
	res := r.db.WithContext(ctx).Create(m)
	return res.Error
}

func (r *notificationRepository) UpdateNotification(ctx context.Context, notification *domain.Notification) error {
	m := fromDomainNotification(notification)
	res := r.db.WithContext(ctx).Save(m)
	return res.Error
}

func (r *notificationRepository) GetNotificationByID(ctx context.Context, id string) (*domain.Notification, error) {
	var m notificationModel
	res := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	return m.toDomain(), nil
}

func (r *notificationRepository) DeleteNotification(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&notificationModel{})
	return res.Error
}

func (r *notificationRepository) GetNotificationsForUser(ctx context.Context, userID string) ([]domain.Notification, error) {
	var models []notificationModel
	res := r.db.WithContext(ctx).Where("recipient = ?", userID).Find(&models)
	if res.Error != nil {
		return nil, res.Error
	}

	var d []domain.Notification
	for _, m := range models {
		d = append(d, *m.toDomain())
	}
	return d, nil
}
