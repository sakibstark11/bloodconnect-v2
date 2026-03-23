package tests

import (
	"context"
	"testing"

	"bloodconnect/application/domain"
)

func TestDonationWaitDays(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	u1 := createTestUser(ctx, ts, "user1@example.com", domain.BloodTypeONeg, 23.8103, 90.4125)
	u2 := createTestUser(ctx, ts, "user2@example.com", domain.BloodTypeONeg, 23.8103, 90.4125)

	reqID1, _ := ts.reqService.SubmitRequest(ctx, u1, domain.BloodTypeONeg, "Need blood 1", "Contact 1", "Hospital 1", 23.8103, 90.4125, 1, domain.Now())

	ctx2 := context.WithValue(ctx, domain.UserIDKey, u2)
	err := ts.reqService.RespondToRequest(ctx2, u2, reqID1, domain.ActionStatusAccepted)
	if err != nil {
		t.Fatalf("Expected User 2 to successfully accept Request 1, got error: %v", err)
	}

	reqID2, _ := ts.reqService.SubmitRequest(ctx, u1, domain.BloodTypeOPos, "Need blood 2", "Contact 2", "Hospital 2", 23.8103, 90.4125, 1, domain.Now())

	err = ts.reqService.RespondToRequest(ctx2, u2, reqID2, domain.ActionStatusAccepted)
	if err == nil {
		t.Errorf("Expected User 2 to be REJECTED when accepting Request 2 within wait days, but it succeeded")
	}

	err = ts.reqService.RespondToRequest(ctx2, u2, reqID2, domain.ActionStatusDeclined)
	if err != nil {
		t.Errorf("Expected User 2 to be allowed to decline Request 2, but got error: %v", err)
	}

	// Now decline Request 1 and try to accept Request 2 again
	err = ts.reqService.RespondToRequest(ctx2, u2, reqID1, domain.ActionStatusDeclined)
	if err != nil {
		t.Fatalf("Expected User 2 to be allowed to decline Request 1, but got error: %v", err)
	}

	err = ts.reqService.RespondToRequest(ctx2, u2, reqID2, domain.ActionStatusAccepted)
	if err != nil {
		t.Errorf("Expected User 2 to be allowed to accept Request 2 after declining Request 1, but got error: %v", err)
	}
}

func TestUnauthorizedDonatedAction(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	u1 := createTestUser(ctx, ts, "user2@example.com", domain.BloodTypeONeg, 23.8103, 90.4125)
	u2 := createTestUser(ctx, ts, "user3@example.com", domain.BloodTypeOPos, 23.8103, 90.4125)

	reqID1, _ := ts.reqService.SubmitRequest(ctx, u1, domain.BloodTypeONeg, "Need blood 1", "Contact 1", "Hospital 1", 23.8103, 90.4125, 1, domain.Now())

	ctx2 := context.WithValue(ctx, domain.UserIDKey, u2)

	// User tries to donate for Request 1 (should fail)
	err := ts.reqService.RespondToRequest(ctx2, u2, reqID1, domain.ActionStatusDonated)
	if err == nil {
		t.Fatalf("Expected Donated action to be REJECTED, but it succeeded")
	} else if err != domain.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got: %v", err)
	}
}
