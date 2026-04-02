package godidit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Decision holds the full response from the session decision endpoint.
type Decision struct {
	SessionID  string `json:"session_id"`
	Status     string `json:"status"`
	VendorData string `json:"vendor_data,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// RetrieveSession retrieves the decision for a KYC session.
// GET /v3/session/{sessionId}/decision/
func (c *Client) RetrieveSession(ctx context.Context, sessionID string) (*Decision, error) {
	if sessionID == "" {
		return nil, ErrEmptySessionID
	}

	apiRequest := &request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v3/session/%s/decision/", sessionID),
	}

	rawResponse, err := c.execute(ctx, apiRequest)
	if err != nil {
		return nil, err
	}

	result := new(Decision)
	if err := json.Unmarshal(rawResponse, result); err != nil {
		return nil, err
	}

	return result, nil
}
