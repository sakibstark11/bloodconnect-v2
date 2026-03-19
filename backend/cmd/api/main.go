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

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	db, err := sqlite.SetupDatabase("bloodconnect.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	userRepo := repos.NewUserRepository(db)
	notifRepo := repos.NewNotificationRepository(db)
	notifSender := dummy.NewNotificationSender()
	requestRepo := repos.NewRequestRepository(db)
	queue := repos.NewJobQueue(db)

	appConfig := application.DefaultAppConfig()

	userService := services.NewUserService(userRepo, appConfig)
	notifService := services.NewNotificationService(notifRepo, notifSender)

	requestService := services.NewRequestService(requestRepo, userRepo, queue, notifService, appConfig, logger)

	workerService, err := services.NewJobWorkerService(queue, requestRepo, userRepo, notifService, appConfig, logger)
	if err != nil {
		log.Fatalf("failed to initialize job worker service: %v", err)
	}

	workerService.Start(context.Background())

	router := api_http.SetupRouter(userService, notifService, requestService, appConfig)
	loggedRouter := api_http.RequestLogger(logger)(router)

	logger.Info("Server starting", zap.String("addr", ":8080"))
	if err := http.ListenAndServe(":8080", loggedRouter); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}
