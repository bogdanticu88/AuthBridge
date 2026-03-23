# 📖 AuthBridge Usage Guide

This guide provides a comprehensive walkthrough of managing credentials, configuring the daemon, and integrating AuthBridge with common penetration testing tools.

---

## 🛠️ CLI Reference

### Daemon Management
Before any tools can fetch credentials, the daemon must be running.
```bash
# Start the daemon on the default port (9999)
authbridge start

# Hardened mode: require an API key for all requests
authbridge start --api-key "my-secure-access-key"
```

### Credential Operations
**Adding a JWT:**
```bash
authbridge add --name my-api --token "eyJ..." --type jwt
```

**Adding OAuth2 (with auto-refresh):**
```bash
authbridge add --name github-oauth \
  --token "gho_access_token" \
  --type oauth2 \
  --metadata '{"client_id": "xxx", "client_secret": "yyy", "refresh_token": "ghr_...", "token_url": "https://github.com/login/oauth/access_token"}'
```

**Listing & Audit:**
```bash
authbridge list
authbridge audit --name internal-api
```

### Cloud Sync (Team Collaboration)
Push and pull your encrypted vault to S3-compatible storage.
```bash
# Push local vault
authbridge cloud push --bucket my-pentesters-vault --region us-east-1

# Pull vault on another machine
authbridge cloud pull --bucket my-pentesters-vault --region us-east-1
```

---

## 🖥️ Web GUI

Access the high-density dashboard at `http://localhost:9999`.

**If hardened with `--api-key`:**
Access via `http://localhost:9999/?api_key=your-key` or provide the `X-AuthBridge-Key` header.

The GUI allows you to:
- Monitor system health and master key status.
- Add and revoke credentials visually.
- Inspect the full audit trail.

---

## 🔌 Tool Integrations

### 1. Burp Suite
The new Burp plugin features a dedicated **AuthBridge Tab** for configuration.

1.  **Build:** `cd plugins/burp && ./gradlew jar`
2.  **Load:** In Burp, go to `Extensions` -> `Add` -> Select `AuthBridgePlugin.jar`.
3.  **UI Config:**
    - **Enabled Toggle:** Instantly stop/start injection.
    - **Credential Name:** Specify which vault object to fetch (defaults to `default`).
    - **Host Pattern:** Regex to control which requests receive the token.

### 2. Nuclei
```bash
nuclei -t template.yaml -u https://target.local \
  -var token=$(curl -s "http://localhost:9999/api/v1/token/api-prod?api_key=mykey" | jq -r .token)
```

---

## 🔐 Advanced Security

- **Master Key Rotation:** To force a re-initialization, delete `~/.authbridge/master.key` or remove the `authbridge/master-key` entry from your OS keyring.
- **Remote Security:** Always use SSH tunneling (`-L 9999:localhost:9999`) if accessing the daemon remotely.
