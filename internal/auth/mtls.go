package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twister69/authbridge/internal/store"
)

type MTLSMetadata struct {
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
	CertData string `json:"cert_data"` // Base64 encoded PEM
	KeyData  string `json:"key_data"`  // Base64 encoded PEM
}

type MTLSHandler struct {
	encryption *store.EncryptionManager
}

func NewMTLSHandler(e *store.EncryptionManager) *MTLSHandler {
	return &MTLSHandler{encryption: e}
}

func (h *MTLSHandler) Authenticate(ctx context.Context, cred *store.Credential) (*Response, error) {
	var meta MTLSMetadata
	if err := json.Unmarshal([]byte(cred.Metadata), &meta); err != nil {
		return nil, fmt.Errorf("invalid mtls metadata: %w", err)
	}

	// For mTLS, the "token" is the certificate chain
	// We'll return the paths or data so tools can use them
	// In some tools, we'd need to provide this via a proxy
	
	return &Response{
		Token:   "MTLS_CERT_READY",
		Type:    "mTLS",
		Headers: map[string]string{
			"X-AuthBridge-Cert": meta.CertPath,
			"X-AuthBridge-Key":  meta.KeyPath,
		},
	}, nil
}

func (h *MTLSHandler) Refresh(ctx context.Context, cred *store.Credential) (*Response, error) {
	return h.Authenticate(ctx, cred)
}

func (h *MTLSHandler) Validate(ctx context.Context, cred *store.Credential) (bool, error) {
	return true, nil
}
