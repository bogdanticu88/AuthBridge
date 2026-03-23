package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "authbridge-store-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	s, err := NewSQLiteStore(dbPath)
	require.NoError(t, err)
	defer s.Close()

	ctx := context.Background()
	cred := &Credential{
		ID:        uuid.New().String(),
		Name:      "test-api",
		Type:      "jwt",
		Token:     "encrypted-token",
		CreatedAt: time.Now(),
	}

	// Test Add
	err = s.AddCredential(ctx, cred)
	assert.NoError(t, err)

	// Test Get
	retrieved, err := s.GetCredential(ctx, "test-api")
	assert.NoError(t, err)
	assert.Equal(t, cred.ID, retrieved.ID)
	assert.Equal(t, "jwt", retrieved.Type)

	// Test List
	list, err := s.ListCredentials(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// Test Update
	cred.Type = "oauth2"
	err = s.UpdateCredential(ctx, cred)
	assert.NoError(t, err)
	retrieved, _ = s.GetCredential(ctx, "test-api")
	assert.Equal(t, "oauth2", retrieved.Type)

	// Test Delete
	err = s.DeleteCredential(ctx, "test-api")
	assert.NoError(t, err)
	_, err = s.GetCredential(ctx, "test-api")
	assert.Equal(t, ErrCredentialNotFound, err)

	// Test Audit Log
	audit := &AuditLog{
		Action:         "token_fetch",
		CredentialName: "test-api",
		SourceIP:       "127.0.0.1",
		Status:         "success",
	}
	err = s.AddAuditLog(ctx, audit)
	assert.NoError(t, err)

	logs, err := s.ListAuditLogs(ctx, "test-api", 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "token_fetch", logs[0].Action)
}
