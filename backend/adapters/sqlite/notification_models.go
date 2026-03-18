package sqlite

import (
	"time"

	"github.com/sakibalam/bloodconnect/domain"
)

type notificationModel struct {
	ID        string `gorm:"primaryKey;size:64"`
	Type      string `gorm:"index:idx_type_recipient;size:50"`
	Recipient string `gorm:"index:idx_type_recipient;size:64"`
	Status    string `gorm:"size:20"`
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *notificationModel) toDomain() *domain.Notification {
	return &domain.Notification{
		ID:        m.ID,
		Type:      domain.NotificationType(m.Type),
		Recipient: m.Recipient,
		Status:    domain.NotificationStatus(m.Status),
		Title:     m.Title,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromDomainNotification(n *domain.Notification) *notificationModel {
	return &notificationModel{
		ID:        n.ID,
		Type:      string(n.Type),
		Recipient: n.Recipient,
		Status:    string(n.Status),
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}
