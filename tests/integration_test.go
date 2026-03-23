package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/bogdanticu88/AuthBridge/internal/api"
	"github.com/bogdanticu88/AuthBridge/internal/store"
	"github.com/zalando/go-keyring"
)

func TestFullLifecycle(t *testing.T) {
	keyring.MockInit()
	
	tempDir, err := os.MkdirTemp("", "authbridge-int-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	defer s.Close()

	e, err := store.NewEncryptionManager("01234567890123456789012345678901")
	require.NoError(t, err)

	addr := "127.0.0.1:0" // Pick a free port
	server := api.NewServer(addr, s, e, "")
	
	// Start server in background
	ln, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	actualAddr := ln.Addr().String()
	ln.Close()

	srv := &http.Server{
		Addr:    actualAddr,
		Handler: server.Handler,
	}

	go func() {
		srv.ListenAndServe()
	}()
	defer srv.Shutdown(context.Background())

	time.Sleep(100 * time.Millisecond) // Wait for server

	// 1. Add Credential via API
	addReq := map[string]string{
		"name":  "api1",
		"type":  "jwt",
		"token": "my-secret-jwt",
	}
	body, _ := json.Marshal(addReq)
	resp, err := http.Post(fmt.Sprintf("http://%s/api/v1/credentials", actualAddr), "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 2. Fetch Token via API
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/token/api1", actualAddr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.Equal(t, "my-secret-jwt", tokenResp["token"])

	// 3. List Credentials via API
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/credentials", actualAddr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResp)
	creds := listResp["credentials"].([]interface{})
	assert.Len(t, creds, 1)
}
