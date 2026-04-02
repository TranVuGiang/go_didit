package godidit_test

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRetrieveSession(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	sessionID := strings.TrimSpace(os.Getenv("DIDIT_SESSION_ID"))
	if sessionID == "" {
		t.Skip("skipping test: DIDIT_SESSION_ID not set")
	}

	decision, err := client.RetrieveSession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("RetrieveSession failed: %v", err)
	}

	if decision.SessionID == "" {
		t.Error("expected non-empty session_id")
	}

	t.Logf("session decision: id=%s status=%s", decision.SessionID, decision.Status)
}
