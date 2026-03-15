package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/sakibalam/bloodconnect/adapters/dummy"
	"github.com/sakibalam/bloodconnect/adapters/sqlite"
	"github.com/sakibalam/bloodconnect/application"
	"github.com/sakibalam/bloodconnect/domain"
	"github.com/uber/h3-go/v4"
)

// Dhaka approximate bounding box for random generation
const (
	minLat      = 23.68
	maxLat      = 23.90
	minLng      = 90.33
	maxLng      = 90.50
	maxBagCount = 5
	maxDaysOut  = 8
	maxRequests = 50
	maxUsers    = 1_000_000
)

var bloodTypes = []domain.BloodType{
	domain.BloodTypeAPos, domain.BloodTypeANeg,
	domain.BloodTypeBPos, domain.BloodTypeBNeg,
	domain.BloodTypeABPos, domain.BloodTypeABNeg,
	domain.BloodTypeOPos, domain.BloodTypeONeg,
}

func main() {
	log.Println("Starting data population...")

	// Initialize DB
	db, err := sqlite.SetupDatabase("bloodconnect.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	userRepo := sqlite.NewUserRepository(db)
	notifRepo := sqlite.NewNotificationRepository(db)
	notifSender := dummy.NewNotificationSender()
	requestRepo := sqlite.NewRequestRepository(db)
	queue := sqlite.NewJobQueue(db)

	appConfig := application.DefaultAppConfig()

	userService := application.NewUserService(userRepo)
	notifService := application.NewNotificationService(notifRepo, notifSender)
	requestService := application.NewRequestService(requestRepo, queue, notifService, appConfig)

	ctx := context.Background()

	// 1. Generate 1000 random users in Dhaka
	log.Printf("Generating %d random users...", maxUsers)
	var userIDs []string

	// Create a run-specific suffix to avoid unique constraint collisions
	runSuffix := time.Now().UnixNano() % 100000

	for i := 0; i < maxUsers; i++ {
		name := fmt.Sprintf("User %d", i)
		email := fmt.Sprintf("user%d_%d@example.com", i, runSuffix)
		phone := fmt.Sprintf("+8801%d%04d", runSuffix, i)

		uid, err := userService.Signup(ctx, name, email, "password123", phone)
		if err != nil {
			log.Printf("Failed to create user %d: %v", i, err)
			continue
		}

		userIDs = append(userIDs, uid)

		// Set random blood type
		bType := bloodTypes[rand.Intn(len(bloodTypes))]
		_ = userService.UpdateHealth(ctx, uid, domain.InfoTypeBloodType, string(bType))

		// Set random location in Bangladesh
		lat := minLat + rand.Float64()*(maxLat-minLat)
		lng := minLng + rand.Float64()*(maxLng-minLng)
		cell, _ := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, 9)
		hexStr := cell.String()
		_ = userService.UpdateLocation(ctx, uid, lat, lng, hexStr)
	}

	// 2. Generate 5 blood requests
	log.Printf("Generating %d blood requests...\n", maxRequests)
	if len(userIDs) == 0 {
		log.Fatalf("No users were created, cannot generate requests!")
	}

	for i := 0; i < maxRequests; i++ {
		uid := userIDs[rand.Intn(len(userIDs))]
		bType := bloodTypes[rand.Intn(len(bloodTypes))]

		lat := minLat + rand.Float64()*(maxLat-minLat)
		lng := minLng + rand.Float64()*(maxLng-minLng)

		daysOut := rand.Intn(maxDaysOut) + 1
		reqDate := time.Now().AddDate(0, 0, daysOut)

		bagCount := rand.Intn(maxBagCount) + 1

		reqID, err := requestService.SubmitRequest(ctx,
			uid, string(bType), "+8801900000000",
			fmt.Sprintf("Urgent blood needed for patient %d", i),
			"Dhaka Medical College", "Dhaka Hospital", lat, lng, bagCount, reqDate)

		if err != nil {
			log.Printf("Failed to create request: %v", err)
			continue
		}

		log.Printf("Created request %s for %s (%d bags)", reqID, bType, bagCount)
	}

	log.Println("Data population complete!")
}
