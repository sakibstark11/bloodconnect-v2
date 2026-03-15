package ports

import (
	"context"

	"github.com/sakibalam/bloodconnect/domain"
)

type RequestRepository interface {
	CreateRequest(ctx context.Context, req *domain.DonationRequest) error
	UpdateRequest(ctx context.Context, req *domain.DonationRequest) error
	GetRequestByID(ctx context.Context, id string) (*domain.DonationRequest, error)

	SaveRequestState(ctx context.Context, state *domain.RequestState) error
	GetRequestState(ctx context.Context, requestID, actionedByID string) (*domain.RequestState, error)
	GetActionedUsers(ctx context.Context, requestID string) ([]domain.RequestState, error) // Returns UserIDs already pinged
}
