# Config File Reference

The `.envguardrc.yaml` file stores default options for EnvGuard.

## File Names

EnvGuard searches for (in order):

1. `.envguardrc.yaml`
2. `.envguardrc.yml`
3. `.envguardrc.json`

## Options

```yaml
# Path to schema file
schema: envguard.yaml

# Environment files to validate (merged right-to-left)
env:
  - .env
  - .env.local

# Output format
format: text

# Strict mode
strict: false

# Environment name
envName: ""

# Scan for secrets
scanSecrets: false

# Enable watch mode
watch: false

# Monorepo packages
packages:
  - apps/web
  - apps/api

# Minimum severity
severity: critical
```

## Environment Variable Overrides

Any config option can be overridden via environment variables:

| Variable | Type | Description |
|----------|------|-------------|
| `ENVGUARD_SCHEMA` | string | Schema file path |
| `ENVGUARD_FORMAT` | string | Output format |
| `ENVGUARD_STRICT` | bool | Strict mode |
| `ENVGUARD_SCAN_SECRETS` | bool | Secret scanning |
| `ENVGUARD_ENV_NAME` | string | Environment name |
| `ENVGUARD_WATCH` | bool | Watch mode |

```bash
ENVGUARD_SCHEMA=config/envguard.yaml envguard validate
ENVGUARD_FORMAT=json envguard validate
ENVGUARD_STRICT=true envguard validate
```

## JSON Format

```json
{
  "schema": "envguard.yaml",
  "env": [".env"],
  "format": "json",
  "strict": true
}
```
