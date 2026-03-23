# AuthBridge: Centralized Authentication Daemon for Penetration Testing

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

## Roadmap

- [ ] **Phase 5:** Cloud sync (Team Encrypted) via shared team key.
- [ ] **Phase 6:** Proxy mode for tools that don't support custom headers.
- [ ] **Phase 7:** Native mobile app for credential approval (MFA).

---

## License

Distributed under the MIT License. See `LICENSE` for more information.
