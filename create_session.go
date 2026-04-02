package godidit

import (
	"context"
	"encoding/json"
	"net/http"
)

// CreateSessionRequest holds the payload for creating a new KYC session.
type CreateSessionRequest struct {
	WorkflowID string `json:"workflow_id"`
	VendorData string `json:"vendor_data,omitempty"`
	Callback   string `json:"callback,omitempty"`
}

// CreateSessionResponse is the response from the create session endpoint.
type CreateSessionResponse struct {
	SessionID  string `json:"session_id"`
	URL        string `json:"url"`
	Status     string `json:"status"`
	VendorData string `json:"vendor_data,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

// CreateSession creates a new KYC verification session.
// POST /v3/session/
func (c *Client) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.WorkflowID == "" {
		return nil, ErrEmptyWorkflowID
	}

	apiRequest := &request{
		Method:   http.MethodPost,
		Endpoint: "/v3/session/",
		Params:   req,
	}

	rawResponse, err := c.execute(ctx, apiRequest)
	if err != nil {
		return nil, err
	}

	result := new(CreateSessionResponse)
	if err := json.Unmarshal(rawResponse, result); err != nil {
		return nil, err
	}

	return result, nil
}
