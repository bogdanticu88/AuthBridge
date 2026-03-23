package authbridge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TokenResponse represents the response from AuthBridge token endpoint
type TokenResponse struct {
	Token     string     `json:"token"`
	Type      string     `json:"type"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Client is a client for the AuthBridge daemon
type Client struct {
	BaseURL string
	HTTPClient *http.Client
}

// NewClient creates a new AuthBridge client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:9999"
	}
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetToken fetches a token for a given credential name
func (c *Client) GetToken(name string) (string, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/v1/token/%s", c.BaseURL, name))
	if err != nil {
		return "", fmt.Errorf("authbridge request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authbridge returned status: %d", resp.StatusCode)
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("failed to decode authbridge response: %w", err)
	}

	return tr.Token, nil
}

// GetAuthHeader returns a formatted Authorization header value
func (c *Client) GetAuthHeader(name string) (string, error) {
	token, err := c.GetToken(name)
	if err != nil {
		return "", err
	}
	// For now we assume Bearer, we could enhance this based on TokenResponse.Type
	return fmt.Sprintf("Bearer %s", token), nil
}
