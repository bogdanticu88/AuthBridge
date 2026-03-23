# AuthBridge: Centralized Authentication Daemon for Penetration Testing

```
    ╔════════════════════════════════╗
    ║                                ║
    ║      A U T H B R I D G E       ║
    ║   Credential Management Daemon ║
    ║                                ║
    ║   🔐  JWT  OAuth2  Kerberos   ║
    ║   🔑  mTLS  Basic  Cookie     ║
    ║                                ║
    ╚════════════════════════════════╝
         Burp Suite  |  Nuclei  |  API
           ╱───────────────────╲
        Connect your tools to one secure credential hub
```

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=flat-square)](https://github.com/bogdanticu88/AuthBridge)
[![Pentest Tool](https://img.shields.io/badge/Tool-Burp%20%7C%20Nuclei%20%7C%20sqlmap-orange?style=flat-square)](https://github.com/bogdanticu88/AuthBridge)

**AuthBridge** is a lightweight, high-performance authentication daemon designed to centralize and automate credential management during security engagements. Stop manually extracting JWTs and refreshing OAuth2 tokens across multiple tools - AuthBridge handles the lifecycle, so you can focus on the exploit.

---

## The Problem: Credential Fatigue in Pentesting

Modern web applications use complex, short-lived authentication (JWT, OAuth2, MFA). Pentesters typically spend **15-20% of their time** manually:
- Extracting tokens from browser dev tools.
- Injecting headers into Burp, Nuclei, and custom scripts.
- Dealing with expired sessions mid-scan.
- Managing multiple credential sets across different environments.

**AuthBridge solves this by acting as a local "Authentication Hub" that your tools query for fresh, valid credentials on-demand.**

---

## Features & Capabilities

| Feature | Description | Support |
| :--- | :--- | :--- |
| **JWT Management** | Automatic parsing, validation, and expiry warnings. | Yes |
| **OAuth2 Lifecycle** | Background token refreshing using stored Refresh Tokens. | Yes |
| **Secure Storage** | AES-256-GCM encryption with OS-native keyring integration. | Yes |
| **Audit Trails** | Immutably log every credential fetch for compliance. | Yes |
| **Multi-Tool** | Native plugins for Burp Suite and Nuclei + REST API. | Yes |
| **Web GUI** | High-density dashboard for credential administration. | Yes |
| **Cloud Sync** | Securely push/pull encrypted vaults to S3 for teams. | Yes |
| **Hardening** | Optional API Key protection for Web GUI and REST API. | Yes |

---

## Architecture & Data Flow

```text
authbridge/
├── cmd/                # CLI Entry points (start, add, list, audit, cloud)
├── internal/
│   ├── api/            # REST API (v1) & Embedded Web GUI
│   ├── auth/           # Specialized Handlers (JWT, OAuth2, Kerberos, mTLS)
│   ├── store/          # Encrypted SQLite (WAL mode) + Keyring
│   └── sync/           # Cloud Sync (S3) implementation
├── plugins/            # Burp (Java) and Nuclei (Go) integrations
└── main.go             # Application entry point
```

**Data Flow:**
1. **Store:** User adds credentials via CLI (Encrypted via Master Key in OS Keyring).
2. **Daemon:** `authbridge start` runs a local REST API (127.0.0.1:9999).
3. **Query:** Burp/Nuclei sends GET request to `/api/v1/token/{name}`.
4. **Auth:** AuthBridge validates/refreshes the token and returns it.
5. **Audit:** The access is logged with timestamp, IP, and tool ID.

---

## Quick Start

### 1. Installation (From Source)
```bash
# Clone and build
git clone https://github.com/bogdanticu88/AuthBridge
cd AuthBridge
go build -o authbridge ./main.go

# (Optional) Move to path
sudo mv authbridge /usr/local/bin/
```

### 2. Basic Usage
```bash
# 1. Start the daemon (Localhost by default)
authbridge start

# 2. Add a JWT credential (in another terminal)
authbridge add --name internal-api --token "eyJ..." --type jwt

# 3. Fetch the token programmatically
curl -s http://localhost:9999/api/v1/token/internal-api | jq .token
```

---

## Integration Examples

### Nuclei Integration
Inject tokens directly into your templates using a variable:
```bash
nuclei -t template.yaml -u https://api.target.com -var token=$(curl -s http://localhost:9999/api/v1/token/target-api | jq -r .token)
```

### Burp Suite Extension
1. Build the plugin: `cd plugins/burp && ./gradlew jar`
2. Load the `.jar` into Burp.
3. Configure the **AuthBridge Tab** inside Burp to point to your daemon.

---

## Security Model

- **Local Only:** Daemon binds to `127.0.0.1` by default.
- **Hardening:** Start with `--api-key <secret>` to require authentication for all requests.
- **Tiered Key Management:** Master key storage prioritizes OS Keyrings (Secret Service, Keychain, DPAPI) before falling back to encrypted local files.
- **Zero-Plaintext:** Sensitive credentials never touch the disk in plaintext.

---

## API Reference

### REST Endpoints

#### Get Token
```http
GET /api/v1/token/{name}
Authorization: X-AuthBridge-Key: <api-key> (optional)
```

**Response:**
```json
{
  "token": "eyJ...",
  "type": "jwt",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

#### List Credentials
```http
GET /api/v1/credentials
```

**Response:**
```json
{
  "credentials": [
    {
      "name": "internal-api",
      "type": "jwt",
      "usage_count": 42
    }
  ]
}
```

#### Add Credential
```http
POST /api/v1/credentials
Content-Type: application/json
{
  "name": "internal-api",
  "type": "jwt",
  "token": "eyJ...",
  "metadata": "{...}"
}
```

#### Delete Credential
```http
DELETE /api/v1/credentials/{name}
```

#### Audit Logs
```http
GET /api/v1/audit?name={name}&limit={limit}
```

#### Health Check
```http
GET /health
```

---

## Command-Line Interface

### Start Daemon
```bash
authbridge start [--port PORT] [--host HOST] [--api-key KEY]

Options:
  -p, --port     Port to listen on (default: 9999)
  -H, --host     Host to bind to (default: 127.0.0.1)
  --api-key      Optional API key for hardening
```

### Manage Credentials
```bash
# Add a credential
authbridge add --name <name> --type <type> --token <token> [--metadata <json>]

# List all credentials
authbridge list

# Remove a credential
authbridge remove <name>

# View audit logs
authbridge audit [--name <name>] [--limit <number>]
```

### Cloud Operations
```bash
# Push vault to S3
authbridge cloud push --bucket <bucket> --region <region> --key <s3-key>

# Pull vault from S3
authbridge cloud pull --bucket <bucket> --region <region> --key <s3-key>
```

---

## Supported Authentication Types

| Type | Description | Metadata | Use Case |
| :--- | :--- | :--- | :--- |
| **jwt** | JSON Web Token (static) | Optional expiry hints | API tokens, OAuth2 access tokens |
| **oauth2** | OAuth2 with auto-refresh | client_id, client_secret, refresh_token, token_url | Long-lived OAuth2 applications |
| **basic** | Base64-encoded username:password | username, password | Legacy HTTP Basic Auth |
| **cookie** | HTTP Cookie (static) | cookie_name, cookie_value | Session cookies |
| **kerberos** | Kerberos/SPNEGO | realm, username, krb5_conf, keytab | AD/Kerberos environments |
| **mtls** | Mutual TLS certificate | cert_path, key_path | Certificate-based auth |

---

## Development & Contributing

### Build from Source
```bash
# Prerequisites
go 1.25+
SQLite3 (libsqlite3-dev on Linux)
OpenSSL (for crypto)

# Build
go build -o authbridge ./main.go

# Run tests
go test -v ./...

# Run linters
go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...

# Build all tools
make build
```

### Project Structure
```
authbridge/
├── cmd/                      # CLI commands
│   ├── server.go            # Daemon startup
│   ├── add.go               # Add credentials
│   ├── list.go              # List credentials
│   ├── remove.go            # Remove credentials
│   ├── audit.go             # View audit logs
│   └── cloud.go             # S3 sync operations
├── internal/
│   ├── api/                 # HTTP API & Web UI
│   │   ├── server.go        # API server setup
│   │   ├── handlers.go      # Request handlers
│   │   └── web/             # Embedded web assets
│   ├── auth/                # Auth handlers
│   │   ├── jwt.go           # JWT validation
│   │   ├── oauth2.go        # OAuth2 refresh
│   │   ├── basic.go         # Basic auth
│   │   ├── cookie.go        # Cookie handling
│   │   ├── kerberos.go      # Kerberos/SPNEGO
│   │   └── mtls.go          # mTLS certs
│   ├── store/               # Encrypted storage
│   │   ├── sqlite.go        # SQLite backend (WAL mode)
│   │   ├── encryption.go    # AES-256-GCM
│   │   └── store.go         # Interface
│   └── sync/                # Cloud sync
│       └── s3.go            # S3 operations
├── plugins/                 # Tool integrations
│   ├── burp/               # Burp Suite extension
│   └── nuclei/             # Nuclei integration
├── tests/                   # Integration tests
├── main.go                  # Entry point
├── go.mod & go.sum         # Dependencies
├── Dockerfile              # Container build
└── .golangci.yml          # Linter config
```

### Code Quality Standards
- **Error Handling:** All errors are logged with context (no silent failures)
- **Concurrency:** Goroutines have timeouts to prevent leaks
- **Security:** No credentials in logs, plaintext, or debug output
- **Testing:** Integration tests verify end-to-end flows
- **Linting:** Enabled bodyclose, errcheck, gosec, contextcheck

### Running Tests
```bash
# Run all tests
go test -v ./...

# Run with race detector (finds concurrency bugs)
go test -race ./...

# Run specific test
go test -v -run TestFullLifecycle ./tests
```

### Debugging
```bash
# Enable debug logging
authbridge start --log-level debug

# Use curl with verbose output
curl -v http://localhost:9999/api/v1/token/test-cred

# Check audit logs
authbridge audit --limit 10
```

---

## Troubleshooting

### "Failed to get user home directory"
**Cause:** HOME environment variable not set
**Solution:** Set HOME or specify database path explicitly

### "Master key stored in OS keyring" (Info) but not persisting
**Cause:** OS keyring service not running or misconfigured
**Solution:** The key falls back to `~/.authbridge/master.key` file (check permissions: 0600)

### HTTP 401 - Unauthorized
**Cause:** API key was set but request doesn't include it
**Solution:** Add header: `curl -H "X-AuthBridge-Key: <your-key>" ...` or use `?api_key=<key>` query param

### Token refresh fails silently
**Cause:** Background goroutine error not visible in logs
**Solution:** Check audit logs for failures: `authbridge audit --limit 50`

### SQLite "database is locked"
**Cause:** WAL mode not working (usually file permissions)
**Solution:** Ensure `~/.authbridge/` has rwx permissions (0700)

---

## Roadmap

- [ ] **Phase 5:** Cloud sync (Team Encrypted) via shared team key.
- [ ] **Phase 6:** Proxy mode for tools that don't support custom headers.
- [ ] **Phase 7:** Native mobile app for credential approval (MFA).

---

## License

Distributed under the MIT License. See `LICENSE` for more information.
