package tests

import (
	"context"
	"testing"
	"time"

	"bloodconnect/application/domain"
)

func TestSearchExhausted(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	uReq := createTestUser(ctx, ts, "requester@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)
	bagCount := 2
	reqID, err := ts.reqService.SubmitRequest(ctx, uReq, domain.BloodTypeOPos, "Need blood", "Contact info", "Hospital", 23.8103, 90.4125, bagCount, domain.Now())
	if err != nil {
		t.Fatalf("Failed to submit request: %v", err)
	}

	maxAttempts := 300
	for i := 0; i < maxAttempts; i++ {
		runWorkerOnce(ctx, ts)
		req, _ := ts.reqRepo.GetRequestByID(ctx, reqID)
		if req.Status == domain.RequestStatusFailed {
			return
		}
		time.Sleep(ts.config.WaveSearchInterval / 2)
	}

	t.Errorf("Request did not transition to Failed status after %d worker runs", maxAttempts)
}
