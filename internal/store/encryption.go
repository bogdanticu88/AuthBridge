package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "authbridge"
	accountName = "master-key"
)

type EncryptionManager struct {
	masterKey []byte
}

func NewEncryptionManager(customKey string) (*EncryptionManager, error) {
	var key []byte
	var err error

	if customKey != "" {
		key = []byte(customKey)
	} else {
		key, err = getOrCreateMasterKey()
		if err != nil {
			return nil, fmt.Errorf("failed to get or create master key: %w", err)
		}
	}

	// Derive a 32-byte key using SHA-256
	hash := sha256.Sum256(key)
	return &EncryptionManager{masterKey: hash[:]}, nil
}

func getOrCreateMasterKey() ([]byte, error) {
	// 1. Try OS keyring
	keyStr, err := keyring.Get(serviceName, accountName)
	if err == nil {
		decoded, decodeErr := hex.DecodeString(keyStr)
		if decodeErr == nil {
			return decoded, nil
		}
	}

	// 2. Fallback to file in ~/.authbridge
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}
	keyPath := filepath.Join(home, ".authbridge", "master.key")
	if _, err := os.Stat(keyPath); err == nil {
		data, err := os.ReadFile(keyPath)
		if err == nil {
			decoded, decodeErr := hex.DecodeString(string(data))
			if decodeErr == nil {
				return decoded, nil
			}
		}
	}

	// 3. Create new key
	log.Info().Msg("Generating new master key...")
	newKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return nil, err
	}

	keyHex := hex.EncodeToString(newKey)

	// Try to store in keyring
	if err := keyring.Set(serviceName, accountName, keyHex); err == nil {
		log.Info().Msg("Master key stored in OS keyring")
	} else {
		log.Warn().Err(err).Msg("Failed to store key in OS keyring, falling back to file")
		// Fallback to file
		if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
			return nil, fmt.Errorf("failed to create directory for master key: %w", err)
		}
		if err := os.WriteFile(keyPath, []byte(keyHex), 0600); err != nil {
			return nil, fmt.Errorf("failed to write master key to file: %w", err)
		}
	}

	return newKey, nil
}

func (e *EncryptionManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (e *EncryptionManager) Decrypt(ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
