package tests

import (
	"context"
	"testing"
	"time"

	"bloodconnect/application/domain"
)

func TestRevocationTrigger(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	u1 := createTestUser(ctx, ts, "user1@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)
	u2 := createTestUser(ctx, ts, "user2@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)
	u3 := createTestUser(ctx, ts, "user3@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)

	bagCount := 2
	reqID, err := ts.reqService.SubmitRequest(ctx, domain.UserID("requester_id"), domain.BloodTypeOPos, "Need blood", "Contact info", "Hospital", 23.8103, 90.4125, bagCount, domain.Now())
	if err != nil {
		t.Fatalf("Failed to submit request: %v", err)
	}

	runWorkerOnce(ctx, ts)

	ctx1 := context.WithValue(ctx, domain.UserIDKey, u1)
	ctx2 := context.WithValue(ctx, domain.UserIDKey, u2)
	_ = ts.reqService.RespondToRequest(ctx1, reqID, domain.ActionStatusAccepted)
	_ = ts.reqService.RespondToRequest(ctx2, reqID, domain.ActionStatusAccepted)

	_ = ts.reqService.RespondToRequest(ctx1, reqID, domain.ActionStatusDeclined)

	time.Sleep(100 * time.Millisecond)
	runWorkerOnce(ctx, ts)

	n3, _ := ts.notifRepo.GetNotificationsForUser(ctx, u3, "", 10)
	if len(n3) == 0 {
		t.Errorf("Expected User 3 to be notified after User 1 revoked their acceptance")
	}
}
