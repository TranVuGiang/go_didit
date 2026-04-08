package godidit_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	godidit "github.com/TranVuGiang/go_didit"
)

// mockHTTPClient implements godidit.HTTPClient for unit tests.
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// --- Input validation tests (no network needed) ---

func TestVerifyKyc_NilInput(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).VerifyKyc(context.Background(), nil)
	if !errors.Is(err, godidit.ErrNilRequest) {
		t.Fatalf("expected ErrNilRequest, got: %v", err)
	}
}

func TestVerifyKyc_NoApplicableData(t *testing.T) {
	t.Parallel()

	// KycInfo with all optional fields nil — nothing to verify.
	_, err := newDummyClient(t).VerifyKyc(context.Background(), &godidit.KycInfo{ID: 1})
	if !errors.Is(err, godidit.ErrNoVerificationPerformed) {
		t.Fatalf("expected ErrNoVerificationPerformed, got: %v", err)
	}
}

// --- Unit tests with mock HTTP client ---

func TestVerifyKyc_AMLOnly(t *testing.T) {
	t.Parallel()

	const amlResponse = `{
		"request_id": "req-aml-001",
		"aml": {"status": "Approved", "entity_type": "person", "total_hits": 0, "score": 100},
		"created_at": "2026-04-08T00:00:00Z"
	}`

	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, amlResponse), nil
		},
	}

	firstName := "Nguyen"
	lastName := "Van A"
	kyc := &godidit.KycInfo{
		FirstName: &firstName,
		LastName:  &lastName,
		// No FrontIdImage, no NationalID → only AML is triggered.
	}

	client, err := godidit.NewClient("test-api-key", godidit.WithHTTPClient(mock))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.VerifyKyc(context.Background(), kyc)
	if err != nil {
		t.Fatalf("VerifyKyc failed: %v", err)
	}

	if result.AMLScreening == nil {
		t.Fatal("expected AMLScreening result, got nil")
	}
	if result.AMLScreening.RequestID != "req-aml-001" {
		t.Errorf("unexpected request_id: %s", result.AMLScreening.RequestID)
	}
	if result.IDVerification != nil {
		t.Error("expected IDVerification to be nil (not triggered)")
	}
	if result.DatabaseValidation != nil {
		t.Error("expected DatabaseValidation to be nil (not triggered)")
	}
}

func TestVerifyKyc_FailFast_OnAPIError(t *testing.T) {
	t.Parallel()

	// Mock returns HTTP 403 for every call.
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusForbidden, `{"message":"insufficient credits","code":403}`), nil
		},
	}

	firstName := "Nguyen"
	lastName := "Van A"
	nationalID := "12345678"
	nationality := "VNM"
	kyc := &godidit.KycInfo{
		FirstName:   &firstName,
		LastName:    &lastName,
		NationalID:  &nationalID,
		Nationality: &nationality,
	}

	client, err := godidit.NewClient("test-api-key", godidit.WithHTTPClient(mock))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.VerifyKyc(context.Background(), kyc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}

	var apiErr *godidit.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got: %T %v", err, err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
}

func TestVerifyKyc_FailFast_CancelsOtherGoroutines(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			callCount.Add(1)
			// First call fails, subsequent calls check context cancellation.
			if callCount.Load() == 1 {
				return jsonResponse(http.StatusInternalServerError,
					`{"message":"server error","code":500}`), nil
			}
			// If context was cancelled, return context error.
			return nil, fmt.Errorf("context cancelled")
		},
	}

	firstName := "Nguyen"
	lastName := "Van A"
	nationalID := "12345678"
	nationality := "VNM"
	kyc := &godidit.KycInfo{
		FirstName:   &firstName,
		LastName:    &lastName,
		NationalID:  &nationalID,
		Nationality: &nationality,
	}

	client, err := godidit.NewClient("test-api-key", godidit.WithHTTPClient(mock))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.VerifyKyc(context.Background(), kyc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Integration test ---

// TestVerifyKyc_Integration requires:
//   - DIDIT_API_KEY
//   - DIDIT_KYC_FIRST_NAME, DIDIT_KYC_LAST_NAME
//
// Optional:
//   - DIDIT_KYC_NATIONAL_ID, DIDIT_KYC_NATIONALITY (triggers DatabaseValidation)
//   - DIDIT_FRONT_IMAGE_URL                         (triggers IDVerification)
func TestVerifyKyc_Integration(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	firstName := strings.TrimSpace(os.Getenv("DIDIT_KYC_FIRST_NAME"))
	lastName := strings.TrimSpace(os.Getenv("DIDIT_KYC_LAST_NAME"))
	if firstName == "" || lastName == "" {
		t.Skip("skipping test: DIDIT_KYC_FIRST_NAME or DIDIT_KYC_LAST_NAME not set")
	}

	kyc := &godidit.KycInfo{
		FirstName: &firstName,
		LastName:  &lastName,
	}

	if v := strings.TrimSpace(os.Getenv("DIDIT_KYC_NATIONAL_ID")); v != "" {
		kyc.NationalID = &v
	}
	if v := strings.TrimSpace(os.Getenv("DIDIT_KYC_NATIONALITY")); v != "" {
		kyc.Nationality = &v
	}
	if v := strings.TrimSpace(os.Getenv("DIDIT_FRONT_IMAGE_URL")); v != "" {
		kyc.FrontIdImage = &v
	}
	if v := strings.TrimSpace(os.Getenv("DIDIT_BACK_IMAGE_URL")); v != "" {
		kyc.BackIdImage = &v
	}

	result, err := client.VerifyKyc(context.Background(), kyc)
	if err != nil {
		t.Fatalf("VerifyKyc failed: %v", err)
	}

	if result.AMLScreening != nil {
		t.Logf("aml: status=%s total_hits=%d score=%d",
			result.AMLScreening.AML.Status,
			result.AMLScreening.AML.TotalHits,
			result.AMLScreening.AML.Score)
	}
	if result.IDVerification != nil {
		t.Logf("id_verification: status=%s name=%s dob=%s",
			result.IDVerification.IDVerification.Status,
			result.IDVerification.IDVerification.FullName,
			result.IDVerification.IDVerification.DateOfBirth)
	}
	if result.DatabaseValidation != nil {
		t.Logf("database_validation: status=%s match_type=%s",
			result.DatabaseValidation.DatabaseValidation.Status,
			result.DatabaseValidation.DatabaseValidation.MatchType)
	}
}
