package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/bogdanticu88/AuthBridge/internal/store"
)

type JWTHandler struct {
	encryption *store.EncryptionManager
}

func NewJWTHandler(e *store.EncryptionManager) *JWTHandler {
	return &JWTHandler{encryption: e}
}

func (h *JWTHandler) Authenticate(ctx context.Context, cred *store.Credential) (*Response, error) {
	tokenStr, err := h.encryption.Decrypt(cred.Token)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		// Not a valid JWT format, but we can still return the raw token
		return &Response{
			Token: tokenStr,
			Type:  "Bearer",
		}, nil
	}

	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		// Non-standard exp claim, or missing
		return &Response{
			Token: tokenStr,
			Type:  "Bearer",
		}, nil
	}

	if exp.Before(time.Now()) {
		return nil, errors.New("jwt expired")
	}

	expiresAt := exp.Time
	return &Response{
		Token:     tokenStr,
		Type:      "Bearer",
		ExpiresAt: &expiresAt,
	}, nil
}

func (h *JWTHandler) Refresh(ctx context.Context, cred *store.Credential) (*Response, error) {
	// JWT refresh typically requires a refresh token in metadata
	// For now, return an error as we'll handle this in OAuth2
	return nil, errors.New("jwt refresh not implemented directly; use oauth2 handler")
}

func (h *JWTHandler) Validate(ctx context.Context, cred *store.Credential) (bool, error) {
	resp, err := h.Authenticate(ctx, cred)
	if err != nil {
		return false, err
	}
	if resp.ExpiresAt != nil && resp.ExpiresAt.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}
