package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bogdanticu88/AuthBridge/internal/store"
)

type OAuth2Metadata struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	TokenURL     string `json:"token_url"`
}

type OAuth2Handler struct {
	store      store.Store
	encryption *store.EncryptionManager
	mu         sync.Mutex
}

func NewOAuth2Handler(s store.Store, e *store.EncryptionManager) *OAuth2Handler {
	return &OAuth2Handler{store: s, encryption: e}
}

func (h *OAuth2Handler) Authenticate(ctx context.Context, cred *store.Credential) (*Response, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// First, check if the current token is valid
	jwtHandler := NewJWTHandler(h.encryption)
	resp, err := jwtHandler.Authenticate(ctx, cred)

	// If it's valid and not expiring soon (e.g., > 5 mins), return it
	if err == nil && resp.ExpiresAt != nil && resp.ExpiresAt.After(time.Now().Add(5*time.Minute)) {
		return resp, nil
	}

	// Token expired or expiring soon, try to refresh
	return h.refreshLocked(ctx, cred)
}

func (h *OAuth2Handler) Refresh(ctx context.Context, cred *store.Credential) (*Response, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.refreshLocked(ctx, cred)
}

func (h *OAuth2Handler) refreshLocked(ctx context.Context, cred *store.Credential) (*Response, error) {
	var meta OAuth2Metadata
	if err := json.Unmarshal([]byte(cred.Metadata), &meta); err != nil {
		return nil, fmt.Errorf("invalid oauth2 metadata: %w", err)
	}

	if meta.RefreshToken == "" || meta.TokenURL == "" {
		return nil, errors.New("missing refresh_token or token_url in metadata")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", meta.RefreshToken)
	if meta.ClientID != "" {
		data.Set("client_id", meta.ClientID)
	}
	if meta.ClientSecret != "" {
		data.Set("client_secret", meta.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", meta.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status: %d", httpResp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Update credential with new access token
	encryptedToken, err := h.encryption.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	cred.Token = encryptedToken
	if tokenResp.RefreshToken != "" {
		meta.RefreshToken = tokenResp.RefreshToken
		newMeta, err := json.Marshal(meta)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal new metadata: %w", err)
		}
		cred.Metadata = string(newMeta)
	}

	if err := h.store.UpdateCredential(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to update store after refresh: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return &Response{
		Token:     tokenResp.AccessToken,
		Type:      tokenResp.TokenType,
		ExpiresAt: &expiresAt,
	}, nil
}

func (h *OAuth2Handler) Validate(ctx context.Context, cred *store.Credential) (bool, error) {
	jwtHandler := NewJWTHandler(h.encryption)
	return jwtHandler.Validate(ctx, cred)
}
