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

	// Now decline Request 1 and try to accept Request 2 again
	err = ts.reqService.RespondToRequest(ctx1, reqID1, domain.ActionStatusDeclined)
	if err != nil {
		t.Fatalf("Expected User 1 to be allowed to decline Request 1, but got error: %v", err)
	}

	err = ts.reqService.RespondToRequest(ctx1, reqID2, domain.ActionStatusAccepted)
	if err != nil {
		t.Errorf("Expected User 1 to be allowed to accept Request 2 after declining Request 1, but got error: %v", err)
	}
}

func TestDonatedActionBlocksUser(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	u1 := createTestUser(ctx, ts, "user2@example.com", domain.BloodTypeONeg, 23.8103, 90.4125)
	reqID1, _ := ts.reqService.SubmitRequest(ctx, u1, domain.BloodTypeONeg, "Need blood 1", "Contact 1", "Hospital 1", 23.8103, 90.4125, 1, domain.Now())

	ctx1 := context.WithValue(ctx, domain.UserIDKey, u1)

	// User donates for Request 1
	err := ts.reqService.RespondToRequest(ctx1, reqID1, domain.ActionStatusDonated)
	if err != nil {
		t.Fatalf("Expected User 1 to successfully mark Request 1 as Donated, got error: %v", err)
	}

	// Verify UserHealth is updated
	health, err := ts.userRepo.GetUserHealth(ctx, u1)
	if err != nil {
		t.Fatalf("Failed to get user health: %v", err)
	}
	found := false
	for _, h := range health {
		if h.InfoType == domain.InfoTypeLastDonation {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected UserHealth to have last_donation_date after Donated action")
	}

	// Now try to accept another request
	reqID2, _ := ts.reqService.SubmitRequest(ctx, domain.UserID("requester2"), domain.BloodTypeOPos, "Need blood 2", "Contact 2", "Hospital 2", 23.8103, 90.4125, 1, domain.Now())

	err = ts.reqService.RespondToRequest(ctx1, reqID2, domain.ActionStatusAccepted)
	if err == nil {
		t.Errorf("Expected User 1 to be BLOCKED after donating, but it succeeded")
	} else if err != domain.ErrDonationWaitPeriodNotMet {
		t.Errorf("Expected ErrDonationWaitPeriodNotMet, got: %v", err)
	}
}
