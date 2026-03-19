package tests

import (
	"context"
	"testing"
	"time"

	"bloodconnect/adapters/memory"
	"bloodconnect/application"
	"bloodconnect/application/domain"
	"bloodconnect/application/services"

	"go.uber.org/zap"
)

type dummyNotificationSender struct {
}

func (s *dummyNotificationSender) Send(ctx context.Context, notification *domain.Notification) error {
	return nil
}

type TestSuite struct {
	userRepo     application.UserRepository
	reqRepo      application.RequestRepository
	notifRepo    application.NotificationRepository
	jobQueue     application.JobQueue
	notifService services.NotificationService
	reqService   services.RequestService
	userService  services.UserService
	jobWorker    services.JobWorkerService
	config       *application.AppConfig
	logger       *zap.Logger
}

func setupTestSuite(t *testing.T) *TestSuite {
	logger := zap.NewNop()
	config := application.DefaultAppConfig()

	config.WaveSearchInterval = 100 * time.Millisecond
	config.WaveSearchRetryDelay = 200 * time.Millisecond
	config.RequestAcceptanceWindow = 500 * time.Millisecond
	config.JobWorkerTickerInterval = 10 * time.Millisecond

	userRepo := memory.NewUserRepository()
	reqRepo := memory.NewRequestRepository()
	notifRepo := memory.NewNotificationRepository()
	jobQueue := memory.NewJobQueue()

	notifService := services.NewNotificationService(notifRepo, &dummyNotificationSender{})
	reqService := services.NewRequestService(reqRepo, userRepo, jobQueue, notifService, config, logger)
	userService := services.NewUserService(userRepo, config)
	jobWorker, _ := services.NewJobWorkerService(jobQueue, reqRepo, userRepo, notifService, config, logger)

	jobWorker.Start(context.Background())

	return &TestSuite{
		userRepo:     userRepo,
		reqRepo:      reqRepo,
		notifRepo:    notifRepo,
		jobQueue:     jobQueue,
		notifService: notifService,
		reqService:   reqService,
		userService:  userService,
		jobWorker:    jobWorker,
		config:       config,
		logger:       logger,
	}
}

func createTestUser(ctx context.Context, ts *TestSuite, email string, bloodType domain.BloodType, lat, lng float64) domain.UserID {
	userID, err := ts.userService.Signup(ctx, "Test User", email, "password", "+880123456789")
	if err != nil {
		panic("Failed to signup test user: " + err.Error())
	}

	ctxUser := context.WithValue(ctx, domain.UserIDKey, userID)

	err = ts.userService.UpdateHealth(ctxUser, domain.InfoTypeBloodType, string(bloodType))
	if err != nil {
		panic("Failed to update health for test user: " + err.Error())
	}

	err = ts.userService.UpdateLocation(ctxUser, lat, lng)
	if err != nil {
		panic("Failed to update location for test user: " + err.Error())
	}

	return userID
}

func runWorkerOnce(ctx context.Context, ts *TestSuite) {

	time.Sleep(ts.config.JobWorkerTickerInterval * 5)
}
