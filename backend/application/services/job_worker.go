package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"

	"github.com/oklog/ulid/v2"
	"github.com/uber/h3-go/v4"
	"go.uber.org/zap"
)

type JobWorkerService interface {
	Start(ctx context.Context)
}

type jobWorkerServiceConfig struct {
	*application.AppConfig
	maxKRings int
}

type jobWorkerService struct {
	queue        application.JobQueue
	reqRepo      application.RequestRepository
	userRepo     application.UserRepository
	notifRepo    application.NotificationRepository
	notifService NotificationService
	notifSender  application.NotificationSender
	config       *jobWorkerServiceConfig
	logger       *zap.Logger
}

func NewJobWorkerService(
	queue application.JobQueue,
	reqRepo application.RequestRepository,
	userRepo application.UserRepository,
	notifRepo application.NotificationRepository,
	notifService NotificationService,
	notifSender application.NotificationSender,
	config *application.AppConfig,
	logger *zap.Logger,
) (JobWorkerService, error) {
	edgeLen, err := h3.HexagonEdgeLengthAvgKm(config.H3HexResolution)
	if err != nil {
		return nil, err
	}

	internalConfig := &jobWorkerServiceConfig{
		AppConfig: config,
		maxKRings: int(math.Ceil(config.SearchRadiusKm / (edgeLen * 1.5))),
	}
	return &jobWorkerService{
		queue:        queue,
		reqRepo:      reqRepo,
		userRepo:     userRepo,
		notifRepo:    notifRepo,
		notifService: notifService,
		notifSender:  notifSender,
		config:       internalConfig,
		logger:       logger.With(zap.String("service", "JobWorkerService")),
	}, nil
}

func (s *jobWorkerService) Start(ctx context.Context) {
	s.logger.Info("Starting background job worker with RabbitMQ...")
	
	// Start consumer for WaveSearch jobs
	go s.runConsumer(ctx, domain.JobTypeWaveSearch)
	
	// Start consumer for Notification jobs
	go s.runConsumer(ctx, domain.JobTypeNotification)
	
	// Start consumer for CheckResponses jobs
	go s.runConsumer(ctx, domain.JobTypeCheckResponses)
}

func (s *jobWorkerService) runConsumer(ctx context.Context, jobType domain.JobType) {
	jobChan, err := s.queue.Consume(ctx, jobType)
	if err != nil {
		s.logger.Error("Failed to start consumer", zap.String("job_type", string(jobType)), zap.Error(err))
		return
	}

	for job := range jobChan {
		jobLogger := s.logger.With(zap.String("job_id", string(job.ID)), zap.String("job_type", string(job.Type)))
		jobLogger.Info("Processing job")

		var processErr error
		switch job.Type {
		case domain.JobTypeWaveSearch:
			processErr = s.processWaveSearch(ctx, job, jobLogger)
		case domain.JobTypeCheckResponses:
			processErr = s.processCheckResponses(ctx, job, jobLogger)
		case domain.JobTypeNotification:
			processErr = s.processNotification(ctx, job, jobLogger)
		default:
			jobLogger.Warn("Unknown job type")
		}

		if processErr != nil {
			jobLogger.Error("Job failed", zap.Error(processErr))
			// In a real system, we might nack or move to DLQ
		} else {
			jobLogger.Info("Job processed successfully")
		}
	}
}

func (s *jobWorkerService) processNotification(ctx context.Context, job *domain.Job, jobLogger *zap.Logger) error {
	payload, ok := job.Payload.(map[string]interface{})
	if !ok {
		// Try unmarshalling if it's a string (e.g. from RabbitMQ)
		if str, ok := job.Payload.(string); ok {
			if err := json.Unmarshal([]byte(str), &payload); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid notification payload type: %T", job.Payload)
		}
	}

	notifIDStr, ok := payload["notification_id"].(string)
	if !ok || notifIDStr == "" {
		return fmt.Errorf("missing notification_id in payload")
	}

	notifID := domain.NotificationID(notifIDStr)
	// Fetch notification from repo
	notif, err := s.notifRepo.GetNotificationByID(ctx, notifID)
	if err != nil {
		return err
	}
	if notif == nil {
		return fmt.Errorf("notification not found: %s", notifID)
	}

	return s.notifSender.Send(ctx, notif)
}

func unmarshalMetadata(notifType domain.NotificationType, raw []byte) domain.NotificationMetadata {
	switch notifType {
	case domain.NotificationTypeDonationRequest:
		var m domain.DonationRequestMetadata
		if err := json.Unmarshal(raw, &m); err == nil {
			return m
		}
	case domain.NotificationTypeDonationCompletion:
		var m domain.DonationCompletionMetadata
		if err := json.Unmarshal(raw, &m); err == nil {
			return m
		}
	case domain.NotificationTypeDonationRequestAcceptance:
		var m domain.DonationAcceptanceMetadata
		if err := json.Unmarshal(raw, &m); err == nil {
			return m
		}
	}
	return nil
}

func (s *jobWorkerService) processWaveSearch(ctx context.Context, job *domain.Job, jobLogger *zap.Logger) error {
	var payload domain.WaveSearchPayload
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return err
	}

	req, err := s.reqRepo.GetRequestByID(ctx, payload.RequestID)
	if err != nil {
		return err
	}
	reqLogger := jobLogger.With(zap.String("request_id", string(req.ID)))

	if req.Status != domain.RequestStatusPending {
		reqLogger.Info("Request is not pending, skipping job", zap.String("status", string(req.Status)))
		return nil
	}

	reqLogger.Info("Processing wave search",
		zap.Int("ring", payload.CurrentRing),
		zap.Int("retry", payload.RetryCount),
		zap.Int("bag_count", req.BagCount),
	)

	actionedUsers, err := s.reqRepo.GetActionedUsers(ctx, payload.RequestID)
	if err != nil {
		return err
	}

	actionedMap := make(map[domain.UserID]domain.ActionStatus)
	acceptedUsers := 0
	declinedUsers := 0
	pendingUsers := 0
	for _, u := range actionedUsers {
		actionedMap[u.ActionedByID] = u.Action

		if u.Action == domain.ActionStatusAccepted {
			acceptedUsers++
		} else if u.Action == domain.ActionStatusDeclined {
			declinedUsers++
		} else if s.shouldActionBeConsideredDormant(&u) {
			declinedUsers++
		} else {
			pendingUsers++
		}
	}

	if acceptedUsers >= req.BagCount {
		reqLogger.Info("All notified users have accepted the request. Stopping search.")
		return nil
	}

	usersLeftToSearch := req.BagCount - (pendingUsers + acceptedUsers)
	if usersLeftToSearch <= 0 {
		reqLogger.Info("All users reached out to, waiting for responses. Scheduling response check.",
			zap.String("payload_request_id", string(payload.RequestID)),
		)
		return s.scheduleCheckResponses(ctx, job, req, &payload, reqLogger)
	}

	reqLogger.Info("Searching for users",
		zap.Int("accepted_users", acceptedUsers),
		zap.Int("declined_users", declinedUsers),
		zap.Int("users_left_to_search", usersLeftToSearch),
	)

	centerCell := h3.Cell(h3.IndexFromString(req.LocationHex))
	ringHexes, err := s.getRingHexes(centerCell, payload.CurrentRing)
	if err != nil {
		return err
	}

	hexStrings := make([]string, len(ringHexes))
	for i, h := range ringHexes {
		hexStrings[i] = h.String()
	}

	actionedUserIDs := make([]domain.UserID, 0, len(actionedMap))
	for id := range actionedMap {
		actionedUserIDs = append(actionedUserIDs, id)
	}

	usersInRing, err := s.userRepo.GetEligibleUsersInHexes(ctx, hexStrings, req.BloodType, usersLeftToSearch, actionedUserIDs)
	if err != nil {
		return err
	}

	if len(usersInRing) == 0 {
		reqLogger.Info("No eligible users found in hexes",
			zap.Int("ring", payload.CurrentRing),
			zap.Int("hexes_count", len(hexStrings)),
		)
	} else {
		reqLogger.Info("Found eligible users in hexes",
			zap.Int("users_found", len(usersInRing)),
			zap.Int("ring", payload.CurrentRing),
			zap.Int("hexes_count", len(hexStrings)),
		)
	}

	if err := s.notifyUsers(ctx, &usersInRing, req, &actionedMap, &usersLeftToSearch, reqLogger); err != nil {
		return err
	}

	if usersLeftToSearch <= 0 {
		reqLogger.Info("All users reached out to, waiting for responses. Scheduling response check.",
			zap.String("payload_request_id", string(payload.RequestID)),
		)
		return s.scheduleCheckResponses(ctx, job, req, &payload, reqLogger)
	}

	return s.scheduleNextWaveSearch(ctx, job, req, &payload, reqLogger)
}

func (s *jobWorkerService) shouldActionBeConsideredDormant(u *domain.RequestState) bool {
	return u.Action == domain.ActionStatusPending && u.UpdatedAt.ToTime().Add(1*time.Hour).Before(time.Now())
}

func (s *jobWorkerService) getRingHexes(centerCell h3.Cell, ring int) ([]h3.Cell, error) {
	ringHexes, err := h3.GridRing(centerCell, ring)
	if err != nil {
		allRings, diskErr := h3.GridDiskDistances(centerCell, ring)
		if diskErr != nil {
			return nil, diskErr
		}
		ringHexes = allRings[ring]
	}

	if ring == 1 {
		ringHexes = append(ringHexes, centerCell)
	}
	return ringHexes, nil
}

func (s *jobWorkerService) notifyUsers(ctx context.Context, users *[]domain.User, req *domain.DonationRequest, actionedMap *map[domain.UserID]domain.ActionStatus, usersLeftToSearch *int, reqLogger *zap.Logger) error {
	for _, u := range *users {
		if _, ok := (*actionedMap)[u.ID]; ok {
			continue
		}
		if u.ID == req.UserID {
			continue
		}

		state := &domain.RequestState{
			RequestID:    req.ID,
			ActionedByID: u.ID,
			Action:       domain.ActionStatusPending,
			CreatedAt:    domain.Now(),
			UpdatedAt:    domain.Now(),
		}
		if err := s.reqRepo.SaveRequestState(ctx, state); err != nil {
			reqLogger.Error("Error saving request state", zap.Error(err), zap.String("user_id", string(u.ID)))
			continue
		}

		title := "Urgent: Blood Donation Request"
		content := "Someone needs " + string(req.BloodType) + " blood near you!"
		metadata := domain.DonationRequestMetadata{RequestID: string(req.ID)}
		
		if _, err := s.notifService.Submit(ctx, domain.NotificationTypeDonationRequest, u.ID, title, content, metadata); err != nil {
			reqLogger.Error("Error submitting notification", zap.Error(err), zap.String("user_id", string(u.ID)))
			continue
		}

		reqLogger.Info("Notified eligible user", zap.String("notified_user_id", string(u.ID)))
		(*usersLeftToSearch)--
	}
	return nil
}

func (s *jobWorkerService) scheduleCheckResponses(ctx context.Context, job *domain.Job, req *domain.DonationRequest, payload *domain.WaveSearchPayload, reqLogger *zap.Logger) error {
	runAt := time.Now().Add(s.config.RequestAcceptanceWindow)

	nextJob := &domain.Job{
		ID:        domain.JobID("job_" + ulid.Make().String()),
		Type:      domain.JobTypeCheckResponses,
		Payload:   job.Payload,
		Status:    domain.JobStatusPending,
		RunAt:     domain.ISOTimestamp(runAt.Format("2006-01-02T15:04:05.000Z")),
		CreatedAt: domain.Now(),
		UpdatedAt: domain.Now(),
	}

	return s.queue.Enqueue(ctx, nextJob)
}

func (s *jobWorkerService) processCheckResponses(ctx context.Context, job *domain.Job, jobLogger *zap.Logger) error {
	jobLogger.Info("Checking responses for request")
	return s.processWaveSearch(ctx, job, jobLogger)
}
func (s *jobWorkerService) scheduleNextWaveSearch(ctx context.Context, job *domain.Job, req *domain.DonationRequest, payload *domain.WaveSearchPayload, reqLogger *zap.Logger) error {
	nextRing := payload.CurrentRing + 1
	runAt := time.Now().Add(s.config.WaveSearchInterval)
	nextRetryCount := payload.RetryCount

	if nextRing > s.config.maxKRings {
		if payload.RetryCount >= s.config.WaveSearchMaxRetries {
			reqLogger.Warn("Exhausted all search rings and max retries. Marking request as failed.", zap.Int("max_retries", s.config.WaveSearchMaxRetries))
			req.Status = domain.RequestStatusFailed
			req.UpdatedAt = domain.Now()
			_ = s.reqRepo.UpdateRequest(ctx, req)
			return nil
		}

		reqLogger.Info("Exhausted search radius. Delaying and restarting from Ring 1",
			zap.Int("retry_count_next", payload.RetryCount+1),
			zap.Int("max_retries", s.config.WaveSearchMaxRetries),
		)
		nextRing = 1
		nextRetryCount = payload.RetryCount + 1
		runAt = time.Now().Add(s.config.WaveSearchRetryDelay)
	}

	nextPayload := domain.WaveSearchPayload{
		RequestID:   req.ID,
		CurrentRing: nextRing,
		RetryCount:  nextRetryCount,
	}
	nextJob := &domain.Job{
		ID:        domain.JobID("job_" + ulid.Make().String()),
		Type:      domain.JobTypeWaveSearch,
		Payload:   nextPayload,
		Status:    domain.JobStatusPending,
		RunAt:     domain.ISOTimestamp(runAt.Format("2006-01-02T15:04:05.000Z")),
		CreatedAt: domain.Now(),
		UpdatedAt: domain.Now(),
	}

	return s.queue.Enqueue(ctx, nextJob)
}
