package main

import (
	"context"
	"log"
	"net/http"

	"bloodconnect/adapters/dummy"
	api_http "bloodconnect/adapters/http"
	"bloodconnect/adapters/sqlite"
	"bloodconnect/adapters/sqlite/repos"
	"bloodconnect/application"
	"bloodconnect/application/services"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Configure Global Zap Logger
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Initialize Database (path is an SQLite adapter concern)
	db, err := sqlite.SetupDatabase("bloodconnect.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	userRepo := repos.NewUserRepository(db)
	notifRepo := repos.NewNotificationRepository(db)
	notifSender := dummy.NewNotificationSender()
	requestRepo := repos.NewRequestRepository(db)
	queue := repos.NewJobQueue(db)

	// AppConfig contains only business-logic parameters
	appConfig := application.DefaultAppConfig()

	userService := services.NewUserService(userRepo, appConfig)
	notifService := services.NewNotificationService(notifRepo, notifSender)

	requestService := services.NewRequestService(requestRepo, userRepo, queue, notifService, appConfig, logger)

	workerService, err := services.NewJobWorkerService(queue, requestRepo, userRepo, notifService, appConfig, logger)
	if err != nil {
		log.Fatalf("failed to initialize job worker service: %v", err)
	}

	// Start Background Workers
	workerService.Start(context.Background())

	// Initialize HTTP Router (address is an HTTP adapter concern)
	router := api_http.SetupRouter(userService, notifService, requestService)
	loggedRouter := api_http.RequestLogger(logger)(router)

	logger.Info("Server starting", zap.String("addr", ":8080"))
	if err := http.ListenAndServe(":8080", loggedRouter); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}
