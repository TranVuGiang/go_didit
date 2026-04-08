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

// TestSubmitIDVerification_Integration requires:
//   - DIDIT_API_KEY
//   - DIDIT_FRONT_IMAGE_B64  base64-encoded front document image
//   - DIDIT_BACK_IMAGE_B64   base64-encoded back document image (optional)
func TestSubmitIDVerification_Integration(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	frontB64 := strings.TrimSpace(os.Getenv("DIDIT_FRONT_IMAGE_B64"))
	if frontB64 == "" {
		t.Skip("skipping test: DIDIT_FRONT_IMAGE_B64 not set")
	}

	frontBytes, err := godidit.DecodeBase64Image(frontB64)
	if err != nil {
		t.Fatalf("failed to decode front image: %v", err)
	}

	req := &godidit.IDVerificationRequest{
		FrontImage: frontBytes,
		VendorData: "test-integration",
	}

	if backB64 := strings.TrimSpace(os.Getenv("DIDIT_BACK_IMAGE_B64")); backB64 != "" {
		backBytes, err := godidit.DecodeBase64Image(backB64)
		if err != nil {
			t.Fatalf("failed to decode back image: %v", err)
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
