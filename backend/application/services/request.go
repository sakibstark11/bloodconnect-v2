package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"

	"github.com/oklog/ulid/v2"
	"github.com/uber/h3-go/v4"
	"go.uber.org/zap"
)

type RequestService interface {
	SubmitRequest(ctx context.Context, userID domain.UserID, bloodType domain.BloodType, desc, reqInfo, locName string, lat, lng float64, count int, requiredBy domain.ISOTimestamp) (domain.RequestID, error)
	RespondToRequest(ctx context.Context, requestID domain.RequestID, action domain.ActionStatus) error
	CancelRequest(ctx context.Context, requestID domain.RequestID) error
	GetRequest(ctx context.Context, requestID domain.RequestID) (*domain.DonationRequest, error)
	GetExtendedRequest(ctx context.Context, requestID domain.RequestID) (*domain.ExtendedDonationRequest, error)
	ListRequests(ctx context.Context, filters application.RequestFilters, lastRequestID domain.RequestID, pageSize int) ([]domain.DonationRequest, error)
}

type requestService struct {
	repo         application.RequestRepository
	userRepo     application.UserRepository
	queue        application.JobQueue
	notifService NotificationService
	config       *application.AppConfig
	logger       *zap.Logger
}

func NewRequestService(repo application.RequestRepository, userRepo application.UserRepository, queue application.JobQueue, notifService NotificationService, config *application.AppConfig, logger *zap.Logger) RequestService {
	return &requestService{
		repo:         repo,
		userRepo:     userRepo,
		queue:        queue,
		notifService: notifService,
		config:       config,
		logger:       logger.With(zap.String("service", "RequestService")),
	}
}

func (s *requestService) getReqLogger(ctx context.Context, requestID domain.RequestID) *zap.Logger {
	traceID, _ := ctx.Value(domain.TraceIDKey).(string)
	l := s.logger
	if traceID != "" {
		l = l.With(zap.String("trace_id", traceID))
	}
	if requestID != "" {
		l = l.With(zap.String("request_id", string(requestID)))
	}
	return l
}

func (s *requestService) ListRequests(ctx context.Context, filters application.RequestFilters, lastRequestID domain.RequestID, pageSize int) ([]domain.DonationRequest, error) {
	return s.repo.ListRequests(ctx, filters, lastRequestID, pageSize)
}

func (s *requestService) GetRequest(ctx context.Context, requestID domain.RequestID) (*domain.DonationRequest, error) {
	return s.repo.GetRequestByID(ctx, requestID)
}

func (s *requestService) GetExtendedRequest(ctx context.Context, requestID domain.RequestID) (*domain.ExtendedDonationRequest, error) {
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
			continue
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

func (s *requestService) CancelRequest(ctx context.Context, requestID domain.RequestID) error {
	userID, _ := ctx.Value(domain.UserIDKey).(domain.UserID)

	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	if req.UserID != userID {
		return errors.New("unauthorized to cancel this request")
	}

	req.Status = domain.RequestStatusCancelled
	req.UpdatedAt = domain.Now()

	return s.repo.UpdateRequest(ctx, req)
}

func (s *requestService) SubmitRequest(ctx context.Context, userID domain.UserID, bloodType domain.BloodType, desc, reqInfo, locName string, lat, lng float64, count int, requiredBy domain.ISOTimestamp) (domain.RequestID, error) {
	cell, _ := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, s.config.H3HexResolution)
	hex := cell.String()
	reqID := domain.RequestID("request_" + ulid.Make().String())
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
		Description:    desc,
		RequesterInfo:  reqInfo,
		LocationName:   locName,
		Status:         domain.RequestStatusPending,
		CreatedAt:      domain.Now(),
		UpdatedAt:      domain.Now(),
	}

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return "", err
	}

	now := domain.Now()
	runAt := now
	daysUntilReq := requiredBy.ToTime().Sub(now.ToTime()).Hours() / 24.0

	if daysUntilReq > float64(s.config.ProcessRequestWindowDays) {
		runAt = domain.ISOTimestamp(requiredBy.ToTime().Add(-time.Duration(s.config.ProcessRequestWindowDays) * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z"))
	}

	payload := domain.WaveSearchPayload{
		RequestID:   reqID,
		CurrentRing: 1,
	}
	payloadBytes, _ := json.Marshal(payload)

	job := &domain.Job{
		ID:        domain.JobID("job_" + ulid.Make().String()),
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
		zap.String("blood_type", string(bloodType)),
	)
	return reqID, nil
}

func (s *requestService) RespondToRequest(ctx context.Context, requestID domain.RequestID, action domain.ActionStatus) error {
	userID, _ := ctx.Value(domain.UserIDKey).(domain.UserID)
	reqLogger := s.getReqLogger(ctx, requestID)

	if action == domain.ActionStatusAccepted {
		since := time.Now().Add(-time.Duration(s.config.MinimumDonationWaitDays) * 24 * time.Hour)
		recentActions, err := s.repo.GetUserRecentActions(ctx, userID, domain.ActionStatusAccepted, domain.ISOTimestamp(since.Format("2006-01-02T15:04:05.000Z")))
		if err != nil {
			reqLogger.Error("Failed to check recent actions", zap.Error(err))
		} else {

			for _, ra := range recentActions {
				if ra.RequestID != requestID {
					return errors.New("you cannot accept another request within the minimum donation wait period")
				}
			}
		}
	}

	prevState, _ := s.repo.GetRequestState(ctx, requestID, userID)

	state := &domain.RequestState{
		RequestID:    requestID,
		ActionedByID: userID,
		Action:       action,
		CreatedAt:    domain.Now(),
		UpdatedAt:    domain.Now(),
	}
	if prevState != nil {
		state.CreatedAt = prevState.CreatedAt
	}

	if err := s.repo.SaveRequestState(ctx, state); err != nil {
		reqLogger.Error("Failed to save request state", zap.String("actioned_user_id", string(userID)), zap.Error(err))
		return err
	}

	if action == domain.ActionStatusAccepted {
		reqLogger.Info("User accepted donation request", zap.String("actioned_user_id", string(userID)))
		req, err := s.repo.GetRequestByID(ctx, requestID)
		if err == nil && req != nil {
			title := "Donation Request Accepted"
			content := "A user has accepted your request to donate blood."
			_, _ = s.notifService.Submit(ctx, domain.NotificationTypeDonationRequestAcceptance, req.UserID, title, content)
		}
	}

	if prevState != nil && prevState.Action == domain.ActionStatusAccepted && action == domain.ActionStatusDeclined {
		reqLogger.Info("User revoked acceptance, checking if new search is needed", zap.String("actioned_user_id", string(userID)))
		req, err := s.repo.GetRequestByID(ctx, requestID)
		if err == nil && req != nil && req.Status == domain.RequestStatusPending {
			actionedUsers, _ := s.repo.GetActionedUsers(ctx, requestID)
			acceptedCount := 0
			for _, u := range actionedUsers {
				if u.Action == domain.ActionStatusAccepted {
					acceptedCount++
				}
			}

			if acceptedCount < req.BagCount {
				reqLogger.Info("Accepted count below requirement, triggering new search wave")
				payload := domain.WaveSearchPayload{
					RequestID:   requestID,
					CurrentRing: 1,
				}
				payloadBytes, _ := json.Marshal(payload)
				job := &domain.Job{
					ID:        domain.JobID("job_" + ulid.Make().String()),
					Type:      domain.JobTypeWaveSearch,
					Payload:   string(payloadBytes),
					Status:    domain.JobStatusPending,
					RunAt:     domain.Now(),
					CreatedAt: domain.Now(),
					UpdatedAt: domain.Now(),
				}
				_ = s.queue.Enqueue(ctx, job)
			}
		}
	}

	return nil
}
