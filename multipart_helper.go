package godidit

import (
	"bytes"
	"fmt"
	"mime/multipart"
)

// buildMultipartBody constructs a multipart/form-data body from text fields and binary files.
// fields: form field name → string value
// files:  form field name → raw bytes (e.g. image data)
// Returns the body buffer, the Content-Type header value (including boundary), and any error.
func buildMultipartBody(fields map[string]string, files map[string][]byte) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return nil, "", fmt.Errorf("multipart: write field %q: %w", k, err)
		}
	}

	for fieldName, data := range files {
		part, err := writer.CreateFormFile(fieldName, fieldName+".jpg")
		if err != nil {
			return nil, "", fmt.Errorf("multipart: create file field %q: %w", fieldName, err)
		}
		if _, err := part.Write(data); err != nil {
			return nil, "", fmt.Errorf("multipart: write file %q: %w", fieldName, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("multipart: close writer: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}
