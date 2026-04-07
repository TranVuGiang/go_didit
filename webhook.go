package godidit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

const webhookTimestampTolerance = 5 * time.Minute

var (
	ErrWebhookMissingTimestamp      = errors.New("missing X-Timestamp header")
	ErrWebhookMissingSignature      = errors.New("missing X-Signature-V2 header")
	ErrWebhookTimestampExpired      = errors.New("webhook timestamp is outside the 5-minute tolerance")
	ErrWebhookInvalidSignature      = errors.New("webhook signature mismatch")
	ErrWebhookSimpleMissingFields   = errors.New("missing fields for simple signature verification")
	ErrWebhookSimpleInvalidSignature = errors.New("webhook simple signature mismatch")
)

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of a Didit webhook request.
//
// Parameters:
//   - secret: the DIDIT_WEBHOOK_SECRET from your Didit Console
//   - rawBody: the raw (unparsed) request body bytes
//   - signature: value of the X-Signature-V2 header
//   - timestampHeader: value of the X-Timestamp header (Unix seconds as string)
//
// Returns nil if the signature is valid and the timestamp is fresh.
func VerifyWebhookSignature(secret string, rawBody []byte, signature, timestampHeader string) error {
	if timestampHeader == "" {
		return ErrWebhookMissingTimestamp
	}

	if signature == "" {
		return ErrWebhookMissingSignature
	}

	ts, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid X-Timestamp value: %w", err)
	}

	age := time.Duration(math.Abs(float64(time.Now().Unix()-ts))) * time.Second
	if age > webhookTimestampTolerance {
		return ErrWebhookTimestampExpired
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrWebhookInvalidSignature
	}

	return nil
}

// VerifySimpleWebhookSignature verifies the X-Signature-Simple header.
// This is useful when middleware (e.g. Cloudflare, ngrok) re-encodes the JSON body,
// which breaks X-Signature-V2. The simple signature is computed over:
// "{timestamp}:{session_id}:{status}:{webhook_type}"
func VerifySimpleWebhookSignature(secret, sessionID, status, webhookType, timestamp, signature string) error {
	if timestamp == "" || signature == "" || sessionID == "" {
		return ErrWebhookSimpleMissingFields
	}

	message := fmt.Sprintf("%s:%s:%s:%s", timestamp, sessionID, status, webhookType)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrWebhookSimpleInvalidSignature
	}

	return nil
}
