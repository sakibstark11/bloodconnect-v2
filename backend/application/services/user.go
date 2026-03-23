package services

import (
	"context"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/uber/h3-go/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Signup(ctx context.Context, name, email, password, phone string) (domain.UserID, error)
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
	GetMe(ctx context.Context, userID domain.UserID) (*domain.User, []domain.UserHealth, error)
	UpdateHealth(ctx context.Context, userID domain.UserID, infoType domain.InfoType, details string) error
	UpdateLocation(ctx context.Context, userID domain.UserID, lat, lng float64) error
}

type userService struct {
	repo   application.UserRepository
	config *application.AppConfig
}

func NewUserService(repo application.UserRepository, config *application.AppConfig) UserService {
	return &userService{repo: repo, config: config}
}

func (s *userService) Signup(ctx context.Context, name, email, password, phone string) (domain.UserID, error) {
	// fail fast: email already exists
	existing, _ := s.repo.GetUserAuthByEmail(ctx, email)
	if existing != nil {
		return "", domain.ErrEmailAlreadyInUse
	}

	id := domain.UserID("user_" + ulid.Make().String())

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &domain.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: domain.Now(),
		UpdatedAt: domain.Now(),
	}

	if err := s.repo.CreateUser(ctx, user, string(hashedPassword)); err != nil {
		return "", err
	}

	return id, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {

	userAuth, err := s.repo.GetUserAuthByEmail(ctx, email)
	switch {
	case err != nil:
		return "", nil, err
	case userAuth == nil:
		return "", nil, domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userAuth.Password), []byte(password)); err != nil {
		return "", nil, domain.ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, userAuth.UserID)
	if err != nil {
		return "", nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *userService) generateToken(userID domain.UserID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": string(userID),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *userService) GetMe(ctx context.Context, userID domain.UserID) (*domain.User, []domain.UserHealth, error) {
	if userID == "" {
		return nil, nil, domain.ErrUnauthorized
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, domain.ErrUserNotFound
	}
	health, err := s.repo.GetUserHealth(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return user, health, nil
}

func (s *userService) UpdateHealth(ctx context.Context, userID domain.UserID, infoType domain.InfoType, details string) error {
	if userID == "" {
		return domain.ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if infoType == domain.InfoTypeBloodType {
		health, err := s.repo.GetUserHealth(ctx, userID)
		if err == nil {
			for _, h := range health {
				if h.InfoType == domain.InfoTypeBloodType && h.Details != "" {
					return domain.ErrBloodTypeUpdateDenied
				}
			}
		}
	}

	health := &domain.UserHealth{
		UserID:    userID,
		InfoType:  infoType,
		Details:   details,
		CreatedAt: domain.Now(),
		UpdatedAt: domain.Now(),
	}

	return s.repo.UpdateUserHealth(ctx, health)
}

func (s *userService) UpdateLocation(ctx context.Context, userID domain.UserID, lat, lng float64) error {
	if userID == "" {
		return domain.ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	cell, _ := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, s.config.H3HexResolution)
	loc := &domain.UserPreferredDonationLocation{
		UserID:    userID,
		Lat:       lat,
		Lng:       lng,
		H3Hex:     cell.String(),
		CreatedAt: domain.Now(),
		UpdatedAt: domain.Now(),
	}

	return s.repo.UpdateUserLocation(ctx, loc)
}
