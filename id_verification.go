package godidit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// IDVerificationRequest holds the inputs for the Didit ID document verification API.
type IDVerificationRequest struct {
	// FrontImage is the raw bytes of the front side of the identity document (required).
	FrontImage []byte
	// BackImage is the raw bytes of the back side (optional).
	BackImage []byte
	// VendorData is an optional internal identifier (e.g. user UUID) for tracking.
	VendorData string
}

// IDVerificationAddress contains parsed address fields returned by Didit.
type IDVerificationAddress struct {
	RawAddress string `json:"raw_address,omitempty"`
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// IDVerificationMRZ contains Machine Readable Zone data extracted from the document.
type IDVerificationMRZ struct {
	Line1    string `json:"line1,omitempty"`
	Line2    string `json:"line2,omitempty"`
	Line3    string `json:"line3,omitempty"`
	Valid    bool   `json:"valid"`
	Checksum string `json:"checksum,omitempty"`
}

// IDVerificationData contains the extracted identity data from the submitted document.
type IDVerificationData struct {
	Status         string                `json:"status"`
	DocumentType   string                `json:"document_type,omitempty"`
	DocumentNumber string                `json:"document_number,omitempty"`
	IssuingState   string                `json:"issuing_state,omitempty"`
	FirstName      string                `json:"first_name,omitempty"`
	LastName       string                `json:"last_name,omitempty"`
	FullName       string                `json:"full_name,omitempty"`
	DateOfBirth    string                `json:"date_of_birth,omitempty"`
	Age            int                   `json:"age,omitempty"`
	Gender         string                `json:"gender,omitempty"`
	Nationality    string                `json:"nationality,omitempty"`
	IssueDate      string                `json:"issue_date,omitempty"`
	ExpirationDate string                `json:"expiration_date,omitempty"`
	Address        IDVerificationAddress `json:"address,omitempty"`
	Portrait       string                `json:"portrait,omitempty"` // base64 portrait image
	MRZ            IDVerificationMRZ     `json:"mrz,omitempty"`
	Warnings       []string              `json:"warnings,omitempty"`
}

// IDVerificationResponse is the response from POST /v3/id-verification/.
type IDVerificationResponse struct {
	RequestID      string             `json:"request_id"`
	IDVerification IDVerificationData `json:"id_verification"`
	CreatedAt      string             `json:"created_at"`
}

// SubmitIDVerification sends document images to Didit for identity verification.
// The API extracts personal data (name, DOB, nationality, etc.) directly from the images.
func (c *Client) SubmitIDVerification(ctx context.Context, req *IDVerificationRequest) (*IDVerificationResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if len(req.FrontImage) == 0 {
		return nil, ErrEmptyFrontImage
	}

	fields := map[string]string{}
	if req.VendorData != "" {
		fields["vendor_data"] = req.VendorData
	}

	files := map[string][]byte{
		"front_image": req.FrontImage,
	}
	if len(req.BackImage) > 0 {
		files["back_image"] = req.BackImage
	}

	body, contentType, err := buildMultipartBody(fields, files)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestEncode, err)
	}

	r := &request{
		Method:   http.MethodPost,
		Endpoint: "/v3/id-verification/",
		Header:   http.Header{"Content-Type": []string{contentType}},
		Body:     body,
	}

	respBody, err := c.execute(ctx, r)
	if err != nil {
		return nil, err
	}

	var resp IDVerificationResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestEncode, err)
	}

	return &resp, nil
}
