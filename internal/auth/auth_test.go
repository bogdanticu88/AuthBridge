package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/twister69/authbridge/internal/store"
)

func TestJWTHandler(t *testing.T) {
	em, _ := store.NewEncryptionManager("01234567890123456789012345678901")
	h := NewJWTHandler(em)

	// Create a valid JWT
	claims := jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("secret"))

	encryptedToken, _ := em.Encrypt(tokenStr)
	cred := &store.Credential{
		Name:  "test-jwt",
		Type:  "jwt",
		Token: encryptedToken,
	}

	resp, err := h.Authenticate(context.Background(), cred)
	assert.NoError(t, err)
	assert.Equal(t, tokenStr, resp.Token)
	assert.Equal(t, "Bearer", resp.Type)
	assert.NotNil(t, resp.ExpiresAt)

	// Test expired JWT
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ = token.SignedString([]byte("secret"))
	encryptedToken, _ = em.Encrypt(tokenStr)
	cred.Token = encryptedToken

	_, err = h.Authenticate(context.Background(), cred)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestBasicAuthHandler(t *testing.T) {
	em, _ := store.NewEncryptionManager("01234567890123456789012345678901")
	h := NewBasicAuthHandler(em)

	plaintext := "admin:password123"
	encrypted, _ := em.Encrypt(plaintext)
	cred := &store.Credential{
		Name:  "test-basic",
		Type:  "basic",
		Token: encrypted,
	}

	resp, err := h.Authenticate(context.Background(), cred)
	assert.NoError(t, err)
	assert.Equal(t, "YWRtaW46cGFzc3dvcmQxMjM=", resp.Token)
	assert.Equal(t, "Basic", resp.Type)
}
