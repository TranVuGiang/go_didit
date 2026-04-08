package godidit

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	// ImageKyc is the S3 URL of the user's selfie image.
	ImageKyc         *string `db:"image_kyc"`
	// FrontIdImage is the S3 URL of the front side of the identity document.
	FrontIdImage     *string `db:"front_id_image"`
	// BackIdImage is the S3 URL of the back side of the identity document.
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
// Images are downloaded from their S3 URLs before upload.
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

			frontBytes, err := downloadImage(ctx, *kyc.FrontIdImage)
			if err != nil {
				fail(fmt.Errorf("%w: front_id_image: %w", ErrImageDownloadFailed, err))
				return
			}

			req := &IDVerificationRequest{
				FrontImage: frontBytes,
				VendorData: vendorData,
			}

			if kyc.BackIdImage != nil {
				backBytes, err := downloadImage(ctx, *kyc.BackIdImage)
				if err != nil {
					fail(fmt.Errorf("%w: back_id_image: %w", ErrImageDownloadFailed, err))
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
				IssuingState:         *kyc.Nationality,
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
				req.Gender = *kyc.Gender
			}
			if kyc.Type != nil {
				req.DocumentType = *kyc.Type
			}
			if kyc.Address != nil {
				req.Address = *kyc.Address
			}

			if kyc.ImageKyc != nil {
				selfieBytes, err := downloadImage(ctx, *kyc.ImageKyc)
				if err != nil {
					fail(fmt.Errorf("%w: image_kyc: %w", ErrImageDownloadFailed, err))
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

// downloadImage fetches raw image bytes from a URL (e.g. an S3 presigned URL).
func downloadImage(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("image download returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
