package godidit

import (
	"errors"
	"fmt"
)

var (
	// Client construction errors.
	ErrEmptyAPIKey = errors.New("api key cannot be empty")

	// Request lifecycle errors.
	ErrNilRequest    = errors.New("request is nil")
	ErrRequestEncode = errors.New("failed to encode request body")
	ErrInvalidParams = errors.New("invalid request params")

	// HTTP / transport errors.
	ErrHTTPFailure      = errors.New("http request failed")
	ErrUnexpectedStatus = errors.New("unexpected http status code")

	// Validation errors.
	ErrEmptySessionID  = errors.New("sessionId is required")
	ErrEmptyWorkflowID = errors.New("workflow_id is required")

	// KYC verification errors.
	ErrEmptyFrontImage        = errors.New("front_image is required for id verification")
	ErrEmptyFullName          = errors.New("full_name is required for AML screening")
	ErrEmptyNationalID        = errors.New("identification_number is required for database validation")
	ErrImageDecodeFailed      = errors.New("failed to decode base64 image")
	ErrNoVerificationPerformed = errors.New("no verification could be performed: insufficient kyc data provided")
)

// APIError is returned when the server responds with a 4xx/5xx status code.
// Callers can inspect it via errors.As.
type APIError struct {
	StatusCode int
	Message    string `json:"message"`
	Code       int    `json:"code"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("didit api error: status=%d code=%d message=%s", e.StatusCode, e.Code, e.Message)
}
