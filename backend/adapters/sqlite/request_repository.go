package sqlite

import (
	"context"

	"gorm.io/gorm"

	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
)

type requestRepository struct {
	db *gorm.DB
}

func NewRequestRepository(db *gorm.DB) ports.RequestRepository {
	return &requestRepository{
		db: db,
	}
}

func (r *requestRepository) CreateRequest(ctx context.Context, req *domain.DonationRequest) error {
	m := fromDomainRequest(req)
	res := r.db.WithContext(ctx).Create(m)
	return res.Error
}

func (r *requestRepository) UpdateRequest(ctx context.Context, req *domain.DonationRequest) error {
	m := fromDomainRequest(req)
	res := r.db.WithContext(ctx).Save(m)
	return res.Error
}

func (r *requestRepository) GetRequestByID(ctx context.Context, id string) (*domain.DonationRequest, error) {
	var m requestModel
	res := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	return m.toDomain(), nil
}

func (r *requestRepository) SaveRequestState(ctx context.Context, state *domain.RequestState) error {
	m := fromDomainRequestState(state)
	res := r.db.WithContext(ctx).Save(m)
	return res.Error
}

func (r *requestRepository) GetRequestState(ctx context.Context, requestID, actionedByID string) (*domain.RequestState, error) {
	var m requestStateModel
	res := r.db.WithContext(ctx).Where("request_id = ? AND actioned_by_id = ?", requestID, actionedByID).First(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	return m.toDomain(), nil
}

func (r *requestRepository) GetActionedUsers(ctx context.Context, requestID string) ([]domain.RequestState, error) {
	var m []requestStateModel
	res := r.db.WithContext(ctx).Where("request_id = ?", requestID).Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	var states []domain.RequestState
	for _, state := range m {
		states = append(states, *state.toDomain())
	}
	return states, nil
}
