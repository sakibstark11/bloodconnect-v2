package models

import (
	"bloodconnect/application/domain"
)

type User struct {
	ID        string `gorm:"primaryKey;size:64"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	Password  string
	Phone     string `gorm:"uniqueIndex"`
	CreatedAt string `gorm:"size:25"`
	UpdatedAt string `gorm:"size:25"`
}



type Profile struct {
	ID        string `gorm:"primaryKey;size:64"`
	Name      string
	Email     string
	Phone     string
	CreatedAt string `gorm:"size:25"`
	UpdatedAt string `gorm:"size:25"`
}



type Auth struct {
	ID       string `gorm:"primaryKey;column:id"`
	Password string `gorm:"column:password"`
}



func (m *Profile) ToDomain() *domain.User {
	return &domain.User{
		ID:        domain.UserID(m.ID),
		Name:      m.Name,
		Email:     m.Email,
		Phone:     m.Phone,
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
		UpdatedAt: domain.ISOTimestamp(m.UpdatedAt),
	}
}

func (m *Auth) ToAuth() *domain.UserAuth {
	return &domain.UserAuth{
		UserID:   domain.UserID(m.ID),
		Password: m.Password,
	}
}

func UserFromDomain(u *domain.User, hashedPassword string) *User {
	return &User{
		ID:        string(u.ID),
		Name:      u.Name,
		Email:     u.Email,
		Password:  hashedPassword,
		Phone:     u.Phone,
		CreatedAt: string(u.CreatedAt),
		UpdatedAt: string(u.UpdatedAt),
	}
}

type UserHealth struct {
	UserID    string `gorm:"primaryKey;size:64"`
	InfoType  string `gorm:"primaryKey;size:50"`
	Details   string
	CreatedAt string `gorm:"size:25"`
	UpdatedAt string `gorm:"size:25"`
}

func (m *UserHealth) ToDomain() domain.UserHealth {
	return domain.UserHealth{
		UserID:    domain.UserID(m.UserID),
		InfoType:  domain.InfoType(m.InfoType),
		Details:   m.Details,
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
		UpdatedAt: domain.ISOTimestamp(m.UpdatedAt),
	}
}

func UserHealthFromDomain(h *domain.UserHealth) UserHealth {
	return UserHealth{
		UserID:    string(h.UserID),
		InfoType:  string(h.InfoType),
		Details:   h.Details,
		CreatedAt: string(h.CreatedAt),
		UpdatedAt: string(h.UpdatedAt),
	}
}

type UserLocation struct {
	UserID    string `gorm:"primaryKey;size:64"`
	Lat       float64
	Lng       float64
	H3Hex     string `gorm:"index;size:20"`
	CreatedAt string `gorm:"size:25"`
	UpdatedAt string `gorm:"size:25"`
}

func (m *UserLocation) ToDomain() *domain.UserPreferredDonationLocation {
	return &domain.UserPreferredDonationLocation{
		UserID:    domain.UserID(m.UserID),
		Lat:       m.Lat,
		Lng:       m.Lng,
		H3Hex:     m.H3Hex,
		CreatedAt: domain.ISOTimestamp(m.CreatedAt),
		UpdatedAt: domain.ISOTimestamp(m.UpdatedAt),
	}
}

func UserLocationFromDomain(l *domain.UserPreferredDonationLocation) *UserLocation {
	return &UserLocation{
		UserID:    string(l.UserID),
		Lat:       l.Lat,
		Lng:       l.Lng,
		H3Hex:     l.H3Hex,
		CreatedAt: string(l.CreatedAt),
		UpdatedAt: string(l.UpdatedAt),
	}
}
