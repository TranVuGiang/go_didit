package godidit_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	godidit "github.com/TranVuGiang/go_didit"
)

func TestSubmitIDVerification_NilRequest(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).SubmitIDVerification(context.Background(), nil)
	if !errors.Is(err, godidit.ErrNilRequest) {
		t.Fatalf("expected ErrNilRequest, got: %v", err)
	}
}

func TestSubmitIDVerification_EmptyFrontImage(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).SubmitIDVerification(context.Background(), &godidit.IDVerificationRequest{})
	if !errors.Is(err, godidit.ErrEmptyFrontImage) {
		t.Fatalf("expected ErrEmptyFrontImage, got: %v", err)
	}
}

// TestSubmitIDVerification_Integration requires DIDIT_FRONT_IMAGE_URL env var
// pointing to a publicly accessible (or presigned) image URL.
func TestSubmitIDVerification_Integration(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	frontURL := strings.TrimSpace(os.Getenv("DIDIT_FRONT_IMAGE_URL"))
	if frontURL == "" {
		t.Skip("skipping test: DIDIT_FRONT_IMAGE_URL not set")
	}

	// Download image to simulate what VerifyKyc does internally.
	frontBytes, err := downloadImageHelper(t, frontURL)
	if err != nil {
		t.Fatalf("failed to download front image: %v", err)
	}

	req := &godidit.IDVerificationRequest{
		FrontImage: frontBytes,
		VendorData: "test-integration",
	}

	backURL := strings.TrimSpace(os.Getenv("DIDIT_BACK_IMAGE_URL"))
	if backURL != "" {
		backBytes, err := downloadImageHelper(t, backURL)
		if err != nil {
			t.Fatalf("failed to download back image: %v", err)
		}
		req.BackImage = backBytes
	}

	resp, err := client.SubmitIDVerification(context.Background(), req)
	if err != nil {
		t.Fatalf("SubmitIDVerification failed: %v", err)
	}

	if resp.RequestID == "" {
		t.Error("expected non-empty request_id")
	}

	t.Logf("id verification: request_id=%s status=%s name=%s",
		resp.RequestID, resp.IDVerification.Status, resp.IDVerification.FullName)
}
