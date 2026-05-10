# EnvGuard

> Validate `.env` files against a declarative YAML schema. Catch misconfigurations before deployment.

[![CI](https://github.com/firasmosbehi/envguard/actions/workflows/ci.yml/badge.svg)](https://github.com/firasmosbehi/envguard/actions)
[![Test Action](https://github.com/firasmosbehi/envguard/actions/workflows/test-action.yml/badge.svg)](https://github.com/firasmosbehi/envguard/actions)

EnvGuard is a fast, language-agnostic CLI tool written in Go that validates environment variable files against a declarative YAML schema. It supports type coercion, regex patterns, enums, required fields, defaults, and strict mode — with wrapper packages for Node.js and Python, plus a native GitHub Action for CI/CD.

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Schema Specification](#schema-specification)
- [CLI Reference](#cli-reference)
- [Node.js Package](#nodejs-package)
- [Python Package](#python-package)
- [GitHub Action](#github-action)
- [Validation Rules](#validation-rules)
- [Type Coercion](#type-coercion)
- [Exit Codes](#exit-codes)
- [CI/CD Integration](#cicd-integration)
- [Architecture](#architecture)
- [Development](#development)
- [Changelog](#changelog)
- [License](#license)

---

## Installation

### macOS / Linux (Binary)

```bash
curl -sSL https://github.com/firasmosbehi/envguard/releases/latest/download/envguard-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/') -o /usr/local/bin/envguard
chmod +x /usr/local/bin/envguard
envguard version
```

### Windows (Binary)

Download `envguard-windows-amd64.exe` from the [latest release](https://github.com/firasmosbehi/envguard/releases/latest) and place it on your `PATH`.

### Node.js

```bash
npm install envguard-validator
```

### Python

```bash
pip install envguard-validator
```

### Homebrew (macOS/Linux)

```bash
brew install --formula https://raw.githubusercontent.com/firasmosbehi/envguard/main/homebrew/envguard.rb
```

### Docker

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/firasmosbehi/envguard:latest validate
```

### Build from Source

```bash
git clone https://github.com/firasmosbehi/envguard.git
cd envguard
make build
./bin/envguard version
```

---

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
    min: 1024
    max: 65535
    default: 3000
    description: "HTTP server port"

  DEBUG:
    type: boolean
    default: false
    description: "Enable debug mode"

  LOG_LEVEL:
    type: string
    enum: [debug, info, warn, error]
    default: info
    description: "Logging verbosity"

  API_KEY:
    type: string
    required: true
    minLength: 32
    pattern: "^[A-Za-z0-9_-]+$"
    description: "API authentication key"

  WEBHOOK_URL:
    type: string
    format: url
    requiredIn: [production, staging]
    description: "Webhook endpoint (required in prod/staging)"

  DEBUG_TOOL:
    type: string
    devOnly: true
    description: "Development-only debugging endpoint"

  ADMIN_EMAIL:
    type: string
    format: email
    description: "Admin contact email"

  ALLOWED_HOSTS:
    type: array
    separator: ","
    minLength: 1
    description: "Comma-separated list of allowed hostnames"

  CORS_ORIGINS:
    type: array
    separator: ","
    enum: ["http://localhost:3000", "https://app.example.com"]
    description: "Allowed CORS origins"
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
✗ Environment validation failed (3 error(s))

  • DATABASE_URL
    └─ required: variable is missing or empty

  • PORT
    └─ type: expected integer, got "eighty"

  • API_KEY
    └─ pattern: value does not match pattern ^[A-Za-z0-9_-]{32,}$
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
    { "key": "DATABASE_URL", "message": "variable is missing or empty", "rule": "required" },
    { "key": "PORT", "message": "expected integer, got 'eighty'", "rule": "type" }
  ],
  "warnings": [
    { "key": "OLD_VAR", "message": "not defined in schema", "rule": "strict" }
  ]
}
```

---

## Schema Specification

EnvGuard schemas are YAML files that declare the expected shape of your `.env` file.

### Top-level structure

```yaml
version: "1.0"           # Schema version (required)
env:                     # Map of variable names to definitions (required)
  VARIABLE_NAME:
    type: string
    required: true
    default: "fallback"
    description: "Human-readable docs"
    pattern: "^regex$"
    enum: [a, b, c]
```

### Field Reference

| Field | Types | Required | Description |
|-------|-------|----------|-------------|
| Field | Types | Required | Description |
|-------|-------|----------|-------------|
| `type` | all | **Yes** | `string`, `integer`, `float`, `boolean`, `array` |
| `required` | all | No | If `true`, variable must be present and non-empty |
| `default` | all | No | Fallback value injected when variable is absent |
| `description` | all | No | Human-readable docs, shown in errors |
| `message` | all | No | Custom error message shown when validation fails |
| `pattern` | `string` | No | Regex the value must match |
| `enum` | `string`, `integer`, `float`, `array` | No | Array of allowed values |
| `min` | `integer`, `float` | No | Minimum numeric value |
| `max` | `integer`, `float` | No | Maximum numeric value |
| `minLength` | `string`, `array` | No | Minimum length (chars for string, items for array) |
| `maxLength` | `string`, `array` | No | Maximum length (chars for string, items for array) |
| `format` | `string` | No | Built-in format: `email`, `url`, `uuid` |
| `disallow` | `string` | No | Array of forbidden string values |
| `requiredIn` | all | No | Array of environment names where variable is required |
| `devOnly` | all | No | If `true`, variable is only allowed in development |
| `separator` | `array` | No | Delimiter for splitting array items (default `,`) |
| `allowEmpty` | all | No | If `false`, reject empty strings even when optional |
| `contains` | `array` | No | Require array to contain this specific item |
| `dependsOn` | all | No | Name of another variable that triggers conditional requirement |
| `when` | all | No | Value the `dependsOn` variable must have to trigger requirement |

### Notes

- `required: true` and `default` are mutually exclusive in practice — if a variable is required, it must be present and a default is never applied.
- `enum` values must be compatible with the variable's `type`.
- `pattern` is only applied to `string` types.
- Empty enums (`enum: []`) are rejected as invalid schema definitions.
- Whitespace-only values (e.g., `"   "`) fail `required` checks.
- `devOnly: true` and `required` / `requiredIn` are mutually exclusive.
- `dependsOn` and `when` must be used together.
- `allowEmpty: false` is redundant when `required: true`.
- `min` cannot be greater than `max`; `minLength` cannot be greater than `maxLength`.

---

## CLI Reference

### Commands

```
envguard validate [flags]       Validate .env against schema
envguard init                   Generate a starter envguard.yaml
envguard generate-example       Generate .env.example from schema
envguard version                Print version
```

### `envguard validate` Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--env` | `-e` | `.env` | Path to `.env` file (repeatable for multiple files) |
| `--format` | `-f` | `text` | Output format: `text` or `json` |
| `--strict` | | `false` | Fail if `.env` contains keys not defined in schema |
| `--env-name` | | `""` | Environment name for `requiredIn`/`devOnly` rules |

### `envguard generate-example` Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--output` | `-o` | `.env.example` | Path to output file |

### Examples

```bash
# Default usage
envguard validate

# Custom paths
envguard validate --schema config/schema.yaml --env config/.env

# Multiple env files (merged right-to-left)
envguard validate -e .env -e .env.local -e .env.production

# JSON output for CI pipelines
envguard validate --format json

# Strict mode: catch unknown variables
envguard validate --strict

# Environment-specific validation
envguard validate --env-name production
```

---

## Node.js Package

Install:
```bash
npm install envguard-validator
```

### Async API

```typescript
import { validate } from "envguard-validator";

const result = await validate({
  schemaPath: "envguard.yaml",
  envPath: ".env",        // string or string[]
  strict: false,
  envName: "production",
});

if (!result.valid) {
  for (const error of result.errors) {
    console.log(`${error.key}: ${error.message}`);
  }
  process.exit(1);
}
```

### Sync API

```typescript
import { validateSync } from "envguard-validator";

const result = validateSync({ schemaPath: "envguard.yaml", envPath: ".env" });
```

### CLI

```bash
npx envguard-validator validate --schema envguard.yaml --env .env
```

The correct Go binary for your platform is downloaded automatically via a `postinstall` hook.

---

## Python Package

Install:
```bash
pip install envguard-validator
```

### API

```python
from envguard import validate

result = validate(
    schema_path="envguard.yaml",
    env_path=".env",           # str or list[str]
    strict=False,
    env_name="production",
)

if not result.valid:
    for error in result.errors:
        print(f"{error.key}: {error.message}")
    exit(1)

print("✓ Environment validated!")
```

### CLI

```bash
envguard-py validate --schema envguard.yaml --env .env
```

The correct Go binary is downloaded automatically on first use to `~/.envguard/bin/`.

---

## GitHub Action

Add EnvGuard validation to any GitHub Actions workflow:

```yaml
- uses: firasmosbehi/envguard@v0.1.7
  with:
    schema: envguard.yaml
    env: .env
    strict: false
    format: text
```

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `schema` | No | `envguard.yaml` | Path to schema YAML file |
| `env` | No | `.env` | Path to `.env` file |
| `strict` | No | `false` | Fail if `.env` contains keys not in schema |
| `env-name` | No | `""` | Environment name for `requiredIn`/`devOnly` rules |
| `version` | No | `0.1.6` | EnvGuard version to download |
| `format` | No | `text` | Output format: `text` or `json` |

### Example Workflow

```yaml
name: Validate Environment

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: firasmosbehi/envguard@v0.1.7
```

---

## Validation Rules

### `required`
- **Applies to:** all types
- **Behavior:** The variable must be present in `.env` and its value must be non-empty after trimming whitespace.
- **Fails on:** missing key, empty string `""`, whitespace-only string `"   "`

### `default`
- **Applies to:** all types
- **Behavior:** If the variable is missing from `.env`, the default value is used as if it were present.
- **Note:** Defaults are type-checked against the variable's `type` at schema parse time.

### `pattern`
- **Applies to:** `string` only
- **Behavior:** The value must match the given regular expression.
- **Fails on:** non-matching strings

### `enum`
- **Applies to:** `string`, `integer`, `float`, `array`
- **Behavior:** The coerced value must be one of the listed values. For arrays, each item is checked individually.
- **Fails on:** values not in the list
- **Note:** Empty enums (`enum: []`) are rejected as invalid schema definitions.

### `min` / `max`
- **Applies to:** `integer`, `float`
- **Behavior:** Numeric value must be within the specified bounds (inclusive).
- **Fails on:** values outside the range

### `minLength` / `maxLength`
- **Applies to:** `string` (character count), `array` (item count)
- **Behavior:** Length must be within the specified bounds (inclusive).
- **Fails on:** strings too short/long, arrays with too few/many items

### `format`
- **Applies to:** `string`
- **Behavior:** Built-in format validation.
- **Supported:** `email`, `url`, `uuid`
- **Fails on:** malformed emails, URLs missing scheme/host, invalid UUIDs

### `disallow`
- **Applies to:** `string`
- **Behavior:** Rejects specific string values.
- **Fails on:** exact matches to any value in the list

### `requiredIn`
- **Applies to:** all types
- **Behavior:** Variable is required only in the specified environments.
- **Use with:** `--env-name` CLI flag
- **Example:** `requiredIn: [production, staging]`

### `devOnly`
- **Applies to:** all types
- **Behavior:** Variable is only allowed in development. Skipped in non-dev environments; required in dev.
- **Mutually exclusive with:** `required`, `requiredIn`

### `message`
- **Applies to:** all types
- **Behavior:** Overrides the default error message for any validation failure on this variable.
- **Example:** `message: "PORT must be a valid port number (1024-65535)"`

### `separator`
- **Applies to:** `array`
- **Behavior:** Defines the delimiter used to split the string into array items.
- **Default:** `,`
- **Example:** `separator: "|"` for `"read|write|admin"`

### `allowEmpty`
- **Applies to:** all types
- **Behavior:** When `false`, rejects empty strings even for optional variables.
- **Default:** `true` (nil = no restriction)
- **Example:** `allowEmpty: false` ensures the variable always has a value

### `contains`
- **Applies to:** `array`
- **Behavior:** Requires the array to contain at least one item matching the given value.
- **Example:** `contains: "admin"` ensures `"admin"` is in the roles list

### `dependsOn` / `when`
- **Applies to:** all types
- **Behavior:** Makes this variable conditionally required. If the `dependsOn` variable has the value `when`, this variable becomes required.
- **Example:**
  ```yaml
  HTTPS:
    type: boolean
  SSL_CERT:
    type: string
    dependsOn: HTTPS
    when: "true"
  ```

### `strict` mode
- **Applies to:** entire validation run
- **Behavior:** Any key present in `.env` but not defined in the schema generates a warning.
- **Use case:** catching typos, deprecated variables, or environment drift

---

## Type Coercion

EnvGuard parses all `.env` values as strings, then coerces them to the declared type:

| Type | Valid Input | Coerced Value | Invalid Input |
|------|-------------|---------------|---------------|
| `string` | any text | trimmed string | (never fails) |
| `integer` | `"42"`, `"-3"` | `42`, `-3` | `"3.14"`, `"abc"` |
| `float` | `"3.14"`, `"-0.5"` | `3.14`, `-0.5` | `"abc"` |
| `boolean` | `"true"`, `"1"`, `"yes"`, `"on"` | `true` | `"maybe"` |
| `boolean` | `"false"`, `"0"`, `"no"`, `"off"` | `false` | `"maybe"` |
| `array` | `"a,b,c"` | `["a", "b", "c"]` | `""` (empty) |

Boolean coercion is case-insensitive.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Validation passed |
| `1` | Validation failed (missing/invalid variables) |
| `2` | I/O or schema parsing error |

All wrappers (Node.js, Python, GitHub Action) preserve these exit codes.

---

## CI/CD Integration

### GitHub Actions
See [GitHub Action](#github-action) above.

### GitLab CI

```yaml
validate-env:
  image: alpine/curl
  script:
    - curl -sSL https://github.com/firasmosbehi/envguard/releases/latest/download/envguard-linux-amd64 -o /usr/local/bin/envguard
    - chmod +x /usr/local/bin/envguard
    - envguard validate --format json
```

### CircleCI

```yaml
jobs:
  validate-env:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      - run:
          name: Validate environment
          command: |
            curl -sSL https://github.com/firasmosbehi/envguard/releases/latest/download/envguard-linux-amd64 -o envguard
            chmod +x envguard
            ./envguard validate
```

### Pre-commit Hook

EnvGuard includes a [pre-commit](https://pre-commit.com/) hook definition. Add it to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/firasmosbehi/envguard
    rev: v0.1.7
    hooks:
      - id: envguard-validate
```

The hook runs `envguard validate` on all `.env` files before each commit.

---

## Architecture

```
envguard/
├── cmd/envguard/          # CLI entrypoint
│   └── main.go
├── internal/
│   ├── cli/               # Cobra commands (validate, init, version)
│   ├── schema/            # YAML schema parsing & validation
│   ├── dotenv/            # .env file parser
│   ├── validator/         # Validation engine
│   └── reporter/          # Output formatters (text, json)
├── packages/
│   ├── node/              # Node.js wrapper (envguard-validator on npm)
│   └── python/            # Python wrapper (envguard-validator on PyPI)
├── action.yml             # GitHub Action definition
├── examples/              # Sample schema and .env files
├── Makefile
└── README.md
```

### Design Principles

1. **Single binary** — The Go CLI compiles to a single static binary with no runtime dependencies.
2. **Language-agnostic core** — All validation logic lives in Go. Language packages are thin wrappers that spawn the CLI and parse JSON.
3. **Auto-distribution** — Wrappers download the correct platform binary from GitHub releases automatically.
4. **Fail fast, report all** — Validation never short-circuits on the first error; all issues are collected and reported.

---

## Development

### Requirements

- Go 1.22+
- Node.js 16+ (for Node.js wrapper)
- Python 3.8+ (for Python wrapper)

### Running Tests

```bash
# Go unit tests + race detector
make test

# E2E tests
./scripts/e2e.sh

# Build all platform binaries
make build-all
```

### Project Scripts

```bash
make build      # Build local binary
make test       # Run Go tests with coverage
make lint       # Run golangci-lint
make clean      # Remove build artifacts
make run        # Build and run locally
```

### Releasing

1. Bump versions in all packages (`cmd/envguard/main.go`, `packages/node/`, `packages/python/`)
2. Commit and push to `main`
3. Create and push a tag:
   ```bash
   git tag v0.1.5
   git push origin v0.1.5
   ```
4. GitHub Actions automatically:
   - Builds cross-platform binaries
   - Creates a GitHub Release
   - Publishes `envguard-validator` to npm
   - Publishes `envguard-validator` to PyPI

---

## Changelog

### v0.1.5
- Added `min`/`max` validation for integers and floats
- Added `minLength`/`maxLength` validation for strings
- Added `format` validator: `email`, `url`, `uuid`
- Added `disallow` list for strings
- Added environment-specific rules: `requiredIn` and `devOnly`
- Added `envguard generate-example` command
- Added `--env-name` flag for environment-specific validation

### v0.1.4
- Renamed packages to `envguard-validator` (npm + PyPI)
- Added GitHub Action (`action.yml`)
- Added automated publish workflows for npm and PyPI
- Cross-platform release builds (linux/amd64, darwin/amd64, darwin/arm64, windows/amd64)

### v0.1.1
- Fixed scanner crash on values >64KB
- Fixed empty enum `[]` being ignored
- Fixed whitespace-only values passing `required` check
- Fixed JSON output polluted with stderr text
- Fixed CI `make test` failure on clean runners
- Added Node.js and Python wrapper packages

### v0.1.0
- Initial release
- `validate`, `init`, `version` commands
- Schema types: `string`, `integer`, `float`, `boolean`
- Rules: `required`, `default`, `pattern`, `enum`
- Strict mode for unknown key detection
- Text and JSON output formats
- 90+ unit tests, 21 E2E tests

---

## License

MIT
