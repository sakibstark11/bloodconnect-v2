package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	db_repos "bloodconnect/adapters/db"
	"bloodconnect/adapters/memory"
	"bloodconnect/adapters/sqlite"
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

	db, err := sqlite.SetupDatabase(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
	if err != nil {
		t.Fatalf("failed to setup test database: %v", err)
	}

	userRepo := db_repos.NewUserRepository(db)
	reqRepo := db_repos.NewRequestRepository(db)
	notifRepo := db_repos.NewNotificationRepository(db)

	queue := memory.NewJobQueue()

	notifSender := &dummyNotificationSender{}
	notifService := services.NewNotificationService(notifRepo, queue)
	reqService := services.NewRequestService(reqRepo, userRepo, queue, notifService, config, logger)
	userService := services.NewUserService(userRepo, config)
	jobWorker, _ := services.NewJobWorkerService(queue, reqRepo, userRepo, notifRepo, notifService, notifSender, config, logger)

	jobWorker.Start(context.Background())

	return &TestSuite{
		userRepo:     userRepo,
		reqRepo:      reqRepo,
		notifRepo:    notifRepo,
		jobQueue:     queue,
		notifService: notifService,
		reqService:   reqService,
		userService:  userService,
		jobWorker:    jobWorker,
		config:       config,
		logger:       logger,
	}
}

var phoneCounter int

func createTestUser(ctx context.Context, ts *TestSuite, email string, bloodType domain.BloodType, lat, lng float64) domain.UserID {
	phoneCounter++
	phone := fmt.Sprintf("+88017%08d", phoneCounter)
	userID, err := ts.userService.Signup(ctx, "Test User", email, "password", phone, bloodType, lat, lng)
	if err != nil {
		panic("Failed to signup test user: " + err.Error())
	}

	return userID
}

func runWorkerOnce(ctx context.Context, ts *TestSuite) {

	time.Sleep(ts.config.JobWorkerTickerInterval * 5)
}
