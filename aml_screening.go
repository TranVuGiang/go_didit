package godidit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AMLScreeningRequest holds the inputs for the Didit AML screening API.
type AMLScreeningRequest struct {
	// FullName is the full name of the person to screen (required).
	FullName string `json:"full_name"`
	// DateOfBirth in YYYY-MM-DD format (optional).
	DateOfBirth string `json:"date_of_birth,omitempty"`
	// Nationality is the ISO 3166-1 alpha-2 country code (optional, persons only).
	// Example: "VN", "US"
	Nationality string `json:"nationality,omitempty"`
	// DocumentNumber is the national ID or passport number (optional, persons only).
	DocumentNumber string `json:"document_number,omitempty"`
	// VendorData is an optional internal identifier for tracking.
	VendorData string `json:"vendor_data,omitempty"`
}

// AMLHit represents a single watchlist/sanctions match.
type AMLHit struct {
	ID           string   `json:"id"`
	MatchScore   int      `json:"match_score"`
	RiskScore    float64  `json:"risk_score"`
	ReviewStatus string   `json:"review_status"`
	Datasets     []string `json:"datasets,omitempty"`
}

// AMLData contains the screening result.
type AMLData struct {
	Status      string   `json:"status"`
	EntityType  string   `json:"entity_type"`
	TotalHits   int      `json:"total_hits"`
	Hits        []AMLHit `json:"hits,omitempty"`
	Score       int      `json:"score"`
	Warnings    []string `json:"warnings,omitempty"`
}

// AMLScreeningResponse is the response from POST /v3/aml/.
type AMLScreeningResponse struct {
	RequestID string  `json:"request_id"`
	AML       AMLData `json:"aml"`
	CreatedAt string  `json:"created_at"`
}

// ScreenAML screens an individual against global sanctions, PEP lists, and adverse media.
// Nationality must be ISO 3166-1 alpha-2 (e.g. "VN", "US").
func (c *Client) ScreenAML(ctx context.Context, req *AMLScreeningRequest) (*AMLScreeningResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if strings.TrimSpace(req.FullName) == "" {
		return nil, ErrEmptyFullName
	}

	r := &request{
		Method:   http.MethodPost,
		Endpoint: "/v3/aml/",
		Params:   req,
	}

	respBody, err := c.execute(ctx, r)
	if err != nil {
		return nil, err
	}

	var resp AMLScreeningResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestEncode, err)
	}

	return &resp, nil
}
