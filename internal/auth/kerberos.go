package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/bogdanticu88/AuthBridge/internal/store"
)

type KerberosMetadata struct {
	Realm    string `json:"realm"`
	Username string `json:"username"`
	Krb5Conf string `json:"krb5_conf"` // Path to krb5.conf
	Keytab   string `json:"keytab"`    // Base64 encoded keytab or path
}

type KerberosHandler struct {
	encryption *store.EncryptionManager
}

func NewKerberosHandler(e *store.EncryptionManager) *KerberosHandler {
	return &KerberosHandler{encryption: e}
}

func (h *KerberosHandler) Authenticate(ctx context.Context, cred *store.Credential) (*Response, error) {
	var meta KerberosMetadata
	if err := json.Unmarshal([]byte(cred.Metadata), &meta); err != nil {
		return nil, fmt.Errorf("invalid kerberos metadata: %w", err)
	}

	cfg, err := config.Load(meta.Krb5Conf)
	if err != nil {
		// Fallback to default if not provided
		cfg = config.New()
		cfg.LibDefaults.DefaultRealm = meta.Realm
	}

	var cl *client.Client
	password, err := h.encryption.Decrypt(cred.Token)
	if err != nil {
		return nil, err
	}

	if meta.Keytab != "" {
		// Keytab auth
		ktData, err := base64.StdEncoding.DecodeString(meta.Keytab)
		if err != nil {
			return nil, fmt.Errorf("failed to decode keytab: %w", err)
		}
		kt := keytab.New()
		if err := kt.Unmarshal(ktData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal keytab: %w", err)
		}
		cl = client.NewWithKeytab(meta.Username, meta.Realm, kt, cfg)
	} else {
		// Password auth
		cl = client.NewWithPassword(meta.Username, meta.Realm, password, cfg)
	}

	if err := cl.Login(); err != nil {
		return nil, fmt.Errorf("kerberos login failed: %w", err)
	}

	// In Kerberos, "token" could be the SPNEGO token for a specific service,
	// but for a general "TGT" fetch, we might return the TGT in Base64
	// or manage a local ccache.
	
	return &Response{
		Token: "TGT_ACQUIRED", // Placeholder for now
		Type:  "Kerberos",
	}, nil
}

func (h *KerberosHandler) Refresh(ctx context.Context, cred *store.Credential) (*Response, error) {
	return h.Authenticate(ctx, cred) // Re-login is essentially a refresh
}

func (h *KerberosHandler) Validate(ctx context.Context, cred *store.Credential) (bool, error) {
	return true, nil
}
