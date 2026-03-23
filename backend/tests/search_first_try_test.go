package tests

import (
	"context"
	"testing"

	"bloodconnect/application/domain"
)

func TestSearchFirstTrySuccess(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	u1 := createTestUser(ctx, ts, "user1@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)
	u2 := createTestUser(ctx, ts, "user2@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)
	u3 := createTestUser(ctx, ts, "user3@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)

	uReq := createTestUser(ctx, ts, "requester@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)
	bagCount := 2
	reqID, err := ts.reqService.SubmitRequest(ctx, uReq, domain.BloodTypeOPos, "Need blood", "Contact info", "Hospital", 23.8103, 90.4125, bagCount, domain.Now())
	if err != nil {
		t.Fatalf("Failed to submit request: %v", err)
	}

	runWorkerOnce(ctx, ts)

	n1, _ := ts.notifRepo.GetNotificationsForUser(ctx, u1, "", 10)
	n2, _ := ts.notifRepo.GetNotificationsForUser(ctx, u2, "", 10)
	n3, _ := ts.notifRepo.GetNotificationsForUser(ctx, u3, "", 10)

	notifiedCount := 0
	if len(n1) > 0 {
		notifiedCount++
	}
	if len(n2) > 0 {
		notifiedCount++
	}
	if len(n3) > 0 {
		notifiedCount++
	}

	if notifiedCount != 2 {
		t.Errorf("Expected 2 users to be notified, got %d", notifiedCount)
	}

	ctx1 := context.WithValue(ctx, domain.UserIDKey, u1)
	ctx2 := context.WithValue(ctx, domain.UserIDKey, u2)
	ctx3 := context.WithValue(ctx, domain.UserIDKey, u3)

	if len(n1) > 0 {
		_ = ts.reqService.RespondToRequest(ctx1, u1, reqID, domain.ActionStatusAccepted)
	}
	if len(n2) > 0 {
		_ = ts.reqService.RespondToRequest(ctx2, u2, reqID, domain.ActionStatusAccepted)
	}
	if len(n3) > 0 {
		_ = ts.reqService.RespondToRequest(ctx3, u3, reqID, domain.ActionStatusAccepted)
	}

	actioned, _ := ts.reqRepo.GetActionedUsers(ctx, reqID)
	acceptedCount := 0
	for _, a := range actioned {
		if a.Action == domain.ActionStatusAccepted {
			acceptedCount++
		}
	}

	if acceptedCount != 2 {
		t.Errorf("Expected 2 accepted users, got %d", acceptedCount)
	}
}
