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
	"go.uber.org/zap"
)

type RequestService interface {
	SubmitRequest(ctx context.Context, userID, bloodType, phone, desc, reqInfo, locName string, lat, lng float64, count int, requiredBy time.Time) (string, error)
	RespondToRequest(ctx context.Context, requestID, userID string, action domain.ActionStatus) error
	CancelRequest(ctx context.Context, requestID, userID string) error
	GetRequest(ctx context.Context, requestID string) (*domain.DonationRequest, error)
	GetExtendedRequest(ctx context.Context, requestID string) (*domain.ExtendedDonationRequest, error)
}

type requestService struct {
	repo         ports.RequestRepository
	userRepo     ports.UserRepository
	queue        ports.JobQueue
	notifService NotificationService
	config       *AppConfig
	logger       *zap.Logger
}

func NewRequestService(repo ports.RequestRepository, userRepo ports.UserRepository, queue ports.JobQueue, notifService NotificationService, config *AppConfig, logger *zap.Logger) RequestService {
	return &requestService{
		repo:         repo,
		userRepo:     userRepo,
		queue:        queue,
		notifService: notifService,
		config:       config,
		logger:       logger.With(zap.String("service", "RequestService")),
	}
}

func (s *requestService) getReqLogger(ctx context.Context, requestID string) *zap.Logger {
	traceID, _ := ctx.Value(domain.TraceIDKey).(string)
	l := s.logger
	if traceID != "" {
		l = l.With(zap.String("trace_id", traceID))
	}
	if requestID != "" {
		l = l.With(zap.String("request_id", requestID))
	}
	return l
}

func (s *requestService) GetRequest(ctx context.Context, requestID string) (*domain.DonationRequest, error) {
	return s.repo.GetRequestByID(ctx, requestID)
}

func (s *requestService) GetExtendedRequest(ctx context.Context, requestID string) (*domain.ExtendedDonationRequest, error) {
	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, nil
	}

	actionedStates, err := s.repo.GetActionedUsers(ctx, requestID)
	if err != nil {
		return nil, err
	}

	var notifiedUsers []domain.RequestActionedUser
	for _, state := range actionedStates {
		loc, err := s.userRepo.GetUserLocation(ctx, state.ActionedByID)
		if err != nil || loc == nil {
			continue // Skip if user location is somehow missing
		}
		notifiedUsers = append(notifiedUsers, domain.RequestActionedUser{
			UserID: state.ActionedByID,
			Action: state.Action,
			Lat:    loc.Lat,
			Lng:    loc.Lng,
			H3Hex:  loc.H3Hex,
		})
	}

	return &domain.ExtendedDonationRequest{
		Request:       req,
		NotifiedUsers: notifiedUsers,
	}, nil
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
	reqLogger := s.getReqLogger(ctx, reqID)
	
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
		reqLogger.Error("Failed to enqueue search wave job", zap.Error(err))
		return "", err
	}

	reqLogger.Info("Successfully submitted request",
		zap.Int("bag_count", count),
		zap.String("blood_type", bloodType),
	)
	return reqID, nil
}

func (s *requestService) RespondToRequest(ctx context.Context, requestID, userID string, action domain.ActionStatus) error {
	reqLogger := s.getReqLogger(ctx, requestID)
	state := &domain.RequestState{
		RequestID:    requestID,
		ActionedByID: userID,
		Action:       action,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.SaveRequestState(ctx, state); err != nil {
		reqLogger.Error("Failed to save request state", zap.String("actioned_user_id", userID), zap.Error(err))
		return err
	}

	if action == domain.ActionStatusAccepted {
		reqLogger.Info("User accepted donation request", zap.String("actioned_user_id", userID))
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
