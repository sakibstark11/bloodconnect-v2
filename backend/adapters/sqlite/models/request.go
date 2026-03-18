package models

import (
	"time"

	"bloodconnect/application/domain"
)

// Request is the GORM model for domain.DonationRequest
type Request struct {
	ID             string `gorm:"primaryKey;size:64"`
	UserID         string `gorm:"index;size:64"`
	LocationHex    string `gorm:"index;size:20"`
	LocationLat    float64
	LocationLng    float64
	BagCount       int
	RequiredByDate time.Time
	BloodType      string `gorm:"size:10"`
	ContactPhone   string
	Description    string
	RequesterInfo  string
	LocationName   string
	Status         string `gorm:"index;size:20"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (m *Request) ToDomain() *domain.DonationRequest {
	return &domain.DonationRequest{
		ID:             m.ID,
		UserID:         m.UserID,
		LocationHex:    m.LocationHex,
		LocationLat:    m.LocationLat,
		LocationLng:    m.LocationLng,
		BagCount:       m.BagCount,
		RequiredByDate: m.RequiredByDate,
		BloodType:      domain.BloodType(m.BloodType),
		ContactPhone:   m.ContactPhone,
		Description:    m.Description,
		RequesterInfo:  m.RequesterInfo,
		LocationName:   m.LocationName,
		Status:         domain.RequestStatus(m.Status),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func RequestFromDomain(r *domain.DonationRequest) *Request {
	return &Request{
		ID:             r.ID,
		UserID:         r.UserID,
		LocationHex:    r.LocationHex,
		LocationLat:    r.LocationLat,
		LocationLng:    r.LocationLng,
		BagCount:       r.BagCount,
		RequiredByDate: r.RequiredByDate,
		BloodType:      string(r.BloodType),
		ContactPhone:   r.ContactPhone,
		Description:    r.Description,
		RequesterInfo:  r.RequesterInfo,
		LocationName:   r.LocationName,
		Status:         string(r.Status),
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// RequestState is the GORM model for domain.RequestState
type RequestState struct {
	RequestID    string `gorm:"primaryKey;size:64"`
	ActionedByID string `gorm:"primaryKey;size:64"`
	Action       string `gorm:"size:20"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (m *RequestState) ToDomain() *domain.RequestState {
	return &domain.RequestState{
		RequestID:    m.RequestID,
		ActionedByID: m.ActionedByID,
		Action:       domain.ActionStatus(m.Action),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func RequestStateFromDomain(s *domain.RequestState) *RequestState {
	return &RequestState{
		RequestID:    s.RequestID,
		ActionedByID: s.ActionedByID,
		Action:       string(s.Action),
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}
