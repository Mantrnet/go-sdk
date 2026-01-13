// Package mantr provides a client for the Mantr API
package mantr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

//errors
var (
	// ErrAuthentication indicates invalid API key
	ErrAuthentication = errors.New("mantr: authentication failed - invalid API key")

	// ErrInsufficientCredits indicates insufficient credits
	ErrInsufficientCredits = errors.New("mantr: insufficient credits")

	// ErrRateLimit indicates rate limit exceeded
	ErrRateLimit = errors.New("mantr: rate limit exceeded")
)

// Client is the Mantr API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Mantr API client
func NewClient(apiKey string, options ...Option) (*Client, error) {
	if len(apiKey) < 4 || apiKey[:4] != "vak_" {
		return nil, fmt.Errorf("invalid API key format, must start with 'vak_'")
	}

	client := &Client{
		apiKey:  apiKey,
		baseURL: "https://api.mantr.net",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range options {
		opt(client)
	}

	return client, nil
}

// Walk traverses the semantic graph
func (c *Client) Walk(req *WalkRequest) (*WalkResponse, error) {
	if len(req.Phonemes) == 0 {
		return nil, fmt.Errorf("phonemes cannot be empty")
	}

	// Set defaults
	if req.Depth == 0 {
		req.Depth = 3
	}
	if req.Limit == 0 {
		req.Limit = 100
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/v1/walk", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("User-Agent", "mantr-go/1.0.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, ErrAuthentication
	} else if resp.StatusCode == 402 {
		return nil, ErrInsufficientCredits
	} else if resp.StatusCode == 429 {
		return nil, ErrRateLimit
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var walkResp WalkResponse
	if err := json.NewDecoder(resp.Body).Decode(&walkResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &walkResp, nil
}

// WalkRequest represents a walk API request
type WalkRequest struct {
	Phonemes []string `json:"phonemes"`
	Pod      string   `json:"pod,omitempty"`
	Depth    int      `json:"depth,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// PathResult represents a single path in the graph
type PathResult struct {
	Nodes []string `json:"nodes"`
	Score float64  `json:"score"`
	Depth int      `json:"depth"`
}

// WalkResponse represents a walk API response
type WalkResponse struct {
	Paths       []PathResult `json:"paths"`
	LatencyUS   int          `json:"latency_us"`
	CreditsUsed int          `json:"credits_used"`
}

// Option is a functional option for Client
type Option func(*Client)

// WithBaseURL sets a custom base URL
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}
