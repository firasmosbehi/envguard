# EnvGuard — Implementation Steps

> Follow these steps in order. Each step is a concrete, verifiable milestone.

---

## Phase 1: Foundation (Day 1) ✅

### Step 1.1 — Bootstrap Go Module ✅
- [x] Initialize `go.mod` at `github.com/envguard/envguard`
- [x] Create directory structure (`cmd/`, `internal/`, `pkg/`, `examples/`)
- [x] Add `Makefile` with targets: `build`, `test`, `lint`, `clean`
- [x] Add `.gitignore` for Go binaries

### Step 1.2 — Define Core Types ✅
- [x] Create `internal/schema/schema.go` with `Schema`, `Variable`, `Type` structs
- [x] Create `internal/validator/result.go` with `ValidationError`, `Result` structs
- [x] Write unit tests for type serialization/deserialization

### Step 1.3 — Schema Parser ✅
- [x] Implement `schema.Parse(path string) (*Schema, error)` using `yaml.v3`
- [x] Handle unknown fields gracefully (error or ignore)
- [x] Add JSON Schema meta-validation for the YAML schema itself
- [x] Write unit tests with valid and invalid YAML samples

### Step 1.4 — .env Parser ✅
- [x] Implement `dotenv.Parse(path string) (map[string]string, error)`
- [x] Handle comments (`#`), quotes (`"`, `'`), and multiline values (basic)
- [x] Handle empty lines and malformed lines gracefully
- [x] Write unit tests

---

## Phase 2: Validation Engine (Day 2) ✅

### Step 2.1 — Type Coercion ✅
- [x] `coerceString(value string) (string, error)`
- [x] `coerceInteger(value string) (int64, error)`
- [x] `coerceFloat(value string) (float64, error)`
- [x] `coerceBoolean(value string) (bool, error)` — supports `true`/`1`/`yes`/`on`, `false`/`0`/`no`/`off`
- [x] Unit tests for each coercer

### Step 2.2 — Rule Validators ✅
- [x] `validateRequired(variable, value)` — check presence & non-empty
- [x] `validateEnum(variable, coercedValue)` — check allowed values
- [x] `validatePattern(variable, stringValue)` — regex match
- [x] `validateDefault(variable)` — inject default if missing
- [x] Unit tests for each rule

### Step 2.3 — Validation Orchestrator ✅
- [x] Implement `validator.Validate(schema, envVars, strict bool) *Result`
- [x] Iterate all schema variables; collect errors; never short-circuit early
- [x] If `strict`, detect unknown keys in `.env`
- [x] Return `Result{Valid bool, Errors []ValidationError, Warnings []ValidationError}`
- [x] Comprehensive integration tests (happy path + all error types)

---

## Phase 3: CLI (Day 3) ✅

### Step 3.1 — Cobra Setup ✅
- [x] Install `cobra` CLI dependency
- [x] Wire up `cmd/envguard/main.go` with root command

### Step 3.2 — Commands ✅
- [x] `envguard validate` with `--schema`, `--env`, `--format`, `--strict` flags
- [x] `envguard init` — write a sample `envguard.yaml` to cwd
- [x] `envguard version` — print version string

### Step 3.3 — Reporters ✅
- [x] `reporter.Text(result *Result)` — colored human-readable output
- [x] `reporter.JSON(result *Result)` — machine-readable JSON output
- [x] Ensure text output uses clear symbols (✓, ✗, ⚠) and indentation

### Step 3.4 — Exit Codes ✅
- [x] Exit 0 on success
- [x] Exit 1 on validation failure
- [x] Exit 2 on file/parse errors
- [x] E2E test: run CLI against sample files and assert exit codes

---

## Phase 4: Polish & Distribution (Day 4) ✅

### Step 4.1 — Examples ✅
- [x] Create `examples/envguard.yaml` with all field types demonstrated
- [x] Create `examples/.env` (valid) and `examples/.env.invalid` (for testing)
- [x] Add `examples/README.md` showing how to run

### Step 4.2 — Tests & CI ✅
- [x] Achieve ≥80% test coverage
- [x] Add GitHub Actions workflow: `go test`, `go vet`, `go build` on PRs
- [x] Add GitHub Actions workflow: build releases for `linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`

### Step 4.3 — Documentation ✅
- [x] Write `README.md` with installation, usage, and schema reference
- [x] Write `CHANGELOG.md`
- [x] Add `--help` text to all CLI commands

### Step 4.4 — Release ✅
- [x] Tag `v0.1.0`
- [x] Attach compiled binaries to GitHub Release
- [x] Publish installation instructions

---

## Phase 5: Language Packages (Post-MVP) ✅

### Step 5.1 — Node.js Package (`envguard-validator`) ✅
- [x] TypeScript wrapper that downloads the correct Go binary for the platform
- [x] Expose `validate(schemaPath, envPath, options)` returning parsed `Result`
- [x] Expose `validateSync()` for synchronous usage
- [x] Publish to npm as `envguard-validator`

### Step 5.2 — Python Package (`envguard-validator`) ✅
- [x] Python wrapper using `subprocess` to call the CLI
- [x] Expose `validate(schema_path, env_path, strict=False)`
- [x] Publish to PyPI as `envguard-validator`

### Step 5.3 — GitHub Action ✅
- [x] Composite action that auto-detects runner OS/arch
- [x] Downloads correct binary from GitHub releases
- [x] Runs validation with configurable inputs
- [x] CI tests on Ubuntu and macOS

---

## Phase 6: Future Ideas

### Step 6.1 — More Validation Rules
- [ ] `min` / `max` for integers and floats
- [ ] `minLength` / `maxLength` for strings
- [ ] `format: email`, `format: url`, `format: uuid`
- [ ] `allowEmpty: false` explicit opt-in

### Step 6.2 — Advanced Features
- [ ] Environment-specific conditional rules (e.g. `required_in: [prod]`)
- [ ] Secret security scanning (detect API keys, tokens)
- [ ] Drift detection / `.env.example` auto-generation
- [ ] Schema inheritance (`extends: ./base-schema.yaml`)

### Step 6.3 — Ecosystem
- [ ] VS Code extension for real-time validation
- [ ] Pre-commit hook integration
- [ ] Homebrew formula
- [ ] Java package (`envguard-java`) on Maven Central
- [ ] Docker image for CI pipelines

---

## Appendix: Quick Commands

```bash
# Dev loop
make build && ./bin/envguard validate -s examples/envguard.yaml -e examples/.env

# Run all tests
make test

# Cross-compile
make build-all

# Release (triggers all publish workflows)
git tag v0.1.5
git push origin v0.1.5
```
