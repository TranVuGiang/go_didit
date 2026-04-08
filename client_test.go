package godidit_test

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	godidit "github.com/TranVuGiang/go_didit"
)

func newTestClient(t *testing.T) *godidit.Client {
	t.Helper()

	apiKey := strings.TrimSpace(os.Getenv("DIDIT_API_KEY"))
	if apiKey == "" {
		t.Skip("skipping test: DIDIT_API_KEY not set")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	client, err := godidit.NewClient(
		apiKey,
		godidit.WithDebug(true),
		godidit.WithLogger(logger),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return client
}

// newDummyClient creates a client with a fake API key for validation-only tests
// that never reach the network.
func newDummyClient(t *testing.T, opts ...godidit.Option) *godidit.Client {
	t.Helper()
	c, err := godidit.NewClient("dummy-api-key-for-validation", opts...)
	if err != nil {
		t.Fatalf("newDummyClient: %v", err)
	}
	return c
}
