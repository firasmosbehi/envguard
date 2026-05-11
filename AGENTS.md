# AGENTS.md — EnvGuard

> Agent-focused guidance for the EnvGuard project. Read this before modifying code.

---

## 1. What is EnvGuard?

EnvGuard is a **language-agnostic CLI tool** written in Go that validates `.env` files against a declarative YAML schema. It catches missing, mistyped, or malformed environment variables before deployment. The Go CLI is the universal core; wrapper packages for Node.js and Python spawn the CLI and parse JSON output. A native GitHub Action, Docker image, Homebrew formula, and pre-commit hook are also provided.

**Motto:** Define once in YAML. Validate everywhere.

---

## 2. Tech Stack

- **Language:** Go 1.26.2 (module version in `go.mod`; CI currently uses Go 1.22)
- **CLI Framework:** `github.com/spf13/cobra`
- **YAML Parser:** `gopkg.in/yaml.v3`
- **Testing:** Standard `testing` package (no external test dependencies in `go.mod`)
- **Linting:** `golangci-lint` (target: zero warnings)
- **Wrappers:**
  - Node.js: TypeScript, compiled to `dist/`, published as `envguard-validator` on npm
  - Python: pure Python, published as `envguard-validator` on PyPI
- **Container:** Alpine 3.19 base image, binaries downloaded from GitHub releases
- **Action:** GitHub composite action (`action.yml`)

---

## 3. Directory Structure

```
envguard/
├── cmd/envguard/              # CLI entrypoint only
│   └── main.go                # Defines version constant; wires cli.Execute()
├── internal/                  # Private implementation
│   ├── cli/                   # Cobra command wiring
│   │   ├── root.go            # Root command & Execute()
│   │   ├── validate.go        # validate command (core user flow)
│   │   ├── init.go            # init command (generate starter schema)
│   │   ├── generate.go        # generate-example command (create .env.example)
│   │   ├── version.go         # version command
│   │   ├── errors.go          # Sentinel errors (ErrValidationFailed, ErrIO)
│   │   └── cli_test.go        # Unit tests for CLI logic
│   ├── schema/                # YAML schema parsing & structural validation
│   │   ├── schema.go          # Schema, Variable types; Parse(); Validate()
│   │   └── *_test.go          # Unit tests for schema parsing
│   ├── dotenv/                # .env file parser
│   │   ├── dotenv.go          # Parse(); handles comments, quotes, escapes
│   │   └── *_test.go          # Unit tests
│   ├── validator/             # Validation engine
│   │   ├── validator.go       # Validate() orchestration; per-type validators
│   │   ├── coerce.go          # Type coercion (string, int, float, bool, array)
│   │   ├── result.go          # Result, ValidationError, Warning types
│   │   └── *_test.go          # Extensive unit tests
│   └── reporter/              # Output formatters
│       ├── text.go            # Human-readable text output
│       ├── json.go            # Machine-readable JSON output
│       └── *_test.go          # Unit tests
├── pkg/envguard/              # PUBLIC API directory (currently empty)
├── e2e/                       # End-to-end tests
│   ├── e2e_test.go            # Core e2e scenarios
│   ├── e2e_more_features_test.go
│   └── e2e_new_features_test.go
├── packages/                  # Language wrappers
│   ├── node/                  # npm package `envguard-validator`
│   │   ├── src/
│   │   │   ├── index.ts       # Public exports
│   │   │   ├── validator.ts   # validate() / validateSync()
│   │   │   ├── types.ts       # TypeScript interfaces
│   │   │   ├── install.ts     # Post-install binary downloader
│   │   │   └── cli.ts         # npx CLI wrapper
│   │   ├── package.json
│   │   └── tsconfig.json
│   └── python/                # PyPI package `envguard-validator`
│       ├── envguard/
│       │   ├── __init__.py
│       │   ├── validator.py   # validate()
│       │   ├── cli.py         # envguard-py CLI
│       │   └── install.py     # Lazy binary downloader
│       └── pyproject.toml
├── .github/workflows/         # CI/CD pipelines
│   ├── ci.yml                 # Build, test, vet, e2e on PR/push to main
│   ├── test-action.yml        # Test GitHub Action on ubuntu-latest + macos-latest
│   ├── release.yml            # Cross-compile binaries & create GitHub Release on tag
│   ├── publish-npm.yml        # Publish Node.js wrapper on tag
│   ├── publish-pypi.yml       # Publish Python wrapper on tag
│   └── docker.yml             # Build & push multi-arch Docker image to GHCR on tag
├── action.yml                 # GitHub Action composite definition
├── Dockerfile                 # Alpine-based image downloading release binary
├── homebrew/
│   └── envguard.rb            # Homebrew formula (downloads release binaries)
├── .pre-commit-hooks.yaml     # pre-commit hook definition
├── examples/                  # Sample schema and .env files for manual testing
│   ├── envguard.yaml
│   ├── .env
│   └── .env.invalid
├── schemas/                   # JSON Schema for YAML meta-validation (if any)
├── testdata/                  # Test fixture files
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── AGENTS.md                  # This file
```

**Rule of thumb:**
- Put CLI-specific code in `cmd/` and `internal/cli/`
- Put reusable business logic in `internal/<domain>/`
- `pkg/envguard/` is reserved for a future public Go API

---

## 4. Coding Conventions

### Go Style
- Follow **Effective Go** and **Go Code Review Comments**
- Use `gofmt` / `goimports` on every save
- Prefer **explicit error handling** over panics
- Exported functions must have doc comments starting with the function name
- Keep functions small and focused (max ~40 lines when possible)
- Prefer `errors.New` / `fmt.Errorf` over custom error types unless necessary

### Naming
- Packages: short, lowercase, no underscores (`schema`, `validator`, `reporter`)
- Files: `snake_case.go` for implementation, `*_test.go` for tests
- Structs: PascalCase, descriptive (`ValidationResult`, `EnvVariable`)
- Interfaces: `-er` suffix when natural (`Parser`, `Reporter`, `Validator`)

### Error Messages
Error messages shown to the user (via CLI) must be:
- **Clear:** say what failed and why
- **Actionable:** suggest how to fix it when possible
- **Concise:** no stack traces in user-facing output

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
| `format` | `string` | Built-in format: `email`, `url`, `uuid` |
| `disallow` | `string` | Array of forbidden string values |
| `requiredIn` | all | Environments where the variable is required |
| `devOnly` | all | Variable only allowed in development; skipped otherwise |
| `separator` | `array` | Delimiter for splitting array items (default `,`) |
| `allowEmpty` | all | If `false`, reject empty strings even when optional |
| `contains` | `array` | Require array to contain this specific item |
| `dependsOn` | all | Name of another variable that triggers conditional requirement |
| `when` | all | Value the `dependsOn` variable must have to trigger requirement |

### Constraints
- `required: true` and `default` are mutually exclusive in practice
- `enum` values must be compatible with the variable's `type`
- `pattern` is only applied to `string` types
- Empty enums (`enum: []`) are rejected as invalid schema definitions
- Whitespace-only values (e.g., `"   "`) fail `required` checks
- `devOnly: true` and `required` / `requiredIn` are mutually exclusive
- `dependsOn` and `when` must be used together
- `allowEmpty: false` is redundant when `required: true`
- `min` cannot be greater than `max`; `minLength` cannot be greater than `maxLength`

### Type Coercion Rules

| Type | Accepted Input | Rejected Input |
|------|---------------|----------------|
| `string` | any text | — |
| `integer` | `42`, `-3`, `0` | `3.14`, `abc`, `12.0` |
| `float` | `3.14`, `-2.5`, `10`, `1.5e10` | `abc` |
| `boolean` | `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off` (case-insensitive) | `2`, `maybe`, empty string |
| `array` | `"a,b,c"` | `""` (empty string) |

### Validation Order
1. Check `devOnly` / `requiredIn` / `dependsOn` to determine requiredness
2. Check `required` (presence + non-empty after trim)
3. Check `allowEmpty`
4. Apply `default` if missing
5. Coerce to `type`
6. Check `enum`, `pattern`, `min`/`max`, `minLength`/`maxLength`, `format`, `disallow`, `contains`

**Never short-circuit.** Collect ALL errors before returning.

---

## 6. CLI Behavior

### Commands

| Command | Purpose |
|---------|---------|
| `envguard validate [flags]` | Validate `.env` against schema |
| `envguard init` | Generate a starter `envguard.yaml` |
| `envguard generate-example` | Generate `.env.example` from schema |
| `envguard version` | Print version |

### `validate` Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--env` | `-e` | `.env` | Path to `.env` file (repeatable for multiple files) |
| `--format` | `-f` | `text` | Output format: `text` or `json` |
| `--strict` | | `false` | Fail if `.env` contains keys not defined in schema |
| `--env-name` | | `""` | Environment name for `requiredIn`/`devOnly` rules |

Multiple `--env` files are merged **right-to-left** (later files override earlier ones).

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Validation passed |
| `1` | Validation failed (missing/invalid variables) |
| `2` | I/O or schema parsing error |

**Do not change exit codes** — they are part of the public contract for CI pipelines and wrappers.

---

## 7. Testing Rules

- Every package in `internal/` must have corresponding `*_test.go` files
- Target **≥80% code coverage** for the validator and parser packages
- Use table-driven tests for validation rules
- Keep test data in `testdata/` subdirectories when files are needed
- E2E tests live in `e2e/` and run the compiled binary against temporary files, asserting exit codes and output

### Running tests

```bash
# Go unit tests + race detector + coverage
make test

# E2E tests
go test -v ./e2e/...

# Build all platform binaries
make build-all
```

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

## 8. Build & Dev Commands

```bash
# Build the CLI binary
make build
# Output: bin/envguard

# Run all tests with coverage
make test

# Run linter
make lint

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

## 9. Wrappers & Distribution

### Node.js (`packages/node/`)
- Exports `validate()` (async) and `validateSync()`
- CLI via `npx envguard-validator validate ...`
- `postinstall` script downloads the correct platform binary from GitHub releases
- Published as `envguard-validator` on npm

### Python (`packages/python/`)
- Exports `validate()`
- CLI via `envguard-py validate ...`
- Lazy-downloads the Go binary to `~/.envguard/bin/` on first use
- Published as `envguard-validator` on PyPI

### GitHub Action (`action.yml`)
- Composite action that detects the runner OS/arch, downloads the matching release binary, and runs `envguard validate`
- Inputs: `schema`, `env`, `strict`, `format`, `env-name`, `version`

### Docker (`Dockerfile`)
- Alpine 3.19 base; downloads the release binary at build time
- Published to `ghcr.io/firasmosbehi/envguard:latest` and version tags
- Default command: `envguard validate`

### Homebrew (`homebrew/envguard.rb`)
- Formula downloads platform-specific release binaries
- Installs as `envguard`

### Pre-commit (`.pre-commit-hooks.yaml`)
- Hook ID: `envguard-validate`
- Runs `envguard validate` on `.env` files before each commit

---

## 10. CI/CD Pipelines

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | push/PR to `main` | Build, unit tests, `go vet`, e2e tests |
| `test-action.yml` | push/PR to `main` | Test the GitHub Action on Ubuntu and macOS |
| `release.yml` | tag `v*` | Cross-compile binaries, create GitHub Release |
| `publish-npm.yml` | tag `v*` | Build and publish Node.js wrapper to npm |
| `publish-pypi.yml` | tag `v*` | Build and publish Python wrapper to PyPI |
| `docker.yml` | tag `v*` or manual | Build and push multi-arch image to GHCR |

All wrapper publishing workflows require repository secrets (`NPM_TOKEN`, `PYPI_API_TOKEN`).

---

## 11. Design Principles

1. **Fail fast, but report everything.** Don't stop at the first error; collect all validation failures so the user can fix them in one pass.
2. **No magic.** The schema is explicit YAML. No inference from `.env.example`, no guessing types.
3. **CLI is the source of truth.** Language packages wrap the CLI and share the same schema format. Don't add language-specific schema extensions.
4. **Zero runtime dependencies for users.** The CLI is a single static binary. Users don't need Go, Node, Python, or anything else installed.
5. **CI-first JSON output.** The `--format json` output must be stable and machine-parseable; treat it as a public API.

---

## 12. Versioning & Releases

- Follow **SemVer**: `vMAJOR.MINOR.PATCH`
- Current version: `0.1.7`
- The version constant is hard-coded in `cmd/envguard/main.go`
- Wrapper versions (`packages/node/package.json`, `packages/python/pyproject.toml`, `homebrew/envguard.rb`, `Dockerfile`, `action.yml`) must be kept in sync
- Tag releases on GitHub; the following artifacts are produced automatically:
  - `linux/amd64`, `linux/arm64`
  - `darwin/amd64`, `darwin/arm64`
  - `windows/amd64`
- Update `CHANGELOG.md` with every release

### Release checklist
1. Bump version in `cmd/envguard/main.go`
2. Bump version in `packages/node/package.json`
3. Bump version in `packages/python/pyproject.toml` and `packages/python/envguard/__init__.py`
4. Bump version in `homebrew/envguard.rb`
5. Bump version in `Dockerfile`
6. Bump version in `action.yml`
7. Update `CHANGELOG.md`
8. Commit and push to `main`
9. Create and push a tag:
   ```bash
   git tag v0.1.8
   git push origin v0.1.8
   ```
10. GitHub Actions automatically build and publish all artifacts

---

## 13. When to Update This File

Update `AGENTS.md` when you:
- Add a new CLI command or flag
- Change the schema format or add/remove validation rules
- Modify exit codes or JSON output structure
- Add/remove a top-level directory
- Change build tools or Go version requirements
- Add or change CI/CD workflows, wrappers, or distribution channels
