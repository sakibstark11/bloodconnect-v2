package models

import (
	"encoding/json"

	"bloodconnect/application/domain"

	"gorm.io/datatypes"
)

type Notification struct {
	ID        string `gorm:"primaryKey;size:64"`
	Type      string `gorm:"index:idx_type_recipient;size:50"`
	Recipient string `gorm:"index:idx_type_recipient;size:64"`
	Title     string
	Content   string
	Metadata  datatypes.JSONMap `gorm:"type:jsonb"`
	CreatedAt string            `gorm:"size:25"`
}



func (m *Notification) ToDomain() *domain.Notification {
	var meta domain.NotificationMetadata
	if len(m.Metadata) > 0 {
		raw, err := json.Marshal(m.Metadata)
		if err == nil {
			meta = unmarshalMetadata(domain.NotificationType(m.Type), raw)
		}
	}

	return &domain.Notification{
		ID:        domain.NotificationID(m.ID),
		Type:      domain.NotificationType(m.Type),
		Recipient: domain.UserID(m.Recipient),
		Title:     m.Title,
		Content:   m.Content,
		Metadata:  meta,
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
	}
}

func unmarshalMetadata(notifType domain.NotificationType, raw []byte) domain.NotificationMetadata {
	switch notifType {
	case domain.NotificationTypeDonationRequest:
		var m domain.DonationRequestMetadata
		if err := json.Unmarshal(raw, &m); err == nil {
			return m
		}
	case domain.NotificationTypeDonationCompletion:
		var m domain.DonationCompletionMetadata
		if err := json.Unmarshal(raw, &m); err == nil {
			return m
		}
	case domain.NotificationTypeDonationRequestAcceptance:
		var m domain.DonationAcceptanceMetadata
		if err := json.Unmarshal(raw, &m); err == nil {
			return m
		}
	}
	return nil
}

func NotificationFromDomain(n *domain.Notification) *Notification {
	var meta datatypes.JSONMap
	if n.Metadata != nil {
		raw, err := json.Marshal(n.Metadata)
		if err == nil {
			if err := json.Unmarshal(raw, &meta); err != nil {
				meta = nil
			}
		}
	}

	return &Notification{
		ID:        string(n.ID),
		Type:      string(n.Type),
		Recipient: string(n.Recipient),
		Title:     n.Title,
		Content:   n.Content,
		Metadata:  meta,
		CreatedAt: string(n.CreatedAt),
	}
}
