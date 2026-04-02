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
