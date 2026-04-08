package godidit

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// KycInfo holds all KYC data for a user, typically sourced from the database.
type KycInfo struct {
	ID               int     `db:"id"`
	UserID           *int    `db:"user_id"`
	FirstName        *string `db:"first_name"`
	LastName         *string `db:"last_name"`
	Email            *string `db:"email"`
	PhoneNumber      *string `db:"phone_number"`
	PhoneCountryCode *string `db:"phone_country_code"`
	DateOfBirth      *string `db:"date_of_birth"`
	Gender           *string `db:"gender"`
	// Nationality must be ISO 3166-1 alpha-2 for AML (e.g. "VN") and
	// ISO 3166-1 alpha-3 for DatabaseValidation / IDVerification (e.g. "VNM").
	// If calling both, ensure the value matches the format expected by each service.
	Nationality      *string `db:"nationality"`
	Type             *string `db:"type"`
	NationalID       *string `db:"national_id"`
	Address          *string `db:"address"`
	AddressLine1     *string `db:"address_line1"`
	AddressLine2     *string `db:"address_line2"`
	City             *string `db:"city"`
	State            *string `db:"state"`
	ZipCode          *string `db:"zip_code"`
	// ImageKyc is the base64-encoded selfie image.
	// Accepts raw base64 or data URI format ("data:image/jpeg;base64,...").
	ImageKyc         *string `db:"image_kyc"`
	// FrontIdImage is the base64-encoded front side of the identity document.
	// Accepts raw base64 or data URI format.
	FrontIdImage     *string `db:"front_id_image"`
	// BackIdImage is the base64-encoded back side of the identity document.
	// Accepts raw base64 or data URI format.
	BackIdImage      *string `db:"back_id_image"`
}

// KycVerifyResult aggregates results from all verification services called by VerifyKyc.
// A nil field means that service was not called (insufficient data) or was skipped.
type KycVerifyResult struct {
	IDVerification     *IDVerificationResponse
	AMLScreening       *AMLScreeningResponse
	DatabaseValidation *DatabaseValidationResponse
}

// VerifyKyc submits KYC data to all applicable Didit verification services in parallel.
//
// Services are selected automatically based on available data:
//   - IDVerification: called when FrontIdImage is set
//   - AMLScreening:   called when FirstName and LastName are set
//   - DatabaseValidation: called when NationalID and Nationality are set
//
// Image fields (FrontIdImage, BackIdImage, ImageKyc) must be base64-encoded strings.
// On any error, the context is cancelled and the first error is returned (fail-fast).
func (c *Client) VerifyKyc(ctx context.Context, kyc *KycInfo) (*KycVerifyResult, error) {
	if kyc == nil {
		return nil, ErrNilRequest
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
		mu       sync.Mutex
		result   KycVerifyResult
	)

	fail := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	vendorData := ""
	if kyc.UserID != nil {
		vendorData = strconv.Itoa(*kyc.UserID)
	}

	// --- ID Verification ---
	if kyc.FrontIdImage != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			frontBytes, err := DecodeBase64Image(*kyc.FrontIdImage)
			if err != nil {
				fail(fmt.Errorf("%w: front_id_image: %w", ErrImageDecodeFailed, err))
				return
			}

			req := &IDVerificationRequest{
				FrontImage: frontBytes,
				VendorData: vendorData,
			}

			if kyc.BackIdImage != nil {
				backBytes, err := DecodeBase64Image(*kyc.BackIdImage)
				if err != nil {
					fail(fmt.Errorf("%w: back_id_image: %w", ErrImageDecodeFailed, err))
					return
				}
				req.BackImage = backBytes
			}

			resp, err := c.SubmitIDVerification(ctx, req)
			if err != nil {
				fail(err)
				return
			}

			mu.Lock()
			result.IDVerification = resp
			mu.Unlock()
		}()
	}

	// --- AML Screening ---
	if kyc.FirstName != nil && kyc.LastName != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := &AMLScreeningRequest{
				FullName:   *kyc.FirstName + " " + *kyc.LastName,
				VendorData: vendorData,
			}
			if kyc.DateOfBirth != nil {
				req.DateOfBirth = *kyc.DateOfBirth
			}
			if kyc.Nationality != nil {
				req.Nationality = *kyc.Nationality
			}
			if kyc.NationalID != nil {
				req.DocumentNumber = *kyc.NationalID
			}

			resp, err := c.ScreenAML(ctx, req)
			if err != nil {
				fail(err)
				return
			}

			mu.Lock()
			result.AMLScreening = resp
			mu.Unlock()
		}()
	}

	// --- Database Validation ---
	if kyc.NationalID != nil && kyc.Nationality != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := &DatabaseValidationRequest{
				IssuingState:         ToAlpha3(*kyc.Nationality), // alpha-2 "VN" → alpha-3 "VNM"
				ValidationType:       "one_by_one",
				IdentificationNumber: *kyc.NationalID,
				VendorData:           vendorData,
			}
			if kyc.FirstName != nil {
				req.FirstName = *kyc.FirstName
			}
			if kyc.LastName != nil {
				req.LastName = *kyc.LastName
			}
			if kyc.DateOfBirth != nil {
				req.DateOfBirth = *kyc.DateOfBirth
			}
			if kyc.Gender != nil {
				req.Gender = ToGenderCode(*kyc.Gender) // "male" → "M"
			}
			if kyc.Type != nil {
				req.DocumentType = ToDocTypeCode(*kyc.Type) // "ID_CARD" → "ID"
			}
			if kyc.Address != nil {
				req.Address = *kyc.Address
			}

			if kyc.ImageKyc != nil {
				selfieBytes, err := DecodeBase64Image(*kyc.ImageKyc)
				if err != nil {
					fail(fmt.Errorf("%w: image_kyc: %w", ErrImageDecodeFailed, err))
					return
				}
				req.Selfie = selfieBytes
			}

			resp, err := c.ValidateDatabase(ctx, req)
			if err != nil {
				fail(err)
				return
			}

			mu.Lock()
			result.DatabaseValidation = resp
			mu.Unlock()
		}()
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	if result.IDVerification == nil && result.AMLScreening == nil && result.DatabaseValidation == nil {
		return nil, ErrNoVerificationPerformed
	}

	return &result, nil
}

// DecodeBase64Image decodes a base64-encoded image string to raw bytes.
// Accepts raw base64 or data URI format ("data:image/jpeg;base64,...").
func DecodeBase64Image(b64str string) ([]byte, error) {
	if idx := strings.Index(b64str, ","); idx != -1 {
		b64str = b64str[idx+1:]
	}
	b64str = strings.TrimSpace(b64str)
	data, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		// Fallback: try without padding (some encoders omit '=')
		data, err = base64.RawStdEncoding.DecodeString(b64str)
	}
	return data, err
}
