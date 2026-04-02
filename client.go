package godidit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const (
	DefaultBaseURL = "https://verification.didit.me"
	UserAgent      = "didit-go-sdk"
)

// Logger is compatible with slog.Logger and any structured logger.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

// HTTPClient allows injecting a custom transport (e.g. for tests).
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Option is a functional option for NewClient.
type Option func(*Client)

// Client holds all runtime configuration.
type Client struct {
	baseURL    string
	apiKey     string
	logger     Logger
	debug      bool
	httpClient HTTPClient
}

// NewClient constructs a Client. Returns an error if apiKey is empty.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, ErrEmptyAPIKey
	}

	c := &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		logger:     slog.Default(),
		debug:      false,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.baseURL = strings.TrimRight(c.baseURL, "/")

	return c, nil
}

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.debug = debug
	}
}

func WithHTTPClient(httpClient HTTPClient) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func (c *Client) logDebug(msg string, attrs ...any) {
	if c.debug {
		c.logger.Debug(msg, attrs...)
	}
}

func (c *Client) buildRequest(req *request) error {
	if req == nil {
		return ErrNilRequest
	}

	fullURL := c.baseURL + req.Endpoint

	headers := http.Header{}
	if req.Header != nil {
		headers = req.Header.Clone()
	}

	headers.Set("User-Agent", UserAgent)
	headers.Set("x-api-key", c.apiKey)

	if req.Params != nil {
		bodyBytes, err := json.Marshal(req.Params)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrRequestEncode, err)
		}

		c.logDebug("http request", "url", fullURL, "body", string(bodyBytes))

		headers.Set("Content-Type", "application/json")
		req.Body = bytes.NewReader(bodyBytes)
	} else {
		c.logDebug("http request", "url", fullURL)
	}

	req.FullURL = fullURL
	req.Header = headers

	return nil
}

func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}
	if err := json.Unmarshal(body, apiErr); err != nil {
		return fmt.Errorf("%w: status=%d body=%s", ErrUnexpectedStatus, statusCode, string(body))
	}

	return apiErr
}

func (c *Client) execute(ctx context.Context, req *request) ([]byte, error) {
	if err := c.buildRequest(req); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.FullURL, req.Body)
	if err != nil {
		return nil, err
	}

	httpReq.Header = req.Header

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrHTTPFailure, err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.logDebug("http response", "status", resp.StatusCode, "body", string(responseBody))

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, c.handleErrorResponse(resp.StatusCode, responseBody)
	}

	return responseBody, nil
}
