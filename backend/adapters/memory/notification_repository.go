package memory

import (
	"context"
	"sync"
	"sort"

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

func (r *InMemoryNotificationRepository) GetNotificationsForUser(ctx context.Context, userID domain.UserID, lastNotificationID domain.NotificationID, pageSize int) ([]domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notifs, ok := r.notifications[userID]
	if !ok {
		return []domain.Notification{}, nil
	}

	// Sort by ID descending
	sort.Slice(notifs, func(i, j int) bool {
		return notifs[i].ID > notifs[j].ID
	})

	var result []domain.Notification
	foundStart := lastNotificationID == ""
	for _, n := range notifs {
		if !foundStart {
			if n.ID == lastNotificationID {
				foundStart = true
			}
			continue
		}

		if len(result) < pageSize {
			if lastNotificationID != "" && n.ID == lastNotificationID {
				continue
			}
			result = append(result, n)
		}
	}

	return result, nil
}
