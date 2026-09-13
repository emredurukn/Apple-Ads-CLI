# asactl (Apple Search Ads Control)

A fast, lightweight, and scriptable CLI for the **Apple Search Ads (ASA) API v5**. Automate campaigns, ad groups, keywords, and reporting workflows directly from your terminal or CI/CD pipelines.

Inspired by and modeled after [`rorkai/App-Store-Connect-CLI`](https://github.com/rorkai/App-Store-Connect-CLI) (`asc`).

---

## Features

- **OAuth 2.0 Client Credentials with ES256 JWT:** Native ECDSA signing from `.p8` private keys without external dependencies.
- **Secure Key Storage:** Credentials can be saved to macOS Keychain (via `go-keyring`) or stored in local configuration files (`--bypass-keychain`) for headless CI/CD.
- **Multi-Profile Support:** Seamlessly switch between multiple accounts and organizations (`--profile client-a`, `--profile staging`).
- **TTY-Aware Output:** Automatically formats as ASCII tables in interactive terminals and clean JSON when piped (`| jq`) or run in CI. Supports `--output table`, `--output json`, and `--output csv`.
- **Diagnostics (`auth doctor`):** Built-in connectivity, key parsing, and permission verification against Apple's servers.

---

## Installation

### From Source (Go 1.22+)

```bash
git clone https://github.com/emredurukan/asactl.git
cd asactl
make build
# Binary is generated at bin/asactl
```

To install globally to your `$GOPATH/bin`:

```bash
make install
```

---

## Quick Start

### 1. Verification

```bash
asactl version
asactl --help
```

### 2. Authentication

Generate an API key in the [Apple Search Ads Console](https://app-ads.apple.com) under **Account Settings > API**.

```bash
asactl auth login \
  --name "default" \
  --key-id "SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --team-id "SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --client-id "SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --org-id "1234567" \
  --private-key ~/.keys/AuthKey_ABC123.p8 \
  --validate
```

For CI/CD or headless environments without Keychain:

```bash
asactl auth login \
  --name "ci" \
  --bypass-keychain \
  --key-id "$ASACTL_KEY_ID" \
  --team-id "$ASACTL_TEAM_ID" \
  --client-id "$ASACTL_CLIENT_ID" \
  --org-id "$ASACTL_ORG_ID" \
  --private-key /path/to/AuthKey.p8
```

### 3. Check Authentication & Permissions

```bash
asactl auth status --validate
asactl auth doctor
```

### 4. Manage Campaigns

```bash
# List campaigns (default table view)
asactl campaigns list

# List campaigns as JSON
asactl campaigns list --output json

# Get single campaign details
asactl campaigns get 12345678
```

---

## Configuration

Configuration is stored in `$HOME/.asactl/config.yaml`:

```yaml
active_profile: default
profiles:
  default:
    name: default
    key_id: SEARCHADS.1234...
    team_id: SEARCHADS.1234...
    client_id: SEARCHADS.1234...
    org_id: "1234567"
    private_key_path: /Users/user/.keys/AuthKey.p8
    bypass_keychain: false
```

### Environment Variables

You can also configure via environment variables (prefixed with `ASACTL_` or `APPLE_ADS_`):

- `ASACTL_PROFILE`
- `ASACTL_KEY_ID`
- `ASACTL_TEAM_ID`
- `ASACTL_CLIENT_ID`
- `ASACTL_ORG_ID`
- `ASACTL_PRIVATE_KEY_PATH`

---

## Architecture

```
asactl/
├── cmd/              # Cobra CLI commands (root, auth, campaigns, version)
├── pkg/
│   ├── appleads/     # Apple Search Ads API v5 client & ES256 auth
│   ├── config/       # Multi-profile & Keychain credentials management
│   └── output/       # Table, JSON, and CSV rendering
├── internal/
│   └── version/      # Build and version information
└── Makefile          # Build and test tasks
```

---

## License

MIT License.
