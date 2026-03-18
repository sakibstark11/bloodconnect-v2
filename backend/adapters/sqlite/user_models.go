package sqlite

import (
	"time"

	"github.com/sakibalam/bloodconnect/domain"
)

// userModel is the SQLite representation of domain.User
type userModel struct {
	ID        string `gorm:"primaryKey;size:64"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	Password  string
	Phone     string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *userModel) toDomain() *domain.User {
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

func fromDomainUser(u *domain.User) *userModel {
	return &userModel{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// userHealthModel is the SQLite representation of domain.UserHealth
type userHealthModel struct {
	UserID    string `gorm:"primaryKey;size:64"`
	InfoType  string `gorm:"primaryKey;size:50"`
	Details   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *userHealthModel) toDomain() domain.UserHealth {
	return domain.UserHealth{
		UserID:    m.UserID,
		InfoType:  domain.InfoType(m.InfoType),
		Details:   m.Details,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromDomainUserHealth(h *domain.UserHealth) userHealthModel {
	return userHealthModel{
		UserID:    h.UserID,
		InfoType:  string(h.InfoType),
		Details:   h.Details,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
}

// userLocationModel is the SQLite representation of domain.UserPreferredDonationLocation
type userLocationModel struct {
	UserID    string `gorm:"primaryKey;size:64"`
	Lat       float64
	Lng       float64
	H3Hex     string `gorm:"index;size:20"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *userLocationModel) toDomain() *domain.UserPreferredDonationLocation {
	return &domain.UserPreferredDonationLocation{
		UserID:    m.UserID,
		Lat:       m.Lat,
		Lng:       m.Lng,
		H3Hex:     m.H3Hex,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromDomainUserLocation(l *domain.UserPreferredDonationLocation) *userLocationModel {
	return &userLocationModel{
		UserID:    l.UserID,
		Lat:       l.Lat,
		Lng:       l.Lng,
		H3Hex:     l.H3Hex,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
