package memory

import (
	"context"
	"errors"
	"sync"

	"bloodconnect/application"
	"bloodconnect/application/domain"
)

type InMemoryUserRepository struct {
	mu        sync.RWMutex
	users     map[domain.UserID]domain.User
	auths     map[string]domain.UserAuth
	healths   map[domain.UserID][]domain.UserHealth
	locations map[domain.UserID]domain.UserPreferredDonationLocation
}

func NewUserRepository() application.UserRepository {
	return &InMemoryUserRepository{
		users:     make(map[domain.UserID]domain.User),
		auths:     make(map[string]domain.UserAuth),
		healths:   make(map[domain.UserID][]domain.UserHealth),
		locations: make(map[domain.UserID]domain.UserPreferredDonationLocation),
	}
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, user *domain.User, hashedPassword string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.auths[user.Email]; ok {
		return errors.New("user already exists")
	}

	r.users[user.ID] = *user
	r.auths[user.Email] = domain.UserAuth{
		UserID:   user.ID,
		Password: hashedPassword,
	}
	return nil
}

func (r *InMemoryUserRepository) GetUserByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return &user, nil
}

func (r *InMemoryUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	auth, ok := r.auths[email]
	if !ok {
		return nil, nil
	}
	user := r.users[auth.UserID]
	return &user, nil
}

func (r *InMemoryUserRepository) GetUserAuthByEmail(ctx context.Context, email string) (*domain.UserAuth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	auth, ok := r.auths[email]
	if !ok {
		return nil, nil
	}
	return &auth, nil
}

func (r *InMemoryUserRepository) UpdateUserHealth(ctx context.Context, health *domain.UserHealth) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.healths[health.UserID] = append(r.healths[health.UserID], *health)
	return nil
}

func (r *InMemoryUserRepository) GetUserHealth(ctx context.Context, userID domain.UserID) ([]domain.UserHealth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.healths[userID], nil
}

func (r *InMemoryUserRepository) UpdateUserLocation(ctx context.Context, loc *domain.UserPreferredDonationLocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.locations[loc.UserID] = *loc
	return nil
}

func (r *InMemoryUserRepository) GetUserLocation(ctx context.Context, userID domain.UserID) (*domain.UserPreferredDonationLocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	loc, ok := r.locations[userID]
	if !ok {
		return nil, nil
	}
	return &loc, nil
}

func (r *InMemoryUserRepository) GetEligibleUsersInHexes(ctx context.Context, hexes []string, bloodType domain.BloodType, count int, excludedUserIDs []domain.UserID) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hexMap := make(map[string]bool)
	for _, h := range hexes {
		hexMap[h] = true
	}

	excluded := make(map[domain.UserID]bool)
	for _, id := range excludedUserIDs {
		excluded[id] = true
	}

	var eligible []domain.User
	for userID, loc := range r.locations {
		if excluded[userID] {
			continue
		}
		if hexMap[loc.H3Hex] {

			healths := r.healths[userID]
			matches := false
			if bloodType == "" {
				matches = true
			} else {
				for _, h := range healths {
					if h.InfoType == domain.InfoTypeBloodType && h.Details == string(bloodType) {
						matches = true
						break
					}
				}
			}

			if matches {
				user := r.users[userID]
				eligible = append(eligible, user)
				if len(eligible) >= count {
					break
				}
			}
		}
	}
	return eligible, nil
}
