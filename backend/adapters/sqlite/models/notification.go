package models

import (
	"time"

	"bloodconnect/application/domain"
)

// Notification is the GORM model for domain.Notification
// Notifications are one-way — no status field.
type Notification struct {
	ID        string `gorm:"primaryKey;size:64"`
	Type      string `gorm:"index:idx_type_recipient;size:50"`
	Recipient string `gorm:"index:idx_type_recipient;size:64"`
	Title     string
	Content   string
	CreatedAt time.Time
}

func (m *Notification) ToDomain() *domain.Notification {
	return &domain.Notification{
		ID:        m.ID,
		Type:      domain.NotificationType(m.Type),
		Recipient: m.Recipient,
		Title:     m.Title,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func NotificationFromDomain(n *domain.Notification) *Notification {
	return &Notification{
		ID:        n.ID,
		Type:      string(n.Type),
		Recipient: n.Recipient,
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
	}
}
