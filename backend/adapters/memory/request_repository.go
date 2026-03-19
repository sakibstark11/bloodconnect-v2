package memory

import (
	"context"
	"sync"

	"bloodconnect/application"
	"bloodconnect/application/domain"
)

type InMemoryRequestRepository struct {
	mu       sync.RWMutex
	requests map[domain.RequestID]domain.DonationRequest
	states   map[domain.RequestID][]domain.RequestState
}

func NewRequestRepository() application.RequestRepository {
	return &InMemoryRequestRepository{
		requests: make(map[domain.RequestID]domain.DonationRequest),
		states:   make(map[domain.RequestID][]domain.RequestState),
	}
}

func (r *InMemoryRequestRepository) CreateRequest(ctx context.Context, req *domain.DonationRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[req.ID] = *req
	return nil
}

func (r *InMemoryRequestRepository) UpdateRequest(ctx context.Context, req *domain.DonationRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[req.ID] = *req
	return nil
}

func (r *InMemoryRequestRepository) GetRequestByID(ctx context.Context, id domain.RequestID) (*domain.DonationRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	req, ok := r.requests[id]
	if !ok {
		return nil, nil
	}
	return &req, nil
}

func (r *InMemoryRequestRepository) ListRequests(ctx context.Context, filters application.RequestFilters, page, pageSize int) ([]domain.DonationRequest, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.DonationRequest
	for _, req := range r.requests {
		if filters.BloodType != "" && req.BloodType != filters.BloodType {
			continue
		}
		if filters.Status != "" && string(req.Status) != string(filters.Status) {
			continue
		}
		if filters.LocationHex != "" && req.LocationHex != filters.LocationHex {
			continue
		}
		filtered = append(filtered, req)
	}

	total := len(filtered)
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= total {
		return []domain.DonationRequest{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (r *InMemoryRequestRepository) SaveRequestState(ctx context.Context, state *domain.RequestState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	states := r.states[state.RequestID]
	found := false
	for i, s := range states {
		if s.ActionedByID == state.ActionedByID {
			states[i] = *state
			found = true
			break
		}
	}
	if !found {
		states = append(states, *state)
	}
	r.states[state.RequestID] = states
	return nil
}

func (r *InMemoryRequestRepository) GetRequestState(ctx context.Context, requestID domain.RequestID, actionedByID domain.UserID) (*domain.RequestState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	states := r.states[requestID]
	for _, s := range states {
		if s.ActionedByID == actionedByID {
			return &s, nil
		}
	}
	return nil, nil
}

func (r *InMemoryRequestRepository) GetActionedUsers(ctx context.Context, requestID domain.RequestID) ([]domain.RequestState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.states[requestID], nil
}

func (r *InMemoryRequestRepository) GetUserRecentActions(ctx context.Context, userID domain.UserID, action domain.ActionStatus, since domain.ISOTimestamp) ([]domain.RequestState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var recent []domain.RequestState
	for _, states := range r.states {
		for _, s := range states {
			if s.ActionedByID == userID && s.Action == action && s.UpdatedAt >= since {
				recent = append(recent, s)
			}
		}
	}
	return recent, nil
}
