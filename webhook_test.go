package godidit_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	godidit "github.com/TranVuGiang/go_didit"
)

const testWebhookSecret = "test-secret"

func validSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func freshTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	t.Parallel()

	body := []byte(`{"session_id":"abc","status":"Approved"}`)
	sig := validSignature(testWebhookSecret, body)

	err := godidit.VerifyWebhookSignature(testWebhookSecret, body, sig, freshTimestamp())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestVerifyWebhookSignature_MissingTimestamp(t *testing.T) {
	t.Parallel()

	body := []byte(`{"session_id":"abc"}`)
	sig := validSignature(testWebhookSecret, body)

	err := godidit.VerifyWebhookSignature(testWebhookSecret, body, sig, "")
	if !errors.Is(err, godidit.ErrWebhookMissingTimestamp) {
		t.Fatalf("expected ErrWebhookMissingTimestamp, got: %v", err)
	}
}

func TestVerifyWebhookSignature_MissingSignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"session_id":"abc"}`)

	err := godidit.VerifyWebhookSignature(testWebhookSecret, body, "", freshTimestamp())
	if !errors.Is(err, godidit.ErrWebhookMissingSignature) {
		t.Fatalf("expected ErrWebhookMissingSignature, got: %v", err)
	}
}

func TestVerifyWebhookSignature_ExpiredTimestamp(t *testing.T) {
	t.Parallel()

	body := []byte(`{"session_id":"abc"}`)
	sig := validSignature(testWebhookSecret, body)
	oldTimestamp := fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())

	err := godidit.VerifyWebhookSignature(testWebhookSecret, body, sig, oldTimestamp)
	if !errors.Is(err, godidit.ErrWebhookTimestampExpired) {
		t.Fatalf("expected ErrWebhookTimestampExpired, got: %v", err)
	}
}

func TestVerifyWebhookSignature_InvalidSignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"session_id":"abc"}`)

	err := godidit.VerifyWebhookSignature(testWebhookSecret, body, "invalidsignature", freshTimestamp())
	if !errors.Is(err, godidit.ErrWebhookInvalidSignature) {
		t.Fatalf("expected ErrWebhookInvalidSignature, got: %v", err)
	}
}

func TestVerifyWebhookSignature_WrongSecret(t *testing.T) {
	t.Parallel()

	body := []byte(`{"session_id":"abc"}`)
	sig := validSignature("wrong-secret", body)

	err := godidit.VerifyWebhookSignature(testWebhookSecret, body, sig, freshTimestamp())
	if !errors.Is(err, godidit.ErrWebhookInvalidSignature) {
		t.Fatalf("expected ErrWebhookInvalidSignature, got: %v", err)
	}
}

func TestVerifyWebhookSignature_InvalidTimestampFormat(t *testing.T) {
	t.Parallel()

	body := []byte(`{"session_id":"abc"}`)
	sig := validSignature(testWebhookSecret, body)

	err := godidit.VerifyWebhookSignature(testWebhookSecret, body, sig, "not-a-number")
	if err == nil {
		t.Fatal("expected error for invalid timestamp format, got nil")
	}
}

func TestVerifyWebhookSignature_WhitespaceMatters(t *testing.T) {
	t.Parallel()

	// Signature is computed over the exact raw body — whitespace changes must invalidate it.
	original := []byte(`{"session_id":"abc","status":"Approved"}`)
	reformatted := []byte(`{ "session_id": "abc", "status": "Approved" }`)

	sig := validSignature(testWebhookSecret, original)

	err := godidit.VerifyWebhookSignature(testWebhookSecret, reformatted, sig, freshTimestamp())
	if !errors.Is(err, godidit.ErrWebhookInvalidSignature) {
		t.Fatalf("expected ErrWebhookInvalidSignature when body whitespace differs, got: %v", err)
	}
}
