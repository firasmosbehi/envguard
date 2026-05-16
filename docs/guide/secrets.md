# Secrets Scanning

EnvGuard can scan `.env` files for hardcoded secrets, API keys, and tokens.

## Built-in Rules

The `scan` command uses 18 built-in detection rules:

| Rule | Pattern | Severity |
|------|---------|----------|
| `aws-access-key` | `AKIA[0-9A-Z]{16}` | High |
| `aws-secret-key` | `^[A-Za-z0-9/+=]{40}$` | Critical |
| `github-token` | `gh[pousr]_[A-Za-z0-9_]{36,}` | High |
| `private-key` | `-----BEGIN (RSA \|EC \|DSA \|OPENSSH )?PRIVATE KEY-----` | Critical |
| `generic-api-key` | `(?i)(api[_-]?key\|apikey)\s*[:=]\s*['"]?([a-z0-9_\-]{16,})['"]?` | Medium |
| `slack-token` | `xox[baprs]-[0-9]{10,13}-[0-9]{10,13}(-[a-zA-Z0-9]{24})?` | High |
| `stripe-key` | `sk_(live\|test)_[0-9a-zA-Z_]{24,}` | Critical |
| `jwt-token` | `eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*` | Medium |
| `azure-key` | `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}` | Medium |
| `gcp-api-key` | `AIza[0-9A-Za-z_-]{35}` | High |
| `telegram-bot-token` | `[0-9]+:AA[0-9A-Za-z_-]{32,}` | High |
| `sendgrid-api-key` | `SG\.[0-9A-Za-z_-]{20,24}\.[0-9A-Za-z_-]{40,50}` | High |
| `twilio-api-key` | `SK[0-9a-f]{32}` | High |
| `npm-token` | `npm_[0-9A-Za-z]{36}` | High |
| `docker-config-auth` | `"auth"\s*:\s*"[A-Za-z0-9+/=]+"` | Medium |
| `firebase-api-key` | `AIza[0-9A-Za-z_-]{35}` | Medium |
| `anthropic-api-key` | `sk-ant-api[0-9A-Za-z_-]{32,}` | Critical |
| `openai-api-key` | `sk-(proj-\|svcacct-)[0-9A-Za-z_-]{40,}` | Critical |

## Scan Command

```bash
# Scan a .env file
envguard scan -e .env

# Scan multiple files
envguard scan -e .env -e .env.local

# JSON output
envguard scan -e .env -f json

# SARIF output for GitHub Advanced Security
envguard scan -e .env -f sarif
```

![Secret scan output detecting AWS and GitHub tokens](/screenshots/scan-secrets.png)

## Validate with Secret Scanning

```bash
envguard validate --scan-secrets
```

This runs validation and secret scanning in a single pass.

## Custom Rules

Define custom rules in your `envguard.yaml`:

```yaml
version: "1.0"

env:
  # ... your variables

secrets:
  custom:
    - name: "internal-api-token"
      pattern: "iat_[a-zA-Z0-9]{32}"
      message: "Internal API token detected"
      severity: "high"
```

Custom rules are loaded by `envguard scan --schema` and `envguard validate --scan-secrets`.

## Entropy Detection

High-entropy strings that don't match any rule are also flagged:

```bash
# High-entropy string detected
MY_SECRET=abcdefghijklmnopqrstuvwxyz123456
```

Entropy detection helps catch unknown secret formats. It is skipped for common non-secret values like URLs, UUIDs, and version strings.

## Baseline

Suppress known false positives with a baseline file:

```bash
envguard scan -e .env --baseline secrets-baseline.json
```

Create a baseline from the current scan results:

```bash
envguard scan -e .env -f json > secrets-baseline.json
```
