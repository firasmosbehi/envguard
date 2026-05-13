<!-- AGENTS.md — EnvGuard -->
# EnvGuard — Agent-Focused Project Guide

> Read this file before modifying any code. It describes the project's architecture, conventions, and processes as they actually exist in the source tree.

---

## 1. Project Overview

EnvGuard is a **language-agnostic CLI tool** written in Go that validates `.env` files against a declarative YAML schema. It catches missing, mistyped, or malformed environment variables before deployment. The Go CLI is the universal core; wrapper packages for Node.js and Python spawn the CLI and parse JSON output. A native GitHub Action, Docker image, Homebrew formula, VS Code extension, and pre-commit hook are also provided.

**Motto:** Define once in YAML. Validate everywhere.

---

## 2. Technology Stack

| Layer | Technology | Version / Notes |
|-------|-----------|-----------------|
| **Core language** | Go | 1.26.2 (`go.mod`); CI workflows currently use Go 1.22 |
| **CLI framework** | `github.com/spf13/cobra` | v1.10.2 |
| **YAML parser** | `gopkg.in/yaml.v3` | v3.0.1 |
| **Testing** | Standard `testing` package | No external test dependencies |
| **Linting** | `golangci-lint` (Go), `ESLint` (TypeScript), `Ruff` (Python) | Target: zero warnings |
| **Node.js wrapper** | TypeScript | Node ≥ 16, TypeScript ~5.4 |
| **Python wrapper** | Pure Python | Python ≥ 3.8 |
| **VS Code extension** | TypeScript | VS Code ^1.74.0, depends on `yaml` ^2.3.0 |
| **Container** | Multi-stage Docker | `golang:1.26-alpine` → `scratch` |

---

## 3. Directory Structure

```
envguard/
├── cmd/envguard/
│   └── main.go                    # Entrypoint only; hard-codes version constant
├── internal/                      # Private implementation
│   ├── cli/
│   │   ├── root.go                # Cobra root command & Execute()
│   │   ├── validate.go            # validate command (core user flow)
│   │   ├── scan.go                # scan command (secret detection)
│   │   ├── lint.go                # lint command (schema best practices)
│   │   ├── init.go                # init command (generate starter schema)
│   │   ├── generate.go            # generate-example command (create .env.example)
│   │   ├── version.go             # version command
│   │   ├── errors.go              # Sentinel errors (ErrValidationFailed, ErrIO)
│   │   ├── cli_test.go            # Unit tests for CLI logic
│   │   └── lint_test.go           # Tests for lint command
│   ├── schema/
│   │   ├── schema.go              # Schema, Variable types; Parse(); Validate()
│   │   └── *_test.go              # Schema parsing & structural validation tests
│   ├── dotenv/
│   │   ├── dotenv.go              # Parse(); handles comments, quotes, escapes
│   │   └── *_test.go              # Parser unit tests
│   ├── validator/
│   │   ├── validator.go           # Validate() orchestration; per-type validators
│   │   ├── coerce.go              # Type coercion (string, int, float, bool, array)
│   │   ├── result.go              # Result, ValidationError, Warning types
│   │   └── *_test.go              # Extensive unit tests
│   ├── reporter/
│   │   ├── text.go                # Human-readable text output
│   │   ├── json.go                # Machine-readable JSON output
│   │   ├── github.go              # GitHub Actions workflow command output
│   │   └── *_test.go              # Reporter tests
│   └── secrets/
│       ├── secrets.go             # Hardcoded-credential scanner (8 built-in rules)
│       └── secrets_test.go        # Scanner tests
├── pkg/envguard/
│   ├── envguard.go                # PUBLIC Go API (Validate, ValidateFile, ParseSchema, ParseEnv)
│   └── envguard_test.go           # Public API tests
├── e2e/
│   ├── e2e_test.go                          # Core e2e scenarios
│   ├── e2e_commands_and_validators_test.go  # Command + format validator e2e tests
│   ├── e2e_more_features_test.go
│   └── e2e_new_features_test.go
├── packages/
│   ├── node/                      # npm package `envguard-validator`
│   │   ├── src/
│   │   │   ├── index.ts           # Public exports
│   │   │   ├── validator.ts       # validate() / validateSync()
│   │   │   ├── types.ts           # TypeScript interfaces
│   │   │   ├── install.ts         # Post-install binary downloader (hardcodes VERSION)
│   │   │   └── cli.ts             # npx CLI wrapper
│   │   ├── package.json
│   │   └── tsconfig.json
│   └── python/                    # PyPI package `envguard-validator`
│       ├── envguard/
│       │   ├── __init__.py
│       │   ├── validator.py       # validate()
│       │   ├── cli.py             # envguard-py CLI
│       │   └── install.py         # Lazy binary downloader (hardcodes VERSION)
│       └── pyproject.toml
├── vscode-extension/
│   ├── src/extension.ts           # Real-time .env validation in VS Code
│   ├── package.json
│   └── tsconfig.json
├── .github/workflows/             # CI/CD pipelines (see §9)
├── action.yml                     # GitHub Action composite definition
├── Dockerfile                     # Multi-stage build
├── homebrew/
│   └── envguard.rb                # Homebrew formula (downloads release binaries)
├── .pre-commit-hooks.yaml         # pre-commit hook definition
├── examples/                      # Sample schema and .env files for manual testing
│   ├── envguard.yaml
│   ├── .env
│   └── .env.invalid
├── testdata/                      # Test fixture directory (currently empty)
├── schemas/                       # JSON Schema directory (currently empty)
├── Makefile
├── go.mod / go.sum
├── README.md
└── AGENTS.md                      # This file
```

**Rule of thumb:**
- Put CLI-specific code in `cmd/` and `internal/cli/`.
- Put reusable business logic in `internal/<domain>/`.
- `pkg/envguard/` is the public Go API.

---

## 4. Coding Conventions

### Go Style
- Follow **Effective Go** and **Go Code Review Comments**.
- Use `gofmt` / `goimports` on every save.
- Prefer **explicit error handling** over panics.
- Exported functions must have doc comments starting with the function name.
- Keep functions small and focused (max ~40 lines when possible).
- Prefer `errors.New` / `fmt.Errorf` over custom error types unless necessary.

### Naming
- Packages: short, lowercase, no underscores (`schema`, `validator`, `reporter`).
- Files: `snake_case.go` for implementation, `*_test.go` for tests.
- Structs: PascalCase, descriptive (`ValidationResult`, `EnvVariable`).
- Interfaces: `-er` suffix when natural (`Parser`, `Reporter`, `Validator`).

### Error Messages
Error messages shown to the user (via CLI) must be:
- **Clear:** say what failed and why.
- **Actionable:** suggest how to fix it when possible.
- **Concise:** no stack traces in user-facing output.

Internal errors (I/O failures, YAML syntax errors) should include context:
```go
fmt.Errorf("failed to parse schema file %s: %w", path, err)
```

---

## 5. Schema Format Reference

EnvGuard schemas are YAML files named `envguard.yaml` by default.

### Top-level structure

```yaml
version: "1.0"           # Schema version (required)
extends: "base.yaml"     # Optional: inherit from another schema file
env:                     # Map of variable names to definitions (required)
  VARIABLE_NAME:
    type: string
    required: true
    default: "fallback"
    description: "Human-readable docs"
    pattern: "^regex$"
    enum: [a, b, c]
```

### Supported types
- `string`
- `integer`
- `float`
- `boolean`
- `array`

### Supported rules / fields

| Field | Types | Description |
|-------|-------|-------------|
| `type` | all | **Required.** Data type of the variable |
| `required` | all | If `true`, variable must be present and non-empty |
| `default` | all | Fallback value injected when variable is absent |
| `description` | all | Human-readable docs, shown in errors |
| `message` | all | Custom error message on any validation failure |
| `pattern` | `string` | Regex the value must match |
| `enum` | `string`, `integer`, `float`, `array` | Array of allowed values |
| `min` | `integer`, `float` | Minimum numeric value (inclusive) |
| `max` | `integer`, `float` | Maximum numeric value (inclusive) |
| `minLength` | `string`, `array` | Minimum length (chars for string, items for array) |
| `maxLength` | `string`, `array` | Maximum length |
| `format` | `string` | Built-in format: `email`, `url`, `uuid`, `base64`, `ip`, `port`, `json`, `duration`, `semver`, `hostname`, `hex`, `cron` |
| `disallow` | `string` | Array of forbidden string values |
| `requiredIn` | all | Environments where the variable is required |
| `devOnly` | all | Variable only allowed in development; skipped otherwise |
| `separator` | `array` | Delimiter for splitting array items (required for `array`) |
| `allowEmpty` | all | If `false`, reject empty strings even when optional |
| `contains` | `array` | Require array to contain this specific item |
| `dependsOn` | all | Name of another variable that triggers conditional requirement |
| `when` | all | Value the `dependsOn` variable must have to trigger requirement |
| `deprecated` | all | Warning message shown when variable is present (suggest replacement) |
| `sensitive` | all | If `true`, redact value in error/output messages |
| `transform` | `string` | Pre-validation transform: `lowercase`, `uppercase`, `trim` |

### Constraints
- `required: true` and `default` are mutually exclusive in practice.
- `enum` values must be compatible with the variable's `type`.
- `pattern` is only applied to `string` types.
- Empty enums (`enum: []`) are rejected as invalid schema definitions.
- Whitespace-only values (e.g., `"   "`) fail `required` checks.
- `devOnly: true` and `required` / `requiredIn` are mutually exclusive.
- `dependsOn` and `when` must be used together.
- `allowEmpty: false` is redundant when `required: true`.
- `min` cannot be greater than `max`; `minLength` cannot be greater than `maxLength`.
- `array` type **requires** a `separator`.
- Circular `extends` inheritance is detected and rejected.
- `transform` can only be used with `string` type.
- `sensitive` has no effect on validation logic, only on output redaction.

### Type Coercion Rules

| Type | Accepted Input | Rejected Input |
|------|---------------|----------------|
| `string` | any text | — |
| `integer` | `42`, `-3`, `0` | `3.14`, `abc`, `12.0` |
| `float` | `3.14`, `-2.5`, `10`, `1.5e10` | `abc` |
| `boolean` | `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off` (case-insensitive) | `2`, `maybe`, empty string |
| `array` | `"a,b,c"` | `""` (empty string) |

### Validation Order
1. Check `devOnly` / `requiredIn` / `dependsOn` to determine requiredness.
2. Warn if `deprecated` and variable is present.
3. Check `required` (presence + non-empty after trim).
4. Check `allowEmpty`.
5. Apply `default` if missing.
6. Apply `transform` if specified.
7. Coerce to `type`.
8. Check `enum`, `pattern`, `min`/`max`, `minLength`/`maxLength`, `format`, `disallow`, `contains`.

**Never short-circuit.** Collect ALL errors before returning.

---

## 6. CLI Behavior

### Commands

| Command | Purpose |
|---------|---------|
| `envguard validate [flags]` | Validate `.env` against schema |
| `envguard scan [flags]` | Scan `.env` for hardcoded secrets |
| `envguard lint [flags]` | Lint schema file for best practices |
| `envguard init` | Generate a starter `envguard.yaml` |
| `envguard generate-example` | Generate `.env.example` from schema |
| `envguard version` | Print version |

### `validate` Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--env` | `-e` | `.env` | Path to `.env` file (repeatable for multiple files) |
| `--format` | `-f` | `text` | Output format: `text`, `json`, or `github` |
| `--strict` | | `false` | Fail if `.env` contains keys not defined in schema |
| `--env-name` | | `""` | Environment name for `requiredIn`/`devOnly` rules |
| `--scan-secrets` | | `false` | Scan for hardcoded secrets in `.env` values |

Multiple `--env` files are merged **right-to-left** (later files override earlier ones).

### `scan` Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--env` | `-e` | `.env` | Path to `.env` file (repeatable) |
| `--format` | `-f` | `text` | Output format: `text` or `json` |
| `--schema` | `-s` | `""` | Optional schema file with custom secret rules |

### `lint` Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--format` | `-f` | `text` | Output format: `text` or `json` |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Validation passed / no secrets found |
| `1` | Validation failed (missing/invalid variables) or secrets detected |
| `2` | I/O or schema parsing error |

**Do not change exit codes** — they are part of the public contract for CI pipelines and wrappers.

---

## 7. Secrets Detection (`internal/secrets`)

The `scan` command and the `--scan-secrets` flag use `internal/secrets.DefaultScanner()`, which includes 8 built-in regex rules:

| Rule | Pattern |
|------|---------|
| `aws-access-key` | `AKIA[0-9A-Z]{16}` |
| `aws-secret-key` | `^[A-Za-z0-9/+=]{40}$` |
| `github-token` | `gh[pousr]_[A-Za-z0-9_]{36,}` |
| `private-key` | `-----BEGIN (RSA \|EC \|DSA \|OPENSSH )?PRIVATE KEY-----` |
| `generic-api-key` | `(?i)(api[_-]?key\|apikey)\s*[:=]\s*['"]?([a-z0-9_\-]{16,})['"]?` |
| `slack-token` | `xox[baprs]-[0-9]{10,13}-[0-9]{10,13}(-[a-zA-Z0-9]{24})?` |
| `stripe-key` | `sk_(live\|test)_[0-9a-zA-Z_]{24,}` |
| `jwt-token` | `eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*` |

Each match reports the key name, rule name, message, and a redacted snippet. Only the first match per rule per variable is reported.

### Custom Secret Rules

Users can define custom secret detection rules in `envguard.yaml`:

```yaml
version: "1.0"
env:
  # ...
secrets:
  custom:
    - name: "internal-api-token"
      pattern: "iat_[a-zA-Z0-9]{32}"
      message: "Internal API token detected"
```

Custom rules are loaded by `envguard scan --schema` and `envguard validate --scan-secrets`.

---

## 8. Testing Rules

- Every package in `internal/` must have corresponding `*_test.go` files.
- Target **≥80% code coverage** for the validator and parser packages.
- Use table-driven tests for validation rules.
- Keep test data in `testdata/` subdirectories when files are needed (currently unused).
- E2E tests live in `e2e/` and run the compiled binary against temporary files, asserting exit codes and output.

### Running tests

```bash
# Go unit tests + race detector + coverage report
make test

# E2E tests (builds the binary internally)
go test -v ./e2e/...

# Build all platform binaries
make build-all
```

**Note:** `make test` runs `go test -v -race -coverprofile=coverage.out ./...` followed by `go tool cover -func=coverage.out`. The CI workflow (`ci.yml`) runs `go test -v -race ./...` without the coverage report step.

Example test pattern:
```go
func TestCoerceBoolean(t *testing.T) {
    tests := []struct {
        input    string
        expected bool
        wantErr  bool
    }{
        {"true", true, false},
        {"FALSE", false, false},
        {"yes", true, false},
        {"2", false, true},
    }
    // ... iterate and assert
}
```

---

## 9. Build & Dev Commands

```bash
# Build the CLI binary
make build
# Output: bin/envguard

# Run all tests with coverage
make test

# Run all linters (Go + TypeScript + Python)
make lint

# Run individual linters
make lint-go       # golangci-lint
make lint-ts       # ESLint
make lint-py       # Ruff check + format check

# Auto-fix lint issues
make lint-fix

# Clean build artifacts
make clean

# Cross-compile for all platforms
make build-all
# Outputs:
#   bin/envguard-linux-amd64
#   bin/envguard-darwin-amd64
#   bin/envguard-darwin-arm64
#   bin/envguard-windows-amd64.exe

# Quick manual validation during dev
make build && ./bin/envguard validate -s examples/envguard.yaml -e examples/.env
```

---

## 10. Wrappers & Distribution

### Design principle
**The Go CLI is the single source of truth.** All wrappers spawn the binary and parse its JSON output.

### Node.js (`packages/node/`)
- **Package:** `envguard-validator` on npm
- **Exports:** `validate()` (async Promise) and `validateSync()`
- **CLI:** `npx envguard-validator validate ...`
- **Binary delivery:** `postinstall` script downloads the correct platform binary from GitHub releases
- **Build:** `tsc` compiles `src/` → `dist/`

### Python (`packages/python/`)
- **Package:** `envguard-validator` on PyPI
- **Exports:** `validate()` function returning `ValidationResult` dataclass
- **CLI:** `envguard-py validate ...`
- **Binary delivery:** `install.py` lazy-downloads the Go binary to `~/.envguard/bin/` on first use
- **Build:** `python -m build` (setuptools backend)

### VS Code Extension (`vscode-extension/`)
- **Package:** `envguard-vscode` (publisher: `firasmosbehi`)
- **Activation:** on `.env` file presence or `envguard.yaml` in workspace
- **Behavior:** watches `.env` files and schema file; runs `envguard validate --format json`; displays diagnostics
- **Config:** `envguard.schemaPath` (default `envguard.yaml`), `envguard.enableValidation` (default `true`)
- **Binary discovery:** checks `PATH` for `envguard`, then `~/.envguard/bin/envguard`, then `/usr/local/bin/envguard`, `/usr/bin/envguard`

### GitHub Action (`action.yml`)
- Composite action that detects the runner OS/arch, downloads the matching release binary, and runs `envguard validate`.
- Inputs: `schema`, `env`, `strict`, `format`, `env-name`, `version`.
- Download retry logic: up to 5 attempts with 10-second delays.

### Docker (`Dockerfile`)
- Multi-stage build: `golang:1.26-alpine` → `scratch`.
- Copies CA certificates and static binary.
- Published to `ghcr.io/firasmosbehi/envguard:latest` and version tags.
- Default command: `envguard validate`.

### Homebrew (`homebrew/envguard.rb`)
- Formula downloads platform-specific release binaries.
- Installs as `envguard`.
- **Note:** The formula references `linux-arm64`, but the release matrix (`release.yml`) currently only builds `linux-amd64`, `darwin-amd64`, `darwin-arm64`, and `windows-amd64`.

### Pre-commit (`.pre-commit-hooks.yaml`)
- Hook ID: `envguard-validate`
- Runs `envguard validate --strict` on `.env` files.
- **Important:** Uses `pass_filenames: false` and `always_run: true`, so the hook does not receive filenames and always runs `envguard validate --strict` against the default schema and env paths.

---

## 11. CI/CD Pipelines

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | push/PR to `main` | Build, Go tests + lint, Node.js tests + lint, Python tests + lint, E2E validation |
| `test-action.yml` | push/PR to `main` | Tests GitHub Action on `ubuntu-latest` + `macos-latest` |
| `release.yml` | tag `v*` | Matrix build for 4 platforms → upload artifacts → create GitHub Release |
| `publish-npm.yml` | tag `v*` | Build and publish Node.js wrapper to npm |
| `publish-pypi.yml` | tag `v*` | Build and publish Python wrapper to PyPI |
| `docker.yml` | push to `main`, tag `v*`, manual | Multi-arch (`linux/amd64`, `linux/arm64`) build & push to GHCR |

### Release matrix (`release.yml`)
- `linux/amd64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`

### Secrets required
- `NPM_TOKEN` (npm publish)
- `PYPI_API_TOKEN` (PyPI publish)
- `GITHUB_TOKEN` (Docker GHCR login)

---

## 12. Design Principles

1. **Fail fast, but report everything.** Don't stop at the first error; collect all validation failures so the user can fix them in one pass.
2. **No magic.** The schema is explicit YAML. No inference from `.env.example`, no guessing types.
3. **CLI is the source of truth.** Language packages wrap the CLI and share the same schema format. Don't add language-specific schema extensions.
4. **Zero runtime dependencies for users.** The CLI is a single static binary. Users don't need Go, Node, Python, or anything else installed.
5. **CI-first JSON output.** The `--format json` output must be stable and machine-parseable; treat it as a public API.

---

## 13. Versioning & Releases

- Follow **SemVer**: `vMAJOR.MINOR.PATCH`
- Current version: `1.0.0`
- The version constant is hard-coded in `cmd/envguard/main.go`.
- Wrapper versions must be kept in sync across all files that hardcode it.
- Tag releases on GitHub; artifacts are produced automatically.

### Release checklist
1. Bump version in `cmd/envguard/main.go`.
2. Bump version in `packages/node/package.json`.
3. Bump version in `packages/node/src/install.ts`.
4. Bump version in `packages/python/pyproject.toml`.
5. Bump version in `packages/python/envguard/__init__.py`.
6. Bump version in `packages/python/envguard/install.py`.
7. Bump version in `vscode-extension/package.json`.
8. Bump version in `homebrew/envguard.rb`.
9. Bump version in `Dockerfile`.
10. Bump version in `action.yml`.
11. Update `CHANGELOG.md`.
12. Commit and push to `main`.
13. Create and push a tag:
    ```bash
    git tag v1.0.0
    git push origin v1.0.0
    ```
14. GitHub Actions automatically build and publish all artifacts.

---

## 14. Security Considerations

### Secret Scanning
- The `scan` command and `--scan-secrets` flag detect 8 built-in secret patterns (AWS keys, GitHub tokens, private keys, Stripe/Slack tokens, JWTs, generic API keys).
- All detected secrets are **redacted** in output using rule-specific redaction functions.
- Custom secret rules can be defined in `envguard.yaml` under `secrets.custom`.

### Sensitive Value Redaction
- Schema variables marked with `sensitive: true` have their values replaced with `***` in validation error and warning messages via `Result.RedactSensitive()`.
- This prevents accidental leakage of credentials in CI logs or wrapper output.

### Binary Distribution
- Wrappers download platform-specific binaries from GitHub releases over HTTPS.
- Node.js wrapper stores the binary inside the package's `dist/` folder.
- Python wrapper stores the binary in `~/.envguard/bin/`.
- The Docker image is built from `scratch` with only the static binary and CA certificates, minimizing attack surface.

### Pre-commit Hook Behavior
- The pre-commit hook (`envguard-validate`) runs `envguard validate` without passing filenames (`pass_filenames: false`).
- This means it validates the default `.env` against `envguard.yaml` in the repo root, not individual staged files.

---

## 15. When to Update This File

Update `AGENTS.md` when you:
- Add a new CLI command or flag.
- Change the schema format or add/remove validation rules.
- Modify exit codes or JSON output structure.
- Add/remove a top-level directory.
- Change build tools or Go version requirements.
- Add or change CI/CD workflows, wrappers, or distribution channels.
