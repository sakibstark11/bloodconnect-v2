package memory

import (
	"context"
	"sync"

	"bloodconnect/application"
	"bloodconnect/application/domain"
)

type InMemoryNotificationRepository struct {
	mu            sync.RWMutex
	notifications map[domain.UserID][]domain.Notification
}

func NewNotificationRepository() application.NotificationRepository {
	return &InMemoryNotificationRepository{
		notifications: make(map[domain.UserID][]domain.Notification),
	}
}

func (r *InMemoryNotificationRepository) CreateNotification(ctx context.Context, n *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications[n.Recipient] = append(r.notifications[n.Recipient], *n)
	return nil
}

func (r *InMemoryNotificationRepository) GetNotificationsForUser(ctx context.Context, userID domain.UserID, page, pageSize int) ([]domain.Notification, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notifs, ok := r.notifications[userID]
	if !ok {
		return []domain.Notification{}, 0, nil
	}

	total := len(notifs)
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= total {
		return []domain.Notification{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return notifs[start:end], total, nil
}
