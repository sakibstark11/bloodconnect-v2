package tests

import (
	"context"
	"testing"

	"bloodconnect/application/domain"
)

func TestBloodTypeCompatibility(t *testing.T) {
	ctx := context.Background()
	ts := setupTestSuite(t)

	// User 1: B+
	u1 := createTestUser(ctx, ts, "donor_b_pos@example.com", domain.BloodTypeBPos, 23.8103, 90.4125)
	
	// Request 1: A+
	reqID1, _ := ts.reqService.SubmitRequest(ctx, domain.UserID("requester"), domain.BloodTypeAPos, "Need A+", "Contact", "Hospital", 23.8103, 90.4125, 1, domain.Now())

	ctx1 := context.WithValue(ctx, domain.UserIDKey, u1)
	
	// B+ cannot donate to A+
	err := ts.reqService.RespondToRequest(ctx1, reqID1, domain.ActionStatusAccepted)
	if err == nil {
		t.Errorf("Expected B+ donor to be REJECTED for A+ request, but it succeeded")
	} else {
		t.Logf("Correctly rejected incompatible donor: %v", err)
	}

	// User 2: O-
	u2 := createTestUser(ctx, ts, "donor_o_neg@example.com", domain.BloodTypeONeg, 23.8103, 90.4125)
	ctx2 := context.WithValue(ctx, domain.UserIDKey, u2)

	// O- can donate to A+
	err = ts.reqService.RespondToRequest(ctx2, reqID1, domain.ActionStatusAccepted)
	if err != nil {
		t.Errorf("Expected O- donor to be ACCEPTED for A+ request, but got error: %v", err)
	}
}
