package main

import (
	"context"
	"log"
	"net/http"

	"github.com/sakibalam/bloodconnect/adapters/dummy"
	api_http "github.com/sakibalam/bloodconnect/adapters/http"
	"github.com/sakibalam/bloodconnect/adapters/sqlite"
	"github.com/sakibalam/bloodconnect/application"
)

func main() {
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

	userService := application.NewUserService(userRepo)
	notifService := application.NewNotificationService(notifRepo, notifSender)

	// Create configuration
	appConfig := application.DefaultAppConfig()

	requestService := application.NewRequestService(requestRepo, queue, notifService, appConfig)

	workerService, err := application.NewJobWorkerService(queue, requestRepo, userRepo, notifService, appConfig)
	if err != nil {
		log.Fatalf("failed to initialize job worker service: %v", err)
	}

	// Start Background Workers
	workerService.Start(context.Background())

	// Initialize HTTP Router
	router := api_http.SetupRouter(userService, notifService, requestService)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
