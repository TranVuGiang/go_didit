package godidit_test

import (
	"context"
	"os"
	"strings"
	"testing"

	godidit "github.com/TranVuGiang/go_didit"
)

func TestCreateSession(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	workflowID := strings.TrimSpace(os.Getenv("DIDIT_WORKFLOW_ID"))
	if workflowID == "" {
		t.Skip("skipping test: DIDIT_WORKFLOW_ID not set")
	}

	req := &godidit.CreateSessionRequest{
		WorkflowID: workflowID,
		VendorData: "test-vendor-data",
	}

	resp, err := client.CreateSession(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if resp.SessionID == "" {
		t.Error("expected non-empty session_id")
	}

	t.Logf("session created: id=%s status=%s url=%s", resp.SessionID, resp.Status, resp.URL)
}
