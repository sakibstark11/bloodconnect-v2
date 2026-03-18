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

	"github.com/uber/h3-go/v4"
)

const (
	baseURL          = "http://localhost:8080"
	maxUsers         = 1_000
	maxRequests      = 50
	numWorkers       = 5
	randomAcceptance = 10
	randomDeclines   = 15
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

	// 1. Generate Users
	userIDsChan := make(chan string, maxUsers)
	var userWg sync.WaitGroup
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

	// 2. Generate Requests
	log.Printf("Generating %d requests...", maxRequests)
	var reqIDs []string
	var reqMu sync.Mutex
	var reqWg sync.WaitGroup

	for i := 0; i < maxRequests; i++ {
		reqWg.Add(1)
		go func(idx int) {
			defer reqWg.Done()
			requesterID := userIDs[rand.Intn(len(userIDs))]
			reqID, err := populateRequest(c, idx, requesterID)
			if err == nil {
				reqMu.Lock()
				reqIDs = append(reqIDs, reqID)
				reqMu.Unlock()
			}
		}(i)
	}
	reqWg.Wait()

	// 3. Random Response Phase
	// This uses the generated pools to simulate "random acceptance" counts
	log.Printf("Simulating %d Acceptances and %d Declines...", randomAcceptance, randomDeclines)

	simulateResponses(c, userIDs, reqIDs, "Accepted", randomAcceptance)
	simulateResponses(c, userIDs, reqIDs, "Declined", randomDeclines)

	log.Printf("Process completed in %v", time.Since(start))
}

func simulateResponses(c *client, userIDs, reqIDs []string, action string, count int) {
	if len(userIDs) == 0 || len(reqIDs) == 0 {
		return
	}

	for i := 0; i < count; i++ {
		randomUser := userIDs[rand.Intn(len(userIDs))]
		randomReq := reqIDs[rand.Intn(len(reqIDs))]

		c.respondToRequest(randomReq, randomUser, action)
	}
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
		return "", err
	}
	uid := res["id"].(string)

	bType := bloodTypes[rand.Intn(len(bloodTypes))]
	_ = c.updateHealth(uid, "blood_type", bType)

	lat := minLat + rand.Float64()*(maxLat-minLat)
	lng := minLng + rand.Float64()*(maxLng-minLng)
	cell, _ := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, 9)
	_ = c.updateLocation(uid, lat, lng, cell.String())

	return uid, nil
}

func populateRequest(c *client, i int, uid string) (string, error) {
	lat := minLat + rand.Float64()*(maxLat-minLat)
	lng := minLng + rand.Float64()*(maxLng-minLng)
	cell, _ := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lng}, 9)

	payload := map[string]interface{}{
		"user_id":          uid,
		"location_hex":     cell.String(),
		"location_lat":     lat,
		"location_lng":     lng,
		"bag_count":        rand.Intn(5) + 1,
		"required_by_date": time.Now().AddDate(0, 0, 3).Format(time.RFC3339),
		"blood_type":       bloodTypes[rand.Intn(len(bloodTypes))],
		"contact_phone":    "+8801900000000",
		"description":      fmt.Sprintf("Urgent request #%d", i),
		"requester_info":   "Emergency Unit",
		"location_name":    "City Hospital",
	}
	res, err := c.doRequest(http.MethodPost, "/requests", payload)
	if err != nil {
		return "", err
	}
	return res["id"].(string), nil
}

func (c *client) respondToRequest(requestID, userID, action string) {
	path := fmt.Sprintf("/requests/%s/respond", requestID)
	payload := map[string]string{
		"user_id": userID,
		"action":  action,
	}
	_, err := c.doRequest(http.MethodPost, path, payload)
	if err != nil {
		log.Printf("Response failed | Req: %s | User: %s | Action: %s | Err: %v", requestID, userID, action, err)
	} else {
		log.Printf("Response recorded | Req: %s | User: %s | Action: %s", requestID, userID, action)
	}
}

// --- Reusable HTTP logic ---

func (c *client) updateHealth(uid, infoType, details string) error {
	_, err := c.doRequest(http.MethodPut, "/users/health", map[string]string{
		"user_id": uid, "info_type": infoType, "details": details,
	})
	return err
}

func (c *client) updateLocation(uid string, lat, lng float64, hex string) error {
	_, err := c.doRequest(http.MethodPut, "/users/location", map[string]interface{}{
		"user_id": uid, "lat": lat, "lng": lng, "h3_hex": hex,
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
