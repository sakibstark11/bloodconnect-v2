package services

import (
	"context"
	"encoding/json"
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
	notifService NotificationService
	config       *jobWorkerServiceConfig
	logger       *zap.Logger
}

func NewJobWorkerService(queue application.JobQueue, reqRepo application.RequestRepository, userRepo application.UserRepository, notifService NotificationService, config *application.AppConfig, logger *zap.Logger) (JobWorkerService, error) {
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
		notifService: notifService,
		config:       internalConfig,
		logger:       logger.With(zap.String("service", "JobWorkerService")),
	}, nil
}

func (s *jobWorkerService) Start(ctx context.Context) {
	s.logger.Info("Starting background job worker...")
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Stopping background job worker...")
				return
			case <-ticker.C:
				s.logger.Info("Processing next job...")
				s.processNextJob(ctx)
			}
		}
	}()
}

func (s *jobWorkerService) processNextJob(ctx context.Context) {
	job, err := s.queue.FetchNextAvailable(ctx)
	if err != nil {
		s.logger.Error("Error fetching next job", zap.Error(err))
		return
	}
	if job == nil {
		return
	}

	jobLogger := s.logger.With(zap.String("job_id", job.ID), zap.String("job_type", string(job.Type)))

	if err := s.queue.MarkStatus(ctx, job.ID, domain.JobStatusProcessing); err != nil {
		jobLogger.Error("Error marking job as processing", zap.Error(err))
		return
	}

	jobLogger.Info("Processing job")

	var processErr error
	switch job.Type {
	case domain.JobTypeWaveSearch:
		processErr = s.processWaveSearch(ctx, job, jobLogger)
	default:
		jobLogger.Warn("Unknown job type")
	}

	if processErr != nil {
		jobLogger.Error("Job failed", zap.Error(processErr))
		_ = s.queue.MarkStatus(ctx, job.ID, domain.JobStatusFailed)
	} else {
		jobLogger.Info("Job processed successfully, deleting from queue")
		_ = s.queue.Delete(ctx, job.ID)
	}
}

func (s *jobWorkerService) processWaveSearch(ctx context.Context, job *domain.Job, jobLogger *zap.Logger) error {
	var payload domain.WaveSearchPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return err
	}

	req, err := s.reqRepo.GetRequestByID(ctx, payload.RequestID)
	if err != nil {
		return err
	}
	reqLogger := jobLogger.With(zap.String("request_id", req.ID))

	if req.Status != domain.RequestStatusPending {
		reqLogger.Info("Request is not pending, skipping job", zap.String("status", string(req.Status)))
		return nil
	}

	centerCell := h3.Cell(h3.IndexFromString(payload.CenterHex))

	reqLogger.Info("Processing wave search",
		zap.Int("ring", payload.CurrentRing),
		zap.Int("retry", payload.RetryCount),
		zap.Int("users_left_to_search", payload.UsersLeftToSearch),
		zap.Int("bag_count", req.BagCount),
	)

	var ringHexes []h3.Cell

	ringHexes, err = h3.GridRing(centerCell, payload.CurrentRing)
	if err != nil {
		allRings, diskErr := h3.GridDiskDistances(centerCell, payload.CurrentRing)
		if diskErr != nil {
			return diskErr
		}
		ringHexes = allRings[payload.CurrentRing]
	}

	if payload.CurrentRing == 1 {
		ringHexes = append(ringHexes, centerCell)
	}

	hexStrings := make([]string, len(ringHexes))
	for i, h := range ringHexes {
		hexStrings[i] = h.String()
	}

	actionedUsers, err := s.reqRepo.GetActionedUsers(ctx, payload.RequestID)
	if err != nil {
		return err
	}

	usersLeftToSearch := payload.UsersLeftToSearch

	actionedMap := make(map[string]bool)
	for _, u := range actionedUsers {
		if u.Action == domain.ActionStatusPending && u.UpdatedAt.ToTime().Add(1*time.Hour).Before(time.Now()) {
			usersLeftToSearch++
		}
		actionedMap[u.ActionedByID] = true
	}

	usersInRing, err := s.userRepo.GetEligibleUsersInHexes(ctx, hexStrings, string(req.BloodType), payload.UsersLeftToSearch)
	if err != nil {
		return err
	}

	reqLogger.Info("Found eligible users in hexes",
		zap.Int("users_found", len(usersInRing)),
		zap.Int("ring", payload.CurrentRing),
		zap.Int("hexes_count", len(hexStrings)),
	)

	for _, u := range usersInRing {
		if usersLeftToSearch <= 0 {
			break
		}
		if actionedMap[u.ID] {
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
		_ = s.reqRepo.SaveRequestState(ctx, state)

		title := "Urgent: Blood Donation Request"
		content := "Someone needs " + string(req.BloodType) + " blood near you!"
		_, _ = s.notifService.Submit(ctx, domain.NotificationTypeDonationRequest, u.ID, title, content)

		usersLeftToSearch--
		reqLogger.Info("Notified eligible user", zap.String("notified_user_id", u.ID))
	}

	if usersLeftToSearch <= 0 {
		reqLogger.Info("All users reached out to, waiting for responses")
		return nil
	}

	nextRing := payload.CurrentRing + 1
	runAt := time.Now().Add(s.config.JobQueueInterval)
	nextRetryCount := payload.RetryCount

	if nextRing > s.config.maxKRings {
		if payload.RetryCount >= s.config.WaveSearchMaxRetries {
			reqLogger.Warn("Exhausted all search rings and max retries. Marking request as failed.", zap.Int("max_retries", s.config.WaveSearchMaxRetries))
			req.Status = domain.RequestStatusFailed
			req.UpdatedAt = domain.Now()
			_ = s.reqRepo.UpdateRequest(ctx, req)
			_ = s.queue.MarkStatus(ctx, job.ID, domain.JobStatusCompleted)
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
		RequestID:         req.ID,
		CurrentRing:       nextRing,
		CenterHex:         payload.CenterHex,
		UsersLeftToSearch: usersLeftToSearch,
		RetryCount:        nextRetryCount,
	}
	nextPayloadBytes, _ := json.Marshal(nextPayload)
	nextJob := &domain.Job{
		ID:        "job_" + ulid.Make().String(),
		Type:      domain.JobTypeWaveSearch,
		Payload:   string(nextPayloadBytes),
		Status:    domain.JobStatusPending,
		RunAt:     domain.ISOTimestamp(runAt.Format("2006-01-02T15:04:05Z")),
		CreatedAt: domain.Now(),
		UpdatedAt: domain.Now(),
	}

	return s.queue.Enqueue(ctx, nextJob)
}
