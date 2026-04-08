package godidit_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	godidit "github.com/TranVuGiang/go_didit"
)

func TestValidateDatabase_NilRequest(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).ValidateDatabase(context.Background(), nil)
	if !errors.Is(err, godidit.ErrNilRequest) {
		t.Fatalf("expected ErrNilRequest, got: %v", err)
	}
}

func TestValidateDatabase_EmptyNationalID(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).ValidateDatabase(context.Background(), &godidit.DatabaseValidationRequest{
		IssuingState: "BRA",
	})
	if !errors.Is(err, godidit.ErrEmptyNationalID) {
		t.Fatalf("expected ErrEmptyNationalID, got: %v", err)
	}
}

func TestValidateDatabase_EmptyIssuingState(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).ValidateDatabase(context.Background(), &godidit.DatabaseValidationRequest{
		IdentificationNumber: "12345678900",
	})
	if !errors.Is(err, godidit.ErrInvalidParams) {
		t.Fatalf("expected ErrInvalidParams, got: %v", err)
	}
}

// TestValidateDatabase_Integration requires:
//   - DIDIT_API_KEY
//   - DIDIT_DB_ISSUING_STATE  (ISO alpha-3, e.g. "BRA")
//   - DIDIT_DB_ID_NUMBER      (national ID number)
func TestValidateDatabase_Integration(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	issuingState := strings.TrimSpace(os.Getenv("DIDIT_DB_ISSUING_STATE"))
	idNumber := strings.TrimSpace(os.Getenv("DIDIT_DB_ID_NUMBER"))
	if issuingState == "" || idNumber == "" {
		t.Skip("skipping test: DIDIT_DB_ISSUING_STATE or DIDIT_DB_ID_NUMBER not set")
	}

	req := &godidit.DatabaseValidationRequest{
		IssuingState:         issuingState,
		IdentificationNumber: idNumber,
		VendorData:           "test-integration",
	}

	resp, err := client.ValidateDatabase(context.Background(), req)
	if err != nil {
		t.Fatalf("ValidateDatabase failed: %v", err)
	}

	if resp.RequestID == "" {
		t.Error("expected non-empty request_id")
	}

	t.Logf("database validation: request_id=%s status=%s match_type=%s",
		resp.RequestID, resp.DatabaseValidation.Status, resp.DatabaseValidation.MatchType)
}
