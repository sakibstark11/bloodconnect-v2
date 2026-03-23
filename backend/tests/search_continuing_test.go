package tests

import (
	"context"
	"testing"
	"time"

	"bloodconnect/application/domain"
)

func TestSearchContinuing(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	u1 := createTestUser(ctx, ts, "userR2_1@example.com", domain.BloodTypeOPos, 23.8123, 90.4145)
	u2 := createTestUser(ctx, ts, "userR2_2@example.com", domain.BloodTypeOPos, 23.8123, 90.4145)

	uReq := createTestUser(ctx, ts, "requester@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)
	bagCount := 2
	_, err := ts.reqService.SubmitRequest(ctx, uReq, domain.BloodTypeOPos, "Need blood", "Contact info", "Hospital", 23.8103, 90.4125, bagCount, domain.Now())
	if err != nil {
		t.Fatalf("Failed to submit request: %v", err)
	}

	runWorkerOnce(ctx, ts)

	n1, _ := ts.notifRepo.GetNotificationsForUser(ctx, u1, "", 10)
	if len(n1) > 0 {
		t.Errorf("Expected no notifications for u1 after Ring 1 search")
	}

	time.Sleep(ts.config.WaveSearchInterval + (50 * time.Millisecond))
	runWorkerOnce(ctx, ts)

	n1, _ = ts.notifRepo.GetNotificationsForUser(ctx, u1, "", 10)
	n2, _ := ts.notifRepo.GetNotificationsForUser(ctx, u2, "", 10)

	if len(n1) == 0 || len(n2) == 0 {
		t.Errorf("Expected both users in Ring 2 to be notified, got u1:%d, u2:%d", len(n1), len(n2))
	}
}
