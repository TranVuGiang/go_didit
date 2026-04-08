package godidit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DatabaseValidationRequest holds the inputs for the Didit government database validation API.
type DatabaseValidationRequest struct {
	// IssuingState is the ISO 3166-1 alpha-3 country code (required).
	// Example: "VNM", "USA", "BRA"
	IssuingState string
	// ValidationType is "one_by_one" or "two_by_two" (required).
	ValidationType string
	// IdentificationNumber is the national ID, passport number, or equivalent (required).
	IdentificationNumber string
	// DocumentType: P (Passport), DL (Driver License), ID (National ID), RP (Residence Permit).
	DocumentType string
	FirstName    string
	LastName     string
	// DateOfBirth in YYYY-MM-DD format.
	DateOfBirth string
	// Nationality is ISO 3166-1 alpha-3 (e.g. "VNM").
	Nationality string
	Address     string
	// Gender: M, F, or X.
	Gender string
	// Selfie is raw image bytes for biometric validation (optional, required for some countries).
	Selfie []byte
	// VendorData is an optional internal identifier for tracking.
	VendorData string
}

// DatabaseValidationField contains per-field match results.
type DatabaseValidationField struct {
	Match bool   `json:"match"`
	Value string `json:"value,omitempty"`
}

// DatabaseValidationEntry holds one validation result set from a government source.
type DatabaseValidationEntry struct {
	Validation map[string]DatabaseValidationField `json:"validation,omitempty"`
	SourceData map[string]any                     `json:"source_data,omitempty"`
}

// DatabaseValidationData contains the full database validation result.
type DatabaseValidationData struct {
	Status         string                    `json:"status"`
	IssuingState   string                    `json:"issuing_state"`
	ValidationType string                    `json:"validation_type"`
	MatchType      string                    `json:"match_type"` // full_match | partial_match | no_match
	Validations    []DatabaseValidationEntry `json:"validations,omitempty"`
	Warnings       []string                  `json:"warnings,omitempty"`
}

// DatabaseValidationResponse is the response from POST /v3/database-validation/.
type DatabaseValidationResponse struct {
	RequestID          string                 `json:"request_id"`
	DatabaseValidation DatabaseValidationData `json:"database_validation"`
	CreatedAt          string                 `json:"created_at"`
}

// ValidateDatabase matches identity data against government databases.
// IssuingState and Nationality must be ISO 3166-1 alpha-3 (e.g. "VNM").
// Note: supported countries are limited — see Didit docs for the full list.
func (c *Client) ValidateDatabase(ctx context.Context, req *DatabaseValidationRequest) (*DatabaseValidationResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if strings.TrimSpace(req.IdentificationNumber) == "" {
		return nil, ErrEmptyNationalID
	}
	if strings.TrimSpace(req.IssuingState) == "" {
		return nil, fmt.Errorf("%w: issuing_state is required", ErrInvalidParams)
	}

	validationType := req.ValidationType
	if validationType == "" {
		validationType = "one_by_one"
	}

	fields := map[string]string{
		"issuing_state":        req.IssuingState,
		"validation_type":      validationType,
		"identification_number": req.IdentificationNumber,
	}
	if req.DocumentType != "" {
		fields["document_type"] = req.DocumentType
	}
	if req.FirstName != "" {
		fields["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		fields["last_name"] = req.LastName
	}
	if req.DateOfBirth != "" {
		fields["date_of_birth"] = req.DateOfBirth
	}
	if req.Nationality != "" {
		fields["nationality"] = req.Nationality
	}
	if req.Address != "" {
		fields["address"] = req.Address
	}
	if req.Gender != "" {
		fields["gender"] = req.Gender
	}
	if req.VendorData != "" {
		fields["vendor_data"] = req.VendorData
	}

	files := map[string][]byte{}
	if len(req.Selfie) > 0 {
		files["selfie"] = req.Selfie
	}

	body, contentType, err := buildMultipartBody(fields, files)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestEncode, err)
	}

	r := &request{
		Method:   http.MethodPost,
		Endpoint: "/v3/database-validation/",
		Header:   http.Header{"Content-Type": []string{contentType}},
		Body:     body,
	}

	respBody, err := c.execute(ctx, r)
	if err != nil {
		return nil, err
	}

	var resp DatabaseValidationResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestEncode, err)
	}

	return &resp, nil
}
