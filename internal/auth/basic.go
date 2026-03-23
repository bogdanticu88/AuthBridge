package auth

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/twister69/authbridge/internal/store"
)

type BasicAuthHandler struct {
	encryption *store.EncryptionManager
}

func NewBasicAuthHandler(e *store.EncryptionManager) *BasicAuthHandler {
	return &BasicAuthHandler{encryption: e}
}

func (h *BasicAuthHandler) Authenticate(ctx context.Context, cred *store.Credential) (*Response, error) {
	creds, err := h.encryption.Decrypt(cred.Token)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(creds))
	return &Response{
		Token: encoded,
		Type:  "Basic",
	}, nil
}

func (h *BasicAuthHandler) Refresh(ctx context.Context, cred *store.Credential) (*Response, error) {
	return nil, fmt.Errorf("basic auth does not support refresh")
}

func (h *BasicAuthHandler) Validate(ctx context.Context, cred *store.Credential) (bool, error) {
	return true, nil // Basic auth is always "valid" if stored
}
