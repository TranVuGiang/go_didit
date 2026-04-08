package godidit_test

import (
	"context"
	"errors"
	"testing"

	godidit "github.com/TranVuGiang/go_didit"
)

func TestScreenAML_NilRequest(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).ScreenAML(context.Background(), nil)
	if !errors.Is(err, godidit.ErrNilRequest) {
		t.Fatalf("expected ErrNilRequest, got: %v", err)
	}
}

func TestScreenAML_EmptyFullName(t *testing.T) {
	t.Parallel()

	_, err := newDummyClient(t).ScreenAML(context.Background(), &godidit.AMLScreeningRequest{FullName: "  "})
	if !errors.Is(err, godidit.ErrEmptyFullName) {
		t.Fatalf("expected ErrEmptyFullName, got: %v", err)
	}
}

func TestScreenAML_Integration(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	req := &godidit.AMLScreeningRequest{
		FullName:    "John Doe",
		VendorData:  "test-integration",
	}

	resp, err := client.ScreenAML(context.Background(), req)
	if err != nil {
		t.Fatalf("ScreenAML failed: %v", err)
	}

	if resp.RequestID == "" {
		t.Error("expected non-empty request_id")
	}

	t.Logf("aml screening: request_id=%s status=%s total_hits=%d score=%d",
		resp.RequestID, resp.AML.Status, resp.AML.TotalHits, resp.AML.Score)
}
