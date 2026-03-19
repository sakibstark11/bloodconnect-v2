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
	baseURL          = "http://localhost:8080"
	maxUsers         = 100
	maxRequests      = 20
	numWorkers       = 5
	randomAcceptance = 10
	randomDeclines   = 15
)

var (
	bloodTypes     = []string{"A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-"}
	minLat, maxLat = 23.68, 23.90
	minLng, maxLng = 90.33, 90.50
)

type userSession struct {
	ID    string
	Token string
}

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

	// 1. Generate Users & Login
	sessionsChan := make(chan userSession, maxUsers)
	var userWg sync.WaitGroup
	jobs := make(chan int, maxUsers)

	for w := 1; w <= numWorkers; w++ {
		userWg.Add(1)
		go func() {
			defer userWg.Done()
			for i := range jobs {
				session, err := populateUser(c, i)
				if err == nil {
					sessionsChan <- session
				}
			}
		}()
	}

	for i := 0; i < maxUsers; i++ {
		jobs <- i
	}
	close(jobs)
	userWg.Wait()
	close(sessionsChan)

	var sessions []userSession
	for s := range sessionsChan {
		sessions = append(sessions, s)
	}

	if len(sessions) == 0 {
		log.Fatal("No users were created. Is the server running?")
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
			session := sessions[rand.Intn(len(sessions))]
			reqID, err := populateRequest(c, idx, session)
			if err != nil {
				log.Printf("Request failed | User: %s | Err: %v", session.ID, err)
			} else {
				reqMu.Lock()
				reqIDs = append(reqIDs, reqID)
				reqMu.Unlock()
			}
		}(i)
	}
	reqWg.Wait()

	// 3. Random Response Phase
	log.Printf("Simulating %d Acceptances and %d Declines...", randomAcceptance, randomDeclines)

	simulateResponses(c, sessions, reqIDs, "Accepted", randomAcceptance)
	simulateResponses(c, sessions, reqIDs, "Declined", randomDeclines)

	log.Printf("Process completed in %v", time.Since(start))
}

func simulateResponses(c *client, sessions []userSession, reqIDs []string, action string, count int) {
	if len(sessions) == 0 || len(reqIDs) == 0 {
		return
	}

	for i := 0; i < count; i++ {
		session := sessions[rand.Intn(len(sessions))]
		randomReq := reqIDs[rand.Intn(len(reqIDs))]

		c.respondToRequest(randomReq, session, action)
	}
}

func populateUser(c *client, index int) (userSession, error) {
	suffix := time.Now().UnixNano() % 100000
	email := fmt.Sprintf("user%d_%d@example.com", index, suffix)
	password := "password123"
	phone := fmt.Sprintf("+8801%d%04d", suffix%100, index%10000)

	// Signup
	signupRes, err := c.doRequest(http.MethodPost, "/users/signup", "", map[string]string{
		"name":     fmt.Sprintf("User %d", index),
		"email":    email,
		"password": password,
		"phone":    phone,
	})
	if err != nil {
		return userSession{}, err
	}
	uid := signupRes["id"].(string)

	// Login to get token
	loginRes, err := c.doRequest(http.MethodPost, "/users/login", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return userSession{}, err
	}
	token := loginRes["token"].(string)
	session := userSession{ID: uid, Token: token}

	// Update Health & Location using token
	bType := bloodTypes[rand.Intn(len(bloodTypes))]
	_ = c.updateHealth(session, "blood_type", bType)

	lat := minLat + rand.Float64()*(maxLat-minLat)
	lng := minLng + rand.Float64()*(maxLng-minLng)
	_ = c.updateLocation(session, lat, lng)

	return session, nil
}

func populateRequest(c *client, i int, session userSession) (string, error) {
	lat := minLat + rand.Float64()*(maxLat-minLat)
	lng := minLng + rand.Float64()*(maxLng-minLng)

	payload := map[string]interface{}{
		"location_lat":     lat,
		"location_lng":     lng,
		"bag_count":        rand.Intn(5) + 1,
		"required_by_date": time.Now().UTC().Add(7*24*time.Hour + 1*time.Minute).Format("2006-01-02T15:04:05.000Z"),
		"blood_type":       bloodTypes[rand.Intn(len(bloodTypes))],
		"description":      fmt.Sprintf("Urgent request #%d", i),
		"requester_info":   "Emergency Unit",
		"location_name":    "City Hospital",
	}
	res, err := c.doRequest(http.MethodPost, "/requests", session.Token, payload)
	if err != nil {
		return "", err
	}
	return res["id"].(string), nil
}

func (c *client) respondToRequest(requestID string, session userSession, action string) {
	path := fmt.Sprintf("/requests/%s/respond", requestID)
	payload := map[string]string{
		"action": action,
	}
	_, err := c.doRequest(http.MethodPost, path, session.Token, payload)
	if err != nil {
		log.Printf("Response failed | Req: %s | User: %s | Action: %s | Err: %v", requestID, session.ID, action, err)
	} else {
		log.Printf("Response recorded | Req: %s | User: %s | Action: %s", requestID, session.ID, action)
	}
}

func (c *client) updateHealth(session userSession, infoType, details string) error {
	_, err := c.doRequest(http.MethodPut, "/users/me/health", session.Token, map[string]string{
		"info_type": infoType, "details": details,
	})
	return err
}

func (c *client) updateLocation(session userSession, lat, lng float64) error {
	_, err := c.doRequest(http.MethodPut, "/users/me/location", session.Token, map[string]interface{}{
		"lat": lat, "lng": lng,
	})
	return err
}

func (c *client) doRequest(method, path, token string, body interface{}) (map[string]interface{}, error) {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, baseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

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
