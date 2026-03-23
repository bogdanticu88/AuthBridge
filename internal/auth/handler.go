package auth

import (
	"context"
	"time"

	"github.com/bogdanticu88/AuthBridge/internal/store"
)

// Response represents the standard response from any auth handler
type Response struct {
	Token     string            `json:"token"`
	Type      string            `json:"type"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Cookies   []store.Cookie    `json:"cookies,omitempty"`
}

// Handler defines the interface for different authentication types
type Handler interface {
	// Authenticate returns the current valid token/session, refreshing if necessary
	Authenticate(ctx context.Context, cred *store.Credential) (*Response, error)
	// Refresh manually triggers a credential refresh
	Refresh(ctx context.Context, cred *store.Credential) (*Response, error)
	// Validate checks if the current credential is still valid
	Validate(ctx context.Context, cred *store.Credential) (bool, error)
}
