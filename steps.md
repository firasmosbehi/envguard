# EnvGuard — Implementation Steps

> Follow these steps in order. Each step is a concrete, verifiable milestone.

---

## Phase 1: Foundation (Day 1)

### Step 1.1 — Bootstrap Go Module
- [ ] Initialize `go.mod` at `github.com/envguard/envguard`
- [ ] Create directory structure (`cmd/`, `internal/`, `pkg/`, `examples/`)
- [ ] Add `Makefile` with targets: `build`, `test`, `lint`, `clean`
- [ ] Add `.gitignore` for Go binaries

### Step 1.2 — Define Core Types
- [ ] Create `internal/schema/schema.go` with `Schema`, `Variable`, `Type` structs
- [ ] Create `internal/validator/result.go` with `ValidationError`, `Result` structs
- [ ] Write unit tests for type serialization/deserialization

### Step 1.3 — Schema Parser
- [ ] Implement `schema.Parse(path string) (*Schema, error)` using `yaml.v3`
- [ ] Handle unknown fields gracefully (error or ignore)
- [ ] Add JSON Schema meta-validation for the YAML schema itself
- [ ] Write unit tests with valid and invalid YAML samples

### Step 1.4 — .env Parser
- [ ] Implement `dotenv.Parse(path string) (map[string]string, error)`
- [ ] Handle comments (`#`), quotes (`"`, `'`), and multiline values (basic)
- [ ] Handle empty lines and malformed lines gracefully
- [ ] Write unit tests

---

## Phase 2: Validation Engine (Day 2)

### Step 2.1 — Type Coercion
- [ ] `coerceString(value string) (string, error)`
- [ ] `coerceInteger(value string) (int64, error)`
- [ ] `coerceFloat(value string) (float64, error)`
- [ ] `coerceBoolean(value string) (bool, error)` — supports `true`/`1`/`yes`/`on`, `false`/`0`/`no`/`off`
- [ ] Unit tests for each coercer

### Step 2.2 — Rule Validators
- [ ] `validateRequired(variable, value)` — check presence & non-empty
- [ ] `validateEnum(variable, coercedValue)` — check allowed values
- [ ] `validatePattern(variable, stringValue)` — regex match
- [ ] `validateDefault(variable)` — inject default if missing
- [ ] Unit tests for each rule

### Step 2.3 — Validation Orchestrator
- [ ] Implement `validator.Validate(schema, envVars, strict bool) *Result`
- [ ] Iterate all schema variables; collect errors; never short-circuit early
- [ ] If `strict`, detect unknown keys in `.env`
- [ ] Return `Result{Valid bool, Errors []ValidationError, Warnings []ValidationError}`
- [ ] Comprehensive integration tests (happy path + all error types)

---

## Phase 3: CLI (Day 3)

### Step 3.1 — Cobra Setup
- [ ] Install `cobra` CLI dependency
- [ ] Wire up `cmd/envguard/main.go` with root command

### Step 3.2 — Commands
- [ ] `envguard validate` with `--schema`, `--env`, `--format`, `--strict` flags
- [ ] `envguard init` — write a sample `envguard.yaml` to cwd
- [ ] `envguard version` — print version string

### Step 3.3 — Reporters
- [ ] `reporter.Text(result *Result)` — colored human-readable output
- [ ] `reporter.JSON(result *Result)` — machine-readable JSON output
- [ ] Ensure text output uses clear symbols (✓, ✗, ⚠) and indentation

### Step 3.4 — Exit Codes
- [ ] Exit 0 on success
- [ ] Exit 1 on validation failure
- [ ] Exit 2 on file/parse errors
- [ ] E2E test: run CLI against sample files and assert exit codes

---

## Phase 4: Polish & Distribution (Day 4)

### Step 4.1 — Examples
- [ ] Create `examples/envguard.yaml` with all field types demonstrated
- [ ] Create `examples/.env` (valid) and `examples/.env.invalid` (for testing)
- [ ] Add `examples/README.md` showing how to run

### Step 4.2 — Tests & CI
- [ ] Achieve ≥80% test coverage
- [ ] Add GitHub Actions workflow: `go test`, `go vet`, `go build` on PRs
- [ ] Add GitHub Actions workflow: build releases for `linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`

### Step 4.3 — Documentation
- [ ] Write `README.md` with installation, usage, and schema reference
- [ ] Write `CHANGELOG.md` for v0.1.0
- [ ] Add `--help` text to all CLI commands

### Step 4.4 — Release
- [ ] Tag `v0.1.0`
- [ ] Attach compiled binaries to GitHub Release
- [ ] Publish installation instructions (Homebrew formula optional)

---

## Phase 5: Language Packages (Post-MVP)

### Step 5.1 — Node.js Package (`@envguard/node`)
- [ ] TypeScript wrapper that downloads the correct Go binary for the platform
- [ ] Expose `validate(schemaPath, envPath, options)` returning parsed `Result`
- [ ] Publish to npm

### Step 5.2 — Python Package (`envguard`)
- [ ] Python wrapper using `subprocess` to call the CLI
- [ ] Expose `validate(schema_path, env_path, strict=False)`
- [ ] Publish to PyPI

### Step 5.3 — Java Package (`envguard-java`)
- [ ] Maven/Gradle project with CLI invocation wrapper
- [ ] Expose `EnvGuard.validate(...)` returning `ValidationResult`
- [ ] Publish to Maven Central

---

## Appendix: Quick Commands

```bash
# Dev loop
make build && ./bin/envguard validate -s examples/envguard.yaml -e examples/.env

# Run all tests
make test

# Cross-compile
make build-all
```
