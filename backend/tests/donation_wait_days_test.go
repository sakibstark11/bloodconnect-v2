package tests

import (
	"context"
	"testing"

	"bloodconnect/application/domain"
)

func TestDonationWaitDays(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	u1 := createTestUser(ctx, ts, "user1@example.com", domain.BloodTypeBNeg, 23.8103, 90.4125)

	reqID1, _ := ts.reqService.SubmitRequest(ctx, u1, domain.BloodTypeBNeg, "Need blood 1", "Contact 1", "Hospital 1", 23.8103, 90.4125, 1, domain.Now())

	ctx1 := context.WithValue(ctx, domain.UserIDKey, u1)
	err := ts.reqService.RespondToRequest(ctx1, reqID1, domain.ActionStatusAccepted)
	if err != nil {
		t.Fatalf("Expected User 1 to successfully accept Request 1, got error: %v", err)
	}

	reqID2, _ := ts.reqService.SubmitRequest(ctx, domain.UserID("requester2"), domain.BloodTypeOPos, "Need blood 2", "Contact 2", "Hospital 2", 23.8103, 90.4125, 1, domain.Now())

	err = ts.reqService.RespondToRequest(ctx1, reqID2, domain.ActionStatusAccepted)
	if err == nil {
		t.Errorf("Expected User 1 to be REJECTED when accepting Request 2 within wait days, but it succeeded")
	}

	err = ts.reqService.RespondToRequest(ctx1, reqID2, domain.ActionStatusDeclined)
	if err != nil {
		t.Errorf("Expected User 1 to be allowed to decline Request 2, but got error: %v", err)
	}
}
