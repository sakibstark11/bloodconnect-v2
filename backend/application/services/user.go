package services

import (
	"context"
	"errors"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"
	"github.com/oklog/ulid/v2"
	"github.com/uber/h3-go/v4"
)

type UserService interface {
	Signup(ctx context.Context, name, email, password, phone string) (string, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	GetMe(ctx context.Context) (*domain.User, error)
	UpdateHealth(ctx context.Context, infoType domain.InfoType, details string) error
	UpdateLocation(ctx context.Context, lat, lng float64) error
}

type userService struct {
	repo   application.UserRepository
	config *application.AppConfig
}

func NewUserService(repo application.UserRepository, config *application.AppConfig) UserService {
	return &userService{repo: repo, config: config}
}

func (s *userService) Signup(ctx context.Context, name, email, password, phone string) (string, error) {
	id := "user_" + ulid.Make().String()

	user := &domain.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  password, // TODO: Hash password
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

// GetMe returns the currently authenticated user (ID read from context).
func (s *userService) GetMe(ctx context.Context) (*domain.User, error) {
	userID, _ := ctx.Value(domain.UserIDKey).(string)
	if userID == "" {
		return nil, errors.New("unauthorized")
	}
	return s.repo.GetUserByID(ctx, userID)
}

// UpdateHealth updates health info for the currently authenticated user.
func (s *userService) UpdateHealth(ctx context.Context, infoType domain.InfoType, details string) error {
	userID, _ := ctx.Value(domain.UserIDKey).(string)

	health := &domain.UserHealth{
		UserID:    userID,
		InfoType:  infoType,
		Details:   details,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.repo.UpdateUserHealth(ctx, health)
}

// UpdateLocation updates the preferred donation location for the currently authenticated user.
func (s *userService) UpdateLocation(ctx context.Context, lat, lng float64) error {
	userID, _ := ctx.Value(domain.UserIDKey).(string)

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
