# EnvGuard — Architecture & Design Plan

## Vision
A fast, language-agnostic CLI tool that validates `.env` files against a declarative YAML schema. EnvGuard catches misconfigurations before deployment and serves as the foundation for future language-specific packages (Node.js, Python, Java, etc.).

## Current Status
**v1.0.0 Released** — Core CLI, wrappers (Node.js, Python), GitHub Action, VS Code extension, Docker, Homebrew, and pre-commit hook are all functional.

## v2.0.0 Theme: "Intelligence & Integration"
Move beyond passive validation to active environment management: auto-detect issues, keep files in sync, and integrate deeper into developer workflows.

---

## 1. Project Structure (Go Monorepo)

```
envguard/
├── cmd/envguard/              # CLI entrypoint
│   └── main.go
├── internal/
│   ├── cli/                   # Cobra commands (validate, init, version, scan, lint, audit, sync, watch)
│   ├── schema/                # YAML schema parsing & model
│   ├── dotenv/                # .env file parser
│   ├── validator/             # Validation engine
│   ├── reporter/              # Output formatters (text, json, github, sarif)
│   ├── secrets/               # Secret scanning engine
│   ├── audit/                 # Source code auditing (new in v2)
│   ├── sync/                  # .env ↔ .env.example sync (new in v2)
│   └── config/                # Config file parser (.envguardrc) (new in v2)
├── pkg/
│   └── envguard/              # Public Go API
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

## 2. Schema Specification (YAML) — v1.0 (Current)

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

### Field Reference (v1.0 — Current)

| Field | Types | Required | Description |
|-------|-------|----------|-------------|
| `type` | all | Yes | `string`, `integer`, `float`, `boolean`, `array` |
| `required` | all | No | If `true`, variable must be present and non-empty |
| `default` | all | No | Fallback value when variable is absent (ignored if `required: true`) |
| `description` | all | No | Human-readable docs, shown in errors |
| `pattern` | `string` | No | Regex the value must match |
| `enum` | `string`, `integer`, `float`, `array` | No | Allowed values list |
| `min` / `max` | `integer`, `float` | No | Numeric bounds |
| `minLength` / `maxLength` | `string`, `array` | No | Length bounds |
| `format` | `string` | No | Built-in: `email`, `url`, `uuid`, `base64`, `ip`, `port`, `json`, `duration`, `semver`, `hostname`, `hex`, `cron` |
| `disallow` | `string` | No | Forbidden values |
| `requiredIn` | all | No | Environments where required |
| `devOnly` | all | No | Variable only allowed in development |
| `separator` | `array` | Yes* | Delimiter for array items |
| `allowEmpty` | all | No | If `false`, reject empty strings |
| `contains` | `array` | No | Require specific array item |
| `dependsOn` / `when` | all | No* | Conditional requirement |
| `deprecated` | all | No | Warning when variable is present |
| `sensitive` | all | No | Redact value in output |
| `transform` | `string` | No | Pre-validation: `lowercase`, `uppercase`, `trim` |

### Type Coercion Rules
- **string:** kept as-is (trimmed)
- **integer:** parsed with `strconv.Atoi`; fails on non-integer strings
- **float:** parsed with `strconv.ParseFloat`; fails on non-numeric strings
- **boolean:** accepts `true`/`1`/`yes`/`on` and `false`/`0`/`no`/`off` (case-insensitive)
- **array:** split by `separator`, validate each item

---

## 3. CLI Design — v2.0.0

### Commands

```bash
# Validate .env against schema (default)
envguard validate --schema envguard.yaml --env .env

# Same, with JSON output for CI
envguard validate --schema envguard.yaml --env .env --format json

# Audit: scan source code for env usage vs .env/schema (NEW in v2)
envguard audit --src ./src --env .env --schema envguard.yaml

# Sync: keep .env and .env.example in sync (NEW in v2)
envguard sync --env .env --example .env.example --schema envguard.yaml

# Watch: continuous validation on file changes (NEW in v2)
envguard watch --schema envguard.yaml --env .env

# Scan: secret security scanning
envguard scan --env .env

# Lint: schema best practices
envguard lint --schema envguard.yaml

# Generate a starter schema file
envguard init

# Print version
envguard version
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML |
| `--env` | `-e` | `.env` | Path to .env file (repeatable) |
| `--format` | `-f` | `text` | Output format: `text`, `json`, `github`, `sarif` |
| `--strict` | | `false` | Fail if .env contains keys not in schema |
| `--env-name` | | `""` | Environment name for `requiredIn`/`devOnly` rules |
| `--scan-secrets` | | `false` | Scan for hardcoded secrets in .env values |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Validation passed |
| `1` | Validation failed (missing/invalid variables) or secrets detected |
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

**These were completed in v1.0.0.**

---

## 8. v2.0.0 Feature Roadmap

### v2.0.0 Theme: "Intelligence & Integration"

| Feature | Status | Description |
|---------|--------|-------------|
| **Source Code Audit** | Planned | Scan code for `process.env.X` / `os.getenv()` usage; detect missing, unused, undocumented variables |
| **.env ↔ .env.example Sync** | Planned | Auto-generate and sync `.env.example` from `.env` + schema; detect drift |
| **Watch Mode** | Planned | Continuous validation on file changes with debounced rebuilds |
| **Variable Interpolation** | Planned | Expand `${VAR}` and `${VAR:-default}` syntax in .env values |
| **Schema Inference** | Planned | Auto-generate `envguard.yaml` from existing `.env` + code analysis |
| **Documentation Generation** | Planned | Generate Markdown docs from schema with examples and descriptions |
| **New Validation Rules** | Planned | `oneOf`/`anyOf`, `prefix`/`suffix`, cross-variable validation, `itemType` for arrays |
| **More Format Validators** | Planned | `datetime`, `timezone`, `color`, `slug`, `filepath`, `directory`, `locale` |
| **Rule Severity Levels** | Planned | Per-rule `severity: error|warn|info` instead of all-or-nothing |
| **Config File Support** | Planned | `.envguardrc.yaml` for project-wide defaults (schema path, env name, etc.) |
| **SARIF Output** | Planned | Standard SARIF format for GitHub Advanced Security / CodeQL integration |
| **Enhanced Secret Scanning** | Planned | Entropy-based detection, more built-in rules, configurable severity |
| **Schema Composition** | Planned | Multiple `extends`, conditional imports, remote schema URLs |
| **Git Hook Integration** | Planned | Built-in `envguard install-hook` for pre-commit/pre-push |
| **Monorepo Support** | Planned | Auto-detect multiple `.env` files in subdirectories; per-service validation |
| **Performance** | Planned | Parallel validation, schema caching, incremental checks |
| **IDE Ecosystem** | Planned | JetBrains plugin, enhanced VS Code extension with quick-fixes |

---

## 9. Dependencies (Go)
- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- Standard library for everything else (regex, JSON, file I/O)
