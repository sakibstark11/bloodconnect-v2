package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
	"github.com/uber/h3-go/v4"
)

type RequestService interface {
	SubmitRequest(ctx context.Context, userID, bloodType, phone, desc, reqInfo, locName string, lat, lng float64, count int, requiredBy time.Time) (string, error)
	RespondToRequest(ctx context.Context, requestID, userID string, action domain.ActionStatus) error
	CancelRequest(ctx context.Context, requestID, userID string) error
	GetRequest(ctx context.Context, requestID string) (*domain.DonationRequest, error)
}

type requestService struct {
	repo         ports.RequestRepository
	queue        ports.JobQueue
	notifService NotificationService
	config       *AppConfig
}

func NewRequestService(repo ports.RequestRepository, queue ports.JobQueue, notifService NotificationService, config *AppConfig) RequestService {
	return &requestService{
		repo:         repo,
		queue:        queue,
		notifService: notifService,
		config:       config,
	}
}

func (s *requestService) GetRequest(ctx context.Context, requestID string) (*domain.DonationRequest, error) {
	return s.repo.GetRequestByID(ctx, requestID)
}

func (s *requestService) CancelRequest(ctx context.Context, requestID, userID string) error {
	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Ensure the user trying to cancel is the requester
	if req.UserID != userID {
		return errors.New("unauthorized to cancel this request")
	}

	req.Status = domain.RequestStatusCancelled
	req.UpdatedAt = time.Now()

	return s.repo.UpdateRequest(ctx, req)
}

func (s *requestService) SubmitRequest(ctx context.Context, userID, bloodType, phone, desc, reqInfo, locName string, lat, lng float64, count int, requiredBy time.Time) (string, error) {
	cell, _ := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, s.config.H3HexResolution)
	hex := cell.String()
	reqID := "request_" + ulid.Make().String()
	req := &domain.DonationRequest{
		ID:             reqID,
		UserID:         userID,
		LocationHex:    hex,
		LocationLat:    lat,
		LocationLng:    lng,
		BagCount:       count,
		RequiredByDate: requiredBy,
		BloodType:      domain.BloodType(bloodType),
		ContactPhone:   phone,
		Description:    desc,
		RequesterInfo:  reqInfo,
		LocationName:   locName,
		Status:         domain.RequestStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return "", err
	}

	// Calculate job run time
	now := time.Now()
	runAt := now
	daysUntilReq := requiredBy.Sub(now).Hours() / 24.0

	if daysUntilReq > float64(s.config.ProcessRequestWindowDays) {
		// Delay until exactly N days before requirement
		runAt = requiredBy.Add(-time.Duration(s.config.ProcessRequestWindowDays) * 24 * time.Hour)
	}

	// Enqueue search wave job
	payload := domain.WaveSearchPayload{
		RequestID:         reqID,
		CurrentRing:       1,
		CenterHex:         hex,
		UsersLeftToSearch: count,
	}
	payloadBytes, _ := json.Marshal(payload)

	job := &domain.Job{
		ID:        "job_" + ulid.Make().String(),
		Type:      domain.JobTypeWaveSearch,
		Payload:   string(payloadBytes),
		Status:    domain.JobStatusPending,
		RunAt:     runAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return "", err
	}

	return reqID, nil
}

func (s *requestService) RespondToRequest(ctx context.Context, requestID, userID string, action domain.ActionStatus) error {
	state := &domain.RequestState{
		RequestID:    requestID,
		ActionedByID: userID,
		Action:       action,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.SaveRequestState(ctx, state); err != nil {
		return err
	}

	if action == domain.ActionStatusAccepted {
		req, err := s.repo.GetRequestByID(ctx, requestID)
		if err == nil && req != nil {
			// Notify the original requester
			title := "Donation Request Accepted"
			content := "A user has accepted your request to donate blood."
			_, _ = s.notifService.Submit(ctx, domain.NotificationTypeDonationRequestAcceptance, req.UserID, title, content)
		}
	}

	return nil
}
