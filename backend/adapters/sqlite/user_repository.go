package sqlite

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new SQLite user repository
func NewUserRepository(db *gorm.DB) ports.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	m := fromDomainUser(user)
	res := r.db.WithContext(ctx).Create(m)
	return res.Error
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var m userModel
	res := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Or a specific domain error
		}
		return nil, res.Error
	}
	return m.toDomain(), nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	res := r.db.WithContext(ctx).Where("email = ?", email).First(&m)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return m.toDomain(), nil
}

func (r *userRepository) UpdateUserHealth(ctx context.Context, health *domain.UserHealth) error {
	m := fromDomainUserHealth(health)
	// Use Save to Insert or Update based on Primary Keys (UserID + InfoType)
	res := r.db.WithContext(ctx).Save(&m)
	return res.Error
}

func (r *userRepository) GetUserHealth(ctx context.Context, userID string) ([]domain.UserHealth, error) {
	var models []userHealthModel
	res := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models)
	if res.Error != nil {
		return nil, res.Error
	}

	var d []domain.UserHealth
	for _, m := range models {
		d = append(d, m.toDomain())
	}
	return d, nil
}

func (r *userRepository) UpdateUserLocation(ctx context.Context, loc *domain.UserPreferredDonationLocation) error {
	m := fromDomainUserLocation(loc)
	// We only expect one location per user, Save on PK
	res := r.db.WithContext(ctx).Save(m)
	return res.Error
}

func (r *userRepository) GetUserLocation(ctx context.Context, userID string) (*domain.UserPreferredDonationLocation, error) {
	var m userLocationModel
	res := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return m.toDomain(), nil
}

func (r *userRepository) GetEligibleUsersInHexes(ctx context.Context, hexes []string, bloodType string, count int) ([]domain.User, error) {
	// 1. Find UserIDs in Location table where H3Hex is in `hexes`
	var userIDs []string
	res := r.db.WithContext(ctx).Model(&userLocationModel{}).Where("h3_hex IN ?", hexes).Limit(count).Pluck("user_id", &userIDs)
	if res.Error != nil {
		return nil, res.Error
	}

	if len(userIDs) == 0 {
		return []domain.User{}, nil
	}

	// 2. Filter these UserIDs by those who have matching BloodType in UserHealth
	var eligibleUserIDs []string
	res = r.db.WithContext(ctx).Model(&userHealthModel{}).
		Where("user_id IN ? AND info_type = ? AND details = ?", userIDs, string(domain.InfoTypeBloodType), bloodType).
		Pluck("user_id", &eligibleUserIDs)
	if res.Error != nil {
		return nil, res.Error
	}

	if len(eligibleUserIDs) == 0 {
		return []domain.User{}, nil
	}

	// 3. Fetch the actual user models
	var models []userModel
	res = r.db.WithContext(ctx).Where("id IN ?", eligibleUserIDs).Find(&models)
	if res.Error != nil {
		return nil, res.Error
	}

	var users []domain.User
	for _, m := range models {
		users = append(users, *m.toDomain())
	}
	return users, nil
}
