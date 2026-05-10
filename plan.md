# EnvGuard — Architecture & Design Plan

## Vision
A fast, language-agnostic CLI tool that validates `.env` files against a declarative YAML schema. EnvGuard catches misconfigurations before deployment and serves as the foundation for future language-specific packages (Node.js, Python, Java, etc.).

## User Choices
- **Name:** EnvGuard (override/rename strategy — scoped packages like `@envguard/cli`, `@envguard/node`)
- **Core CLI Language:** Go (single binary, fast, cross-platform)
- **Initial Delivery:** CLI only; language packages in later phases
- **MVP Schema Features:** Types, Required, Defaults, Regex patterns, Enums

---

## 1. Project Structure (Go Monorepo)

```
envguard/
├── cmd/envguard/              # CLI entrypoint
│   └── main.go
├── internal/
│   ├── cli/                   # Cobra commands (validate, init, version)
│   ├── schema/                # YAML schema parsing & model
│   ├── dotenv/                # .env file parser
│   ├── validator/             # Validation engine
│   └── reporter/              # Output formatters (text, json)
├── pkg/
│   └── envguard/              # Public Go API (for future packages)
├── schemas/
│   └── env-schema-v1.json     # JSON Schema for the YAML schema itself
├── examples/
│   ├── envguard.yaml          # Sample schema
│   └── .env                   # Sample .env
├── Makefile
├── go.mod
└── README.md
```

---

## 2. Schema Specification (YAML)

The user defines environment variables in a single YAML file:

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
    pattern: "^[A-Za-z0-9_-]{32,}$"
    description: "API authentication key"
```

### Field Reference

| Field | Types | Required | Description |
|-------|-------|----------|-------------|
| `type` | all | Yes | `string`, `integer`, `float`, `boolean` |
| `required` | all | No | If `true`, variable must be present and non-empty |
| `default` | all | No | Fallback value when variable is absent (ignored if `required: true`) |
| `description` | all | No | Human-readable docs, shown in errors |
| `pattern` | `string` | No | Regex the value must match |
| `enum` | `string`, `integer`, `float` | No | Allowed values list |

### Type Coercion Rules
- **string:** kept as-is (trimmed)
- **integer:** parsed with `strconv.Atoi`; fails on non-integer strings
- **float:** parsed with `strconv.ParseFloat`; fails on non-numeric strings
- **boolean:** accepts `true`/`1`/`yes`/`on` and `false`/`0`/`no`/`off` (case-insensitive)

---

## 3. CLI Design

### Commands

```bash
# Validate .env against schema (default)
envguard validate --schema envguard.yaml --env .env

# Same, with JSON output for CI
envguard validate --schema envguard.yaml --env .env --format json

# Generate a starter schema file
envguard init

# Print version
envguard version
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML |
| `--env` | `-e` | `.env` | Path to .env file |
| `--format` | `-f` | `text` | Output format: `text` or `json` |
| `--strict` | | `false` | Fail if .env contains keys not in schema |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Validation passed |
| `1` | Validation failed (missing/invalid variables) |
| `2` | I/O or schema parsing error |

---

## 4. Validation Engine

### Validation Flow
1. Parse schema YAML into internal model
2. Validate schema itself against JSON Schema meta-validator
3. Parse `.env` file into `map[string]string`
4. For each variable in schema:
   - If `required: true` → check presence and non-emptiness
   - If missing and `default` exists → inject default
   - Coerce value to declared `type`
   - If `enum` → check membership
   - If `pattern` → check regex match
5. If `--strict` → check for keys in `.env` not defined in schema
6. Collect all errors; do not short-circuit on first failure
7. Output results

### Error Format (Internal)
```go
type ValidationError struct {
    Key     string `json:"key"`
    Message string `json:"message"`
    Rule    string `json:"rule"`   // e.g. "required", "type", "pattern"
}
```

---

## 5. Output Formats

### Text (Human)
```
✗ Environment validation failed (3 errors)

  • DATABASE_URL
    └─ required: variable is missing

  • PORT
    └─ type: expected integer, got "eighty"

  • API_KEY
    └─ pattern: value does not match regex ^[A-Za-z0-9_-]{32,}$
```

### JSON (CI)
```json
{
  "valid": false,
  "errors": [
    { "key": "DATABASE_URL", "message": "variable is missing", "rule": "required" },
    { "key": "PORT", "message": "expected integer, got 'eighty'", "rule": "type" },
    { "key": "API_KEY", "message": "value does not match regex", "rule": "pattern" }
  ],
  "warnings": [
    { "key": "OLD_VAR", "message": "not defined in schema", "rule": "strict" }
  ]
}
```

---

## 6. Future Language Packages

The CLI is the universal validator. Future packages wrap it:

| Package | Language | Integration Pattern |
|---------|----------|---------------------|
| `@envguard/node` | Node.js / TS | Spawn CLI, parse JSON output, provide typed JS API |
| `envguard-py` | Python | Python wrapper calling the Go binary |
| `envguard-java` | Java | JAR wrapper / JNI binding |
| `envguard-rust` | Rust | Direct crate dependency on `pkg/envguard` |

All packages share the **same YAML schema format** and validation rules, ensuring consistency across stacks.

---

## 7. Non-Goals (MVP)
- No environment-specific conditional rules (e.g. `required_in: [prod]`)
- No secret security scanning
- No drift detection / `.env.example` auto-generation
- No OS keychain integration
- No GUI / web dashboard

These are reserved for v2.

---

## 8. Dependencies (Go)
- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- Standard library for everything else (regex, JSON, file I/O)
