# EnvGuard

> Validate `.env` files against a declarative YAML schema. Catch misconfigurations before deployment.

[![CI](https://github.com/envguard/envguard/actions/workflows/ci.yml/badge.svg)](https://github.com/envguard/envguard/actions)

## Install

### macOS / Linux
```bash
curl -sSL https://github.com/envguard/envguard/releases/latest/download/envguard-$(uname -s)-$(uname -m) -o /usr/local/bin/envguard
chmod +x /usr/local/bin/envguard
```

### Or build from source
```bash
git clone https://github.com/envguard/envguard.git
cd envguard
make build
```

## Quick Start

### 1. Define a schema (`envguard.yaml`)

```yaml
version: "1.0"

env:
  DATABASE_URL:
    type: string
    required: true
    description: "PostgreSQL connection string"

  PORT:
    type: integer
    default: 3000
    description: "Server port"

  DEBUG:
    type: boolean
    default: false

  LOG_LEVEL:
    type: string
    enum: [debug, info, warn, error]
    default: info

  API_KEY:
    type: string
    required: true
    pattern: "^[A-Za-z0-9_-]{32,}$"
```

Generate a starter file with:
```bash
envguard init
```

### 2. Validate your `.env`

```bash
envguard validate
```

**Success:**
```
✓ All environment variables validated.
```

**Failure:**
```
✗ Environment validation failed (2 error(s))

  • DATABASE_URL
    └─ required: variable is missing or empty

  • PORT
    └─ type: expected integer, got "eighty"
```

### 3. Use in CI/CD

```bash
envguard validate --format json
```

JSON output (for machine parsing):
```json
{
  "valid": false,
  "errors": [
    { "key": "DATABASE_URL", "message": "variable is missing or empty", "rule": "required" }
  ],
  "warnings": []
}
```

## CLI Reference

```
envguard validate [flags]

Flags:
  -s, --schema string   Path to schema YAML file (default "envguard.yaml")
  -e, --env string      Path to .env file (default ".env")
  -f, --format string   Output format: text or json (default "text")
      --strict          Fail if .env contains keys not defined in schema
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Validation passed |
| `1` | Validation failed |
| `2` | I/O or schema parsing error |

## Schema Reference

| Field | Types | Required | Description |
|-------|-------|----------|-------------|
| `type` | all | **Yes** | `string`, `integer`, `float`, `boolean` |
| `required` | all | No | If `true`, variable must be present and non-empty |
| `default` | all | No | Fallback value when variable is absent |
| `description` | all | No | Human-readable docs |
| `pattern` | `string` | No | Regex the value must match |
| `enum` | `string`, `integer`, `float` | No | Allowed values list |

### Type Coercion

- **string:** kept as-is (trimmed)
- **integer:** parsed with base-10; rejects non-integer strings
- **float:** parsed as floating-point; rejects non-numeric strings
- **boolean:** accepts `true`/`1`/`yes`/`on` and `false`/`0`/`no`/`off` (case-insensitive)

## Language Packages

EnvGuard is also available as a library for popular languages:

| Package | Install | Docs |
|---------|---------|------|
| **Node.js** | `npm install @envguard/node` | [packages/node/README.md](packages/node/README.md) |
| **Python** | `pip install envguard` | [packages/python/README.md](packages/python/README.md) |

All packages share the **same YAML schema format** and use the same Go binary under the hood.

## Roadmap

- [x] CLI tool with YAML schema validation
- [x] Type coercion (string, integer, float, boolean)
- [x] Rules: required, default, pattern, enum
- [x] Text and JSON output
- [x] Strict mode (warn on unknown keys)
- [x] Node.js package (`@envguard/node`)
- [x] Python package (`envguard`)
- [ ] Java package (`envguard-java`)
- [ ] Environment-specific conditional rules
- [ ] Secret security scanning
- [ ] Drift detection / `.env.example` generation

## License

MIT
