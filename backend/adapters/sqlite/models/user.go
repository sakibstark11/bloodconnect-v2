package models

import (
	"time"

	"bloodconnect/application/domain"
)

// User is the GORM model for domain.User
type User struct {
	ID        string `gorm:"primaryKey;size:64"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	Password  string
	Phone     string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *User) ToDomain() *domain.User {
	return &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		Phone:     m.Phone,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserFromDomain(u *domain.User) *User {
	return &User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserHealth is the GORM model for domain.UserHealth
type UserHealth struct {
	UserID    string `gorm:"primaryKey;size:64"`
	InfoType  string `gorm:"primaryKey;size:50"`
	Details   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *UserHealth) ToDomain() domain.UserHealth {
	return domain.UserHealth{
		UserID:    m.UserID,
		InfoType:  domain.InfoType(m.InfoType),
		Details:   m.Details,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserHealthFromDomain(h *domain.UserHealth) UserHealth {
	return UserHealth{
		UserID:    h.UserID,
		InfoType:  string(h.InfoType),
		Details:   h.Details,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
}

// UserLocation is the GORM model for domain.UserPreferredDonationLocation
type UserLocation struct {
	UserID    string `gorm:"primaryKey;size:64"`
	Lat       float64
	Lng       float64
	H3Hex     string `gorm:"index;size:20"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *UserLocation) ToDomain() *domain.UserPreferredDonationLocation {
	return &domain.UserPreferredDonationLocation{
		UserID:    m.UserID,
		Lat:       m.Lat,
		Lng:       m.Lng,
		H3Hex:     m.H3Hex,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserLocationFromDomain(l *domain.UserPreferredDonationLocation) *UserLocation {
	return &UserLocation{
		UserID:    l.UserID,
		Lat:       l.Lat,
		Lng:       l.Lng,
		H3Hex:     l.H3Hex,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
