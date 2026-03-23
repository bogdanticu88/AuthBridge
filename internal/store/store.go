package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
)

type Credential struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Token       string    `json:"token"`
	Metadata    string    `json:"metadata"` // JSON string
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
	UsageCount  int       `json:"usage_count"`
}

type AuditLog struct {
	ID             int       `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	Action         string    `json:"action"`
	CredentialName string    `json:"credential_name"`
	SourceIP       string    `json:"source_ip"`
	SourceTool     string    `json:"source_tool"`
	Status         string    `json:"status"`
	Details        string    `json:"details"`
}

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	HttpOnly bool   `json:"http_only"`
	Secure   bool   `json:"secure"`
}

type Store interface {
	GetCredential(ctx context.Context, name string) (*Credential, error)
	AddCredential(ctx context.Context, cred *Credential) error
	UpdateCredential(ctx context.Context, cred *Credential) error
	DeleteCredential(ctx context.Context, name string) error
	ListCredentials(ctx context.Context) ([]*Credential, error)
	UpdateLastUsed(ctx context.Context, name string) error
	AddAuditLog(ctx context.Context, log *AuditLog) error
	ListAuditLogs(ctx context.Context, name string, limit int) ([]*AuditLog, error)
	Close() error
}
