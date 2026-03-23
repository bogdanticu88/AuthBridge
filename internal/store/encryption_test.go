package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestEncryptionManager(t *testing.T) {
	keyring.MockInit()
	// Setup a temporary directory for master.key fallback
	tempHome, err := os.MkdirTemp("", "authbridge-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// Mock home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Test with a custom key
	customKey := "01234567890123456789012345678901" // 32 bytes
	em, err := NewEncryptionManager(customKey)
	require.NoError(t, err)

	plaintext := "secret-token-123"
	ciphertext, err := em.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := em.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// Test with auto-generated key (fallback to file)
	keyring.Delete(serviceName, accountName) // Force it to NOT find it in keyring
	em2, err := NewEncryptionManager("")
	require.NoError(t, err)

	ciphertext2, err := em2.Encrypt(plaintext)
	assert.NoError(t, err)

	decrypted2, err := em2.Decrypt(ciphertext2)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}
