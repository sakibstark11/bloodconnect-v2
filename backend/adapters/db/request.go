package repos

import (
	"context"

	"bloodconnect/adapters/models"
	"bloodconnect/application"
	"bloodconnect/application/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *requestRepository) GetRequestByID(ctx context.Context, id domain.RequestID) (*domain.DonationRequest, error) {
	var m models.Request
	if err := r.db.WithContext(ctx).Where("id = ?", string(id)).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *requestRepository) ListRequests(ctx context.Context, filters application.RequestFilters, lastRequestID domain.RequestID, pageSize int) ([]domain.DonationRequest, error) {
	q := r.db.WithContext(ctx).Model(&models.Request{})

	if filters.BloodType != "" {
		q = q.Where("blood_type = ?", string(filters.BloodType))
	}
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.LocationHex != "" {
		q = q.Where("location_hex = ?", filters.LocationHex)
	}

	if lastRequestID != "" {
		q = q.Where("id < ?", string(lastRequestID))
	}

	var ms []models.Request
	if err := q.Order("id desc").Limit(pageSize).Find(&ms).Error; err != nil {
		return nil, err
	}

	requests := make([]domain.DonationRequest, len(ms))
	for i, m := range ms {
		requests[i] = *m.ToDomain()
	}
	return requests, nil
}

func (r *requestRepository) SaveRequestState(ctx context.Context, state *domain.RequestState) error {
	m := models.RequestStateFromDomain(state)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "request_id"}, {Name: "actioned_by_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"action", "updated_at"}),
	}).Create(m).Error
}

func (r *requestRepository) GetRequestState(ctx context.Context, requestID domain.RequestID, actionedByID domain.UserID) (*domain.RequestState, error) {
	var m models.RequestState
	if err := r.db.WithContext(ctx).Where("request_id = ? AND actioned_by_id = ?", string(requestID), string(actionedByID)).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *requestRepository) GetActionedUsers(ctx context.Context, requestID domain.RequestID) ([]domain.RequestState, error) {
	var ms []models.RequestState
	if err := r.db.WithContext(ctx).Where("request_id = ?", string(requestID)).Find(&ms).Error; err != nil {
		return nil, err
	}
	states := make([]domain.RequestState, len(ms))
	for i, m := range ms {
		states[i] = *m.ToDomain()
	}
	return states, nil
}

func (r *requestRepository) GetUserRecentActions(ctx context.Context, userID domain.UserID, actions []domain.ActionStatus, since domain.ISOTimestamp) ([]domain.RequestState, error) {
	var ms []models.RequestState
	if err := r.db.WithContext(ctx).Where("actioned_by_id = ? AND action IN ? AND updated_at >= ?", string(userID), actions, since).Find(&ms).Error; err != nil {
		return nil, err
	}
	states := make([]domain.RequestState, len(ms))
	for i, m := range ms {
		states[i] = *m.ToDomain()
	}
	return states, nil
}
