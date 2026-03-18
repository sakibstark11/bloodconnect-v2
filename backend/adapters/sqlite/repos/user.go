package repos

import (
	"context"
	"errors"

	"bloodconnect/adapters/sqlite/models"
	"bloodconnect/application"
	"bloodconnect/application/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) application.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(models.UserFromDomain(user)).Error
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var m models.User
	res := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return m.ToDomain(), nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m models.User
	res := r.db.WithContext(ctx).Where("email = ?", email).First(&m)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return m.ToDomain(), nil
}

func (r *userRepository) UpdateUserHealth(ctx context.Context, health *domain.UserHealth) error {
	m := models.UserHealthFromDomain(health)
	return r.db.WithContext(ctx).Save(&m).Error
}

func (r *userRepository) GetUserHealth(ctx context.Context, userID string) ([]domain.UserHealth, error) {
	var ms []models.UserHealth
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]domain.UserHealth, len(ms))
	for i, m := range ms {
		result[i] = m.ToDomain()
	}
	return result, nil
}

func (r *userRepository) UpdateUserLocation(ctx context.Context, loc *domain.UserPreferredDonationLocation) error {
	return r.db.WithContext(ctx).Save(models.UserLocationFromDomain(loc)).Error
}

func (r *userRepository) GetUserLocation(ctx context.Context, userID string) (*domain.UserPreferredDonationLocation, error) {
	var m models.UserLocation
	res := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return m.ToDomain(), nil
}

func (r *userRepository) GetEligibleUsersInHexes(ctx context.Context, hexes []string, bloodType string, count int) ([]domain.User, error) {
	var userIDs []string
	if err := r.db.WithContext(ctx).Model(&models.UserLocation{}).Where("h3_hex IN ?", hexes).Limit(count).Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		return []domain.User{}, nil
	}

	var eligibleUserIDs []string
	if err := r.db.WithContext(ctx).Model(&models.UserHealth{}).
		Where("user_id IN ? AND info_type = ? AND details = ?", userIDs, string(domain.InfoTypeBloodType), bloodType).
		Pluck("user_id", &eligibleUserIDs).Error; err != nil {
		return nil, err
	}
	if len(eligibleUserIDs) == 0 {
		return []domain.User{}, nil
	}

	var ms []models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", eligibleUserIDs).Find(&ms).Error; err != nil {
		return nil, err
	}
	users := make([]domain.User, len(ms))
	for i, m := range ms {
		users[i] = *m.ToDomain()
	}
	return users, nil
}
