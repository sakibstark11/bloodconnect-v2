package main

import (
	"context"
	"log"
	"net/http"

	"github.com/sakibalam/bloodconnect/adapters/dummy"
	api_http "github.com/sakibalam/bloodconnect/adapters/http"
	"github.com/sakibalam/bloodconnect/adapters/sqlite"
	"github.com/sakibalam/bloodconnect/application"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Configure Global Zap Logger
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Enforce ISO-format timestamps
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Initialize Database
	db, err := sqlite.SetupDatabase("bloodconnect.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	userRepo := sqlite.NewUserRepository(db)
	notifRepo := sqlite.NewNotificationRepository(db)
	notifSender := dummy.NewNotificationSender()
	requestRepo := sqlite.NewRequestRepository(db)
	queue := sqlite.NewJobQueue(db)

	// Create configuration
	appConfig := application.DefaultAppConfig()

	userService := application.NewUserService(userRepo, appConfig)
	notifService := application.NewNotificationService(notifRepo, notifSender)

	requestService := application.NewRequestService(requestRepo, userRepo, queue, notifService, appConfig, logger)

	workerService, err := application.NewJobWorkerService(queue, requestRepo, userRepo, notifService, appConfig, logger)
	if err != nil {
		log.Fatalf("failed to initialize job worker service: %v", err)
	}

	// Start Background Workers
	workerService.Start(context.Background())

	// Initialize HTTP Router
	router := api_http.SetupRouter(userService, notifService, requestService)
	
	// Wrap router with Zap Logger Middleware
	loggedRouter := api_http.RequestLogger(logger)(router)

	logger.Info("Server starting", zap.String("port", ":8080"))
	if err := http.ListenAndServe(":8080", loggedRouter); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}
