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

// IDVerificationLocation is a lat/lng coordinate.
type IDVerificationLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// IDVerificationParsedAddress is the geocoded address extracted from the document.
type IDVerificationParsedAddress struct {
	City             string                 `json:"city,omitempty"`
	Label            string                 `json:"label,omitempty"`
	Region           string                 `json:"region,omitempty"`
	Country          string                 `json:"country,omitempty"`
	Category         string                 `json:"category,omitempty"`
	Street1          string                 `json:"street_1,omitempty"`
	Street2          string                 `json:"street_2,omitempty"`
	IsVerified       bool                   `json:"is_verified"`
	PostalCode       string                 `json:"postal_code,omitempty"`
	AddressType      string                 `json:"address_type,omitempty"`
	FormattedAddress string                 `json:"formatted_address,omitempty"`
	DocumentLocation IDVerificationLocation `json:"document_location,omitempty"`
}

// IDVerificationExtra holds additional document-specific fields.
type IDVerificationExtra struct {
	DLCategories []any   `json:"dl_categories,omitempty"`
	BloodGroup   *string `json:"blood_group"`
}

// IDVerificationMRZ contains Machine Readable Zone data extracted from the document.
type IDVerificationMRZ struct {
	Surname            string `json:"surname,omitempty"`
	Name               string `json:"name,omitempty"`
	Country            string `json:"country,omitempty"`
	Nationality        string `json:"nationality,omitempty"`
	BirthDate          string `json:"birth_date,omitempty"`
	ExpiryDate         string `json:"expiry_date,omitempty"`
	Sex                string `json:"sex,omitempty"`
	DocumentType       string `json:"document_type,omitempty"`
	DocumentNumber     string `json:"document_number,omitempty"`
	OptionalData       string `json:"optional_data,omitempty"`
	OptionalData2      string `json:"optional_data_2,omitempty"`
	BirthDateHash      string `json:"birth_date_hash,omitempty"`
	ExpiryDateHash     string `json:"expiry_date_hash,omitempty"`
	DocumentNumberHash string `json:"document_number_hash,omitempty"`
	FinalHash          string `json:"final_hash,omitempty"`
	PersonalNumber     string `json:"personal_number,omitempty"`
	MRZType            string `json:"mrz_type,omitempty"`
	MRZString          string `json:"mrz_string,omitempty"`
	MRZKey             string `json:"mrz_key,omitempty"`
	Warnings           []any  `json:"warnings,omitempty"`
	Errors             []any  `json:"errors,omitempty"`
}

// IDVerificationBarcodeData holds the structured data parsed from a barcode.
type IDVerificationBarcodeData struct {
	First           string  `json:"first,omitempty"`
	Last            string  `json:"last,omitempty"`
	Middle          *string `json:"middle"`
	City            string  `json:"city,omitempty"`
	State           string  `json:"state,omitempty"`
	Address         string  `json:"address,omitempty"`
	IssueIdentifier string  `json:"issue_identifier,omitempty"`
	DocumentNumber  string  `json:"document_number,omitempty"`
	ExpirationDate  string  `json:"expiration_date,omitempty"`
	DateOfBirth     string  `json:"date_of_birth,omitempty"`
	PostalCode      string  `json:"postal_code,omitempty"`
	Sex             string  `json:"sex,omitempty"`
	Height          string  `json:"height,omitempty"`
	Weight          string  `json:"weight,omitempty"`
	Hair            string  `json:"hair,omitempty"`
	Eyes            string  `json:"eyes,omitempty"`
	Issued          string  `json:"issued,omitempty"`
	Units           string  `json:"units,omitempty"`
}

// IDVerificationBarcode holds a single barcode scan result from the document.
type IDVerificationBarcode struct {
	Type    string                    `json:"type,omitempty"`
	Data    IDVerificationBarcodeData `json:"data,omitempty"`
	DataRaw string                    `json:"data_raw,omitempty"`
	Side    string                    `json:"side,omitempty"`
}

// IDVerificationWarning describes a risk or issue found during verification.
type IDVerificationWarning struct {
	Risk             string  `json:"risk,omitempty"`
	AdditionalData   *string `json:"additional_data"`
	LogType          string  `json:"log_type,omitempty"`
	ShortDescription string  `json:"short_description,omitempty"`
	LongDescription  string  `json:"long_description,omitempty"`
}

// IDVerificationData contains the extracted identity data from the submitted document.
type IDVerificationData struct {
	Status             string                      `json:"status"`
	IssuingState       string                      `json:"issuing_state,omitempty"`
	IssuingStateName   string                      `json:"issuing_state_name,omitempty"`
	Region             *string                     `json:"region"`
	DocumentType       string                      `json:"document_type,omitempty"`
	DocumentNumber     string                      `json:"document_number,omitempty"`
	PersonalNumber     string                      `json:"personal_number,omitempty"`
	DateOfBirth        string                      `json:"date_of_birth,omitempty"`
	Age                int                         `json:"age,omitempty"`
	ExpirationDate     string                      `json:"expiration_date,omitempty"`
	DateOfIssue        string                      `json:"date_of_issue,omitempty"`
	FirstName          string                      `json:"first_name,omitempty"`
	LastName           string                      `json:"last_name,omitempty"`
	FullName           string                      `json:"full_name,omitempty"`
	Gender             string                      `json:"gender,omitempty"`
	Address            string                      `json:"address,omitempty"`
	FormattedAddress   string                      `json:"formatted_address,omitempty"`
	PlaceOfBirth       string                      `json:"place_of_birth,omitempty"`
	MaritalStatus      *string                     `json:"marital_status"`
	Nationality        string                      `json:"nationality,omitempty"`
	ExtraFields        IDVerificationExtra         `json:"extra_fields,omitempty"`
	ParsedAddress      IDVerificationParsedAddress `json:"parsed_address,omitempty"`
	PortraitImage      string                      `json:"portrait_image,omitempty"`
	FrontDocumentImage string                      `json:"front_document_image,omitempty"`
	BackDocumentImage  string                      `json:"back_document_image,omitempty"`
	MRZ                IDVerificationMRZ           `json:"mrz,omitempty"`
	Barcodes           []IDVerificationBarcode     `json:"barcodes,omitempty"`
	Matches            []any                       `json:"matches,omitempty"`
	Warnings           []IDVerificationWarning     `json:"warnings,omitempty"`
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
