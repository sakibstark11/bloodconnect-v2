package repos

import (
	"context"

	"bloodconnect/adapters/sqlite/models"
	"bloodconnect/application"
	"bloodconnect/application/domain"
	"gorm.io/gorm"
)

type requestRepository struct {
	db *gorm.DB
}

func NewRequestRepository(db *gorm.DB) application.RequestRepository {
	return &requestRepository{db: db}
}

func (r *requestRepository) CreateRequest(ctx context.Context, req *domain.DonationRequest) error {
	return r.db.WithContext(ctx).Create(models.RequestFromDomain(req)).Error
}

func (r *requestRepository) UpdateRequest(ctx context.Context, req *domain.DonationRequest) error {
	return r.db.WithContext(ctx).Save(models.RequestFromDomain(req)).Error
}

func (r *requestRepository) GetRequestByID(ctx context.Context, id string) (*domain.DonationRequest, error) {
	var m models.Request
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *requestRepository) ListRequests(ctx context.Context, filters application.RequestFilters, page, pageSize int) ([]domain.DonationRequest, int, error) {
	q := r.db.WithContext(ctx).Model(&models.Request{})

	if filters.BloodType != "" {
		q = q.Where("blood_type = ?", filters.BloodType)
	}
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.LocationHex != "" {
		q = q.Where("location_hex = ?", filters.LocationHex)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var ms []models.Request
	if err := q.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&ms).Error; err != nil {
		return nil, 0, err
	}

	requests := make([]domain.DonationRequest, len(ms))
	for i, m := range ms {
		requests[i] = *m.ToDomain()
	}
	return requests, int(total), nil
}

func (r *requestRepository) SaveRequestState(ctx context.Context, state *domain.RequestState) error {
	return r.db.WithContext(ctx).Save(models.RequestStateFromDomain(state)).Error
}

func (r *requestRepository) GetRequestState(ctx context.Context, requestID, actionedByID string) (*domain.RequestState, error) {
	var m models.RequestState
	if err := r.db.WithContext(ctx).Where("request_id = ? AND actioned_by_id = ?", requestID, actionedByID).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *requestRepository) GetActionedUsers(ctx context.Context, requestID string) ([]domain.RequestState, error) {
	var ms []models.RequestState
	if err := r.db.WithContext(ctx).Where("request_id = ?", requestID).Find(&ms).Error; err != nil {
		return nil, err
	}
	states := make([]domain.RequestState, len(ms))
	for i, m := range ms {
		states[i] = *m.ToDomain()
	}
	return states, nil
}
