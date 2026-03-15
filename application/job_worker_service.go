package application

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sakibalam/bloodconnect/domain"
	"github.com/sakibalam/bloodconnect/ports"
	"github.com/uber/h3-go/v4"
)

type JobWorkerService interface {
	Start(ctx context.Context)
}

type jobWorkerServiceConfig struct {
	*AppConfig
	maxKRings int
}

type jobWorkerService struct {
	queue        ports.JobQueue
	reqRepo      ports.RequestRepository
	userRepo     ports.UserRepository
	notifService NotificationService
	config       *jobWorkerServiceConfig
}

func NewJobWorkerService(queue ports.JobQueue, reqRepo ports.RequestRepository, userRepo ports.UserRepository, notifService NotificationService, config *AppConfig) (JobWorkerService, error) {
	edgeLen, err := h3.HexagonEdgeLengthAvgKm(config.H3HexResolution)
	if err != nil {
		return nil, err
	}

	internalConfig := &jobWorkerServiceConfig{
		AppConfig: config,
		// Maximum possible K we are willing to search
		maxKRings: int(math.Ceil(config.SearchRadiusKm / (edgeLen * 1.5))),
	}
	return &jobWorkerService{
		queue:        queue,
		reqRepo:      reqRepo,
		userRepo:     userRepo,
		notifService: notifService,
		config:       internalConfig,
	}, nil
}

func (s *jobWorkerService) Start(ctx context.Context) {
	log.Println("Starting background job worker...")
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping background job worker...")
				return
			case <-ticker.C:
				s.processNextJob(ctx)
			}
		}
	}()
}

func (s *jobWorkerService) processNextJob(ctx context.Context) {
	job, err := s.queue.FetchNextAvailable(ctx)
	if err != nil {
		log.Printf("Error fetching next job: %v\n", err)
		return
	}
	if job == nil {
		return // No jobs ready
	}

	// Mark as processing
	if err := s.queue.MarkStatus(ctx, job.ID, domain.JobStatusProcessing); err != nil {
		log.Printf("Error marking job %s as processing: %v\n", job.ID, err)
		return
	}

	log.Printf("Processing job: %s of type %s\n", job.ID, job.Type)

	var processErr error
	switch job.Type {
	case domain.JobTypeWaveSearch:
		processErr = s.processWaveSearch(ctx, job)
	default:
		log.Printf("Unknown job type: %s\n", job.Type)
	}

	if processErr != nil {
		log.Printf("Job failed: %s, error: %v\n", job.ID, processErr)
		_ = s.queue.MarkStatus(ctx, job.ID, domain.JobStatusFailed)
	} else {
		log.Printf("Job completed: %s\n", job.ID)
		_ = s.queue.MarkStatus(ctx, job.ID, domain.JobStatusCompleted)
	}
}

func (s *jobWorkerService) processWaveSearch(ctx context.Context, job *domain.Job) error {
	var payload domain.WaveSearchPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return err
	}

	// Get the request to see if it's still pending
	req, err := s.reqRepo.GetRequestByID(ctx, payload.RequestID)
	if err != nil {
		return err
	}
	if req.Status != domain.RequestStatusPending {
		// Stop processing if request was fulfilled or cancelled
		log.Printf("Request %s is not pending, skipping job %s\n", req.ID, job.ID)
		return nil
	}

	centerCell := h3.Cell(h3.IndexFromString(payload.CenterHex))

	log.Printf("Processing wave search for request %s, ring %d, users left to search %d, bags needed %d\n", req.ID, payload.CurrentRing, payload.UsersLeftToSearch, req.BagCount)

	var ringHexes []h3.Cell

	ringHexes, err = h3.GridRing(centerCell, payload.CurrentRing)
	if err != nil {
		// fallback for pentagons like the example code
		allRings, diskErr := h3.GridDiskDistances(centerCell, payload.CurrentRing)
		if diskErr != nil {
			return diskErr
		}
		ringHexes = allRings[payload.CurrentRing]
	}

	if payload.CurrentRing == 1 {
		ringHexes = append(ringHexes, centerCell)
	}
	// Convert ring hexes to strings
	hexStrings := make([]string, len(ringHexes))
	for i, h := range ringHexes {
		hexStrings[i] = h.String()
	}

	// Find actioned users so we don't ping them again

	actionedUsers, err := s.reqRepo.GetActionedUsers(ctx, payload.RequestID)
	if err != nil {
		return err
	}

	usersLeftToSearch := payload.UsersLeftToSearch

	actionedMap := make(map[string]bool)
	for _, u := range actionedUsers {
		// If actioned users have been pending for > 1 hour, treat them as declined
		if u.Action == domain.ActionStatusPending && u.UpdatedAt.Add(1*time.Hour).Before(time.Now()) {
			usersLeftToSearch++
		}
		actionedMap[u.ActionedByID] = true
	}

	// Get users in these hexes
	usersInRing, err := s.userRepo.GetEligibleUsersInHexes(ctx, hexStrings, string(req.BloodType), payload.UsersLeftToSearch)
	if err != nil {
		return err
	}

	log.Printf("found %d users in ring %d for request %s", len(usersInRing), payload.CurrentRing, req.ID)

	for _, u := range usersInRing {
		if usersLeftToSearch <= 0 {
			break
		}
		if actionedMap[u.ID] {
			continue // Already reached out
		}
		if u.ID == req.UserID {
			continue // Don't notify the requester
		}

		// Save state
		state := &domain.RequestState{
			RequestID:    req.ID,
			ActionedByID: u.ID,
			Action:       domain.ActionStatusPending,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_ = s.reqRepo.SaveRequestState(ctx, state)

		// Notify
		title := "Urgent: Blood Donation Request"
		content := "Someone needs " + string(req.BloodType) + " blood near you!"
		_, _ = s.notifService.Submit(ctx, domain.NotificationTypeDonationRequest, u.ID, title, content)

		usersLeftToSearch--
		log.Printf("Notified user %s for request %s, ring %d, users left to search %d, bags needed %d\n", u.ID, req.ID, payload.CurrentRing, usersLeftToSearch, req.BagCount)
	}

	log.Printf("Processed wave search for request %s, ring %d, users left to search %d, bags needed %d\n", req.ID, payload.CurrentRing, usersLeftToSearch, req.BagCount)
	if usersLeftToSearch <= 0 {
		log.Printf("Request %s completed\n", req.ID)
		return nil
	}
	// Schedule the next check (1 hour from now) to see if we need a new wave
	nextPayload := domain.WaveSearchPayload{
		RequestID:         req.ID,
		CurrentRing:       payload.CurrentRing + 1,
		CenterHex:         payload.CenterHex,
		UsersLeftToSearch: usersLeftToSearch,
	}
	nextPayloadBytes, _ := json.Marshal(nextPayload)
	nextJob := &domain.Job{
		ID:        "job_" + ulid.Make().String(),
		Type:      domain.JobTypeWaveSearch,
		Payload:   string(nextPayloadBytes),
		Status:    domain.JobStatusPending,
		RunAt:     time.Now().Add(s.config.JobQueueInterval),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// We enqueue the next wave as a fallback. If the request gets completed before 1 hour, that future job will just exit early on `req.Status != Pending`
	return s.queue.Enqueue(ctx, nextJob)
}
