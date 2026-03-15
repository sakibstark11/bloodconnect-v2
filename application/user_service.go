package application

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
	"github.com/uber/h3-go/v4"
)

type UserService interface {
	Signup(ctx context.Context, name, email, password, phone string) (string, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	UpdateHealth(ctx context.Context, userID string, infoType domain.InfoType, details string) error
	UpdateLocation(ctx context.Context, userID string, lat, lng float64) error
}

type userService struct {
	repo   ports.UserRepository
	config *AppConfig
}

func NewUserService(repo ports.UserRepository, config *AppConfig) UserService {
	return &userService{repo: repo, config: config}
}

func (s *userService) Signup(ctx context.Context, name, email, password, phone string) (string, error) {
	// Generate unique ID
	id := "user_" + ulid.Make().String()
	
	user := &domain.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  password, // TODO: Hash password here
		Phone:     phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return "", err
	}

	return id, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// TODO: verify hashed password
	if user.Password != password {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *userService) UpdateHealth(ctx context.Context, userID string, infoType domain.InfoType, details string) error {
	health := &domain.UserHealth{
		UserID:    userID,
		InfoType:  infoType,
		Details:   details,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.repo.UpdateUserHealth(ctx, health)
}

func (s *userService) UpdateLocation(ctx context.Context, userID string, lat, lng float64) error {
	cell, _ := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, s.config.H3HexResolution)
	loc := &domain.UserPreferredDonationLocation{
		UserID:    userID,
		Lat:       lat,
		Lng:       lng,
		H3Hex:     cell.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.repo.UpdateUserLocation(ctx, loc)
}
