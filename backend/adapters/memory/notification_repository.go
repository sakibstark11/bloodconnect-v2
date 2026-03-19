package memory

import (
	"context"
	"sort"
	"sync"

	"bloodconnect/application"
	"bloodconnect/application/domain"
)

type InMemoryNotificationRepository struct {
	mu            sync.RWMutex
	notifications map[domain.UserID][]domain.NotificationForUser
}

func NewNotificationRepository() application.NotificationRepository {
	return &InMemoryNotificationRepository{
		notifications: make(map[domain.UserID][]domain.NotificationForUser),
	}
}

func (r *InMemoryNotificationRepository) CreateNotification(ctx context.Context, n *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	notificationForUser := &domain.NotificationForUser{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
	}
	r.notifications[n.Recipient] = append(r.notifications[n.Recipient], *notificationForUser)
	return nil
}

func (r *InMemoryNotificationRepository) GetNotificationsForUser(ctx context.Context, userID domain.UserID, lastNotificationID domain.NotificationID, pageSize int) ([]domain.NotificationForUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notifs, ok := r.notifications[userID]
	if !ok {
		return []domain.NotificationForUser{}, nil
	}

	// Sort by ID descending
	sort.Slice(notifs, func(i, j int) bool {
		return notifs[i].ID > notifs[j].ID
	})

	var result []domain.NotificationForUser
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
