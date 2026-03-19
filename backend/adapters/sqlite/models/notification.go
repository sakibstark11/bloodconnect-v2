package models

import (
	"bloodconnect/application/domain"
)

type Notification struct {
	ID        string `gorm:"primaryKey;size:64"`
	Type      string `gorm:"index:idx_type_recipient;size:50"`
	Recipient string `gorm:"index:idx_type_recipient;size:64"`
	Title     string
	Content   string
	CreatedAt string `gorm:"size:25"`
}

func (Notification) TableName() string {
	return "notifications"
}

type NotificationForUser struct {
	ID        string `gorm:"primaryKey;size:64"`
	Type      string `gorm:"index:idx_type_recipient;size:50"`
	Title     string
	Content   string
	CreatedAt string `gorm:"size:25"`
}

func (NotificationForUser) TableName() string {
	return "notifications"
}

func (m *NotificationForUser) ToDomain() *domain.NotificationForUser {
	return &domain.NotificationForUser{
		ID:        domain.NotificationID(m.ID),
		Type:      domain.NotificationType(m.Type),
		Title:     m.Title,
		Content:   m.Content,
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
	}
}
func (m *Notification) ToDomain() *domain.Notification {
	return &domain.Notification{
		ID:        domain.NotificationID(m.ID),
		Type:      domain.NotificationType(m.Type),
		Recipient: domain.UserID(m.Recipient),
		Title:     m.Title,
		Content:   m.Content,
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
	}
}

func NotificationFromDomain(n *domain.Notification) *Notification {
	return &Notification{
		ID:        string(n.ID),
		Type:      string(n.Type),
		Recipient: string(n.Recipient),
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: string(n.CreatedAt),
	}
}
