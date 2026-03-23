package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bogdanticu88/AuthBridge/internal/store"
)

type CookieHandler struct {
	encryption *store.EncryptionManager
}

func NewCookieHandler(e *store.EncryptionManager) *CookieHandler {
	return &CookieHandler{encryption: e}
}

func (h *CookieHandler) Authenticate(ctx context.Context, cred *store.Credential) (*Response, error) {
	// For cookies, the main "token" might be the primary session ID, 
	// but multiple cookies are stored in metadata
	primaryCookie, err := h.encryption.Decrypt(cred.Token)
	if err != nil {
		return nil, err
	}

	var cookies []store.Cookie
	if cred.Metadata != "" {
		if err := json.Unmarshal([]byte(cred.Metadata), &cookies); err != nil {
			return nil, fmt.Errorf("failed to parse cookies from metadata: %w", err)
		}
	}

	return &Response{
		Token:   primaryCookie,
		Type:    "Cookie",
		Cookies: cookies,
	}, nil
}

func (h *CookieHandler) Refresh(ctx context.Context, cred *store.Credential) (*Response, error) {
	return nil, fmt.Errorf("cookie refresh not implemented; manually update via CLI")
}

func (h *CookieHandler) Validate(ctx context.Context, cred *store.Credential) (bool, error) {
	return true, nil
}
