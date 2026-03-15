package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

)

const (
	baseURL     = "http://localhost:8080"
	maxUsers    = 10_000
	maxRequests = 50
	numWorkers  = 5
)

var (
	bloodTypes     = []string{"A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-"}
	minLat, maxLat = 23.68, 23.90
	minLng, maxLng = 90.33, 90.50
)

type client struct {
	hc *http.Client
}

func main() {
	start := time.Now()
	log.Printf("Starting data population with %d workers...", numWorkers)

	c := &client{
		hc: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        numWorkers,
				MaxIdleConnsPerHost: numWorkers,
			},
		},
	}

	userIDsChan := make(chan string, maxUsers)
	var userWg sync.WaitGroup

	// 1. Worker Pool for User Creation
	jobs := make(chan int, maxUsers)
	for w := 1; w <= numWorkers; w++ {
		userWg.Add(1)
		go func() {
			defer userWg.Done()
			for i := range jobs {
				uid, err := populateUser(c, i)
				if err == nil {
					userIDsChan <- uid
				}
			}
		}()
	}

	for i := 0; i < maxUsers; i++ {
		jobs <- i
	}
	close(jobs)
	userWg.Wait()
	close(userIDsChan)

	var userIDs []string
	for uid := range userIDsChan {
		userIDs = append(userIDs, uid)
	}

	// 2. Generate Blood Requests with WaitGroup
	log.Printf("Generating %d requests...", maxRequests)
	var reqWg sync.WaitGroup // NEW: Added WaitGroup for requests
	for i := 0; i < maxRequests; i++ {
		if len(userIDs) > 0 {
			reqWg.Add(1)
			go func(idx int) {
				defer reqWg.Done()
				uid := userIDs[rand.Intn(len(userIDs))]
				populateRequest(c, idx, uid)
			}(i)
		}
	}

	reqWg.Wait() // NEW: Wait for all request goroutines to finish
	log.Printf("Process completed in %v", time.Since(start))
}

func populateUser(c *client, index int) (string, error) {
	suffix := time.Now().UnixNano() % 100000
	email := fmt.Sprintf("user%d_%d@example.com", index, suffix)
	phone := fmt.Sprintf("+8801%d%04d", suffix%100, index%10000)

	res, err := c.doRequest(http.MethodPost, "/users/signup", map[string]string{
		"name":     fmt.Sprintf("User %d", index),
		"email":    email,
		"password": "password123",
		"phone":    phone,
	})
	if err != nil {
		log.Printf("User creation failed: %v", err)
		return "", err
	}
	uid := res["id"].(string)

	bType := bloodTypes[rand.Intn(len(bloodTypes))]
	_ = c.updateHealth(uid, "blood_type", bType)

	lat := minLat + rand.Float64()*(maxLat-minLat)
	lng := minLng + rand.Float64()*(maxLng-minLng)
	_ = c.updateLocation(uid, lat, lng)

	return uid, nil
}

func populateRequest(c *client, i int, uid string) {
	lat := minLat + rand.Float64()*(maxLat-minLat)
	lng := minLng + rand.Float64()*(maxLng-minLng)

	payload := map[string]interface{}{
		"user_id":          uid,
		"location_lat":     lat,
		"location_lng":     lng,
		"bag_count":        rand.Intn(5) + 1,
		"required_by_date": time.Now().AddDate(0, 0, 3).Format(time.RFC3339),
		"blood_type":       bloodTypes[rand.Intn(len(bloodTypes))],
		"contact_phone":    fmt.Sprintf("+8801%d%04d", i%100, i%10000),
		"description":      fmt.Sprintf("Urgent request #%d", i),
		"requester_info":   "Emergency Unit",
		"location_name":    "City Hospital",
	}
	_, err := c.doRequest(http.MethodPost, "/requests", payload)
	if err != nil {
		log.Printf("Failed to create request %d: %v", i, err)
	} else {
		log.Printf("Successfully created request %d", i)
	}
}

// --- Reusable HTTP logic ---

func (c *client) updateHealth(uid, infoType, details string) error {
	_, err := c.doRequest(http.MethodPut, "/users/health", map[string]string{
		"user_id": uid, "info_type": infoType, "details": details,
	})
	return err
}

func (c *client) updateLocation(uid string, lat, lng float64) error {
	_, err := c.doRequest(http.MethodPut, "/users/location", map[string]interface{}{
		"user_id": uid, "lat": lat, "lng": lng,
	})
	return err
}

func (c *client) doRequest(method, path string, body interface{}) (map[string]interface{}, error) {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, baseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if resp.ContentLength != 0 {
		json.NewDecoder(resp.Body).Decode(&result)
	}
	return result, nil
}
