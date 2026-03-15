package ports

import (
	"context"

	"github.com/sakibalam/bloodconnect/domain"
)

// UserRepository defines the interface for interacting with User data storage
type UserRepository interface {
	// User operations
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)

	// Health operations
	UpdateUserHealth(ctx context.Context, health *domain.UserHealth) error
	GetUserHealth(ctx context.Context, userID string) ([]domain.UserHealth, error)

	// Location operations
	UpdateUserLocation(ctx context.Context, loc *domain.UserPreferredDonationLocation) error
	GetUserLocation(ctx context.Context, userID string) (*domain.UserPreferredDonationLocation, error)

	// Search operations
	GetEligibleUsersInHexes(ctx context.Context, hexes []string, bloodType string, count int) ([]domain.User, error)
}
