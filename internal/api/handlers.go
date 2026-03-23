package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/bogdanticu88/AuthBridge/internal/auth"
	"github.com/bogdanticu88/AuthBridge/internal/store"
)

type APIHandler struct {
	store      store.Store
	encryption *store.EncryptionManager
	handlers   map[string]auth.Handler // Initialized once, never modified - safe for concurrent reads
}

func NewAPIHandler(s store.Store, e *store.EncryptionManager) *APIHandler {
	return &APIHandler{
		store:      s,
		encryption: e,
		handlers: map[string]auth.Handler{
			"jwt":      auth.NewJWTHandler(e),
			"oauth2":   auth.NewOAuth2Handler(s, e),
			"basic":    auth.NewBasicAuthHandler(e),
			"cookie":   auth.NewCookieHandler(e),
			"kerberos": auth.NewKerberosHandler(e),
			"mtls":     auth.NewMTLSHandler(e),
		},
	}
}

func (h *APIHandler) GetToken(c *gin.Context) {
	name := c.Param("name")
	
	cred, err := h.store.GetCredential(c.Request.Context(), name)
	if err != nil {
		h.logAudit(c.ClientIP(), c.GetHeader("User-Agent"), "token_fetch", name, "failed", err.Error())
		if err == store.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	handler := h.getHandler(cred.Type)
	resp, err := handler.Authenticate(c.Request.Context(), cred)
	if err != nil {
		h.logAudit(c.ClientIP(), c.GetHeader("User-Agent"), "token_fetch", name, "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update last used and log audit in background
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	go func() {
		// Use a context with timeout to prevent goroutine leaks
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.store.UpdateLastUsed(ctx, name); err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed to update last_used")
		}
		h.logAudit(ip, ua, "token_fetch", name, "success", "")
	}()

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) ListCredentials(c *gin.Context) {
	creds, err := h.store.ListCredentials(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credentials"})
		return
	}

	// Don't return the encrypted tokens in the list for security
	type CredInfo struct {
		Name       string `json:"name"`
		Type       string `json:"type"`
		UsageCount int    `json:"usage_count"`
	}
	var results []CredInfo
	for _, cr := range creds {
		results = append(results, CredInfo{
			Name:       cr.Name,
			Type:       cr.Type,
			UsageCount: cr.UsageCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"credentials": results})
}

type AddCredentialRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Token    string `json:"token" binding:"required"`
	Metadata string `json:"metadata"`
}

func (h *APIHandler) AddCredential(c *gin.Context) {
	var req AddCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedToken, err := h.encryption.Encrypt(req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
		return
	}

	cred := &store.Credential{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Type:      req.Type,
		Token:     encryptedToken,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
	}

	if err := h.store.AddCredential(c.Request.Context(), cred); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store credential"})
		return
	}

	h.logAudit(c.ClientIP(), c.GetHeader("User-Agent"), "credential_add", req.Name, "success", "")
	c.JSON(http.StatusCreated, gin.H{"status": "created", "name": req.Name})
}

func (h *APIHandler) DeleteCredential(c *gin.Context) {
	name := c.Param("name")
	if err := h.store.DeleteCredential(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete credential"})
		return
	}

	h.logAudit(c.ClientIP(), c.GetHeader("User-Agent"), "credential_delete", name, "success", "")
	c.Status(http.StatusNoContent)
}

func (h *APIHandler) getHandler(credType string) auth.Handler {
	handler, exists := h.handlers[credType]
	if !exists {
		// Fallback to simple JWT-like handler for unknown types
		return h.handlers["jwt"]
	}
	return handler
}

func (h *APIHandler) logAudit(ip, userAgent, action, name, status, details string) {
	entry := &store.AuditLog{
		Action:         action,
		CredentialName: name,
		SourceIP:       ip,
		SourceTool:     userAgent,
		Status:         status,
		Details:        details,
	}
	// Log audit in background with timeout to prevent goroutine leaks
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.store.AddAuditLog(ctx, entry); err != nil {
			log.Error().Err(err).Str("action", action).Str("name", name).Msg("failed to add audit log")
		}
	}()
}

func (h *APIHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
