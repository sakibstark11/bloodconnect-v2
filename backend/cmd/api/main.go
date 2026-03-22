package main

import (
	"context"
	"log"
	"net/http"

	"bloodconnect/adapters/dummy"
	api_http "bloodconnect/adapters/http"
	"bloodconnect/adapters/postgres"
	"bloodconnect/adapters/postgres/repos"
	"bloodconnect/adapters/rabbitmq"
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

	appConfig := application.DefaultAppConfig()

	db, err := postgres.SetupDatabase(appConfig.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	userRepo := repos.NewUserRepository(db)
	notifRepo := repos.NewNotificationRepository(db)
	notifSender := dummy.NewNotificationSender()
	requestRepo := repos.NewRequestRepository(db)
	
	queue, err := rabbitmq.NewJobQueue(appConfig.RabbitMQURL)
	if err != nil {
		logger.Fatal("failed to initialize rabbitmq queue", zap.Error(err))
	}

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
