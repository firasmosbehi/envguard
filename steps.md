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

## Phase 6: Extended Validation ✅

### Step 6.1 — More Validation Rules ✅
- [x] `min` / `max` for integers and floats
- [x] `minLength` / `maxLength` for strings
- [x] `format: email`, `format: url`, `format: uuid`
- [x] `disallow` list for rejecting specific string values

### Step 6.2 — Environment-Specific Rules ✅
- [x] `requiredIn` for environment-specific required checks
- [x] `devOnly` for development-only variables
- [x] `--env-name` CLI flag

### Step 6.3 — Generate Example ✅
- [x] `envguard generate-example` command
- [x] `--output` flag for custom output path
- [x] `--include-dev` flag to include devOnly variables

---

## Phase 7: Developer Experience ✅

### Step 7.1 — Array Type ✅
- [x] `type: array` with configurable `separator`
- [x] `minLength`/`maxLength` for array item count
- [x] `enum` validation for array items

### Step 7.2 — Custom Messages ✅
- [x] `message` field on schema variables for custom error text

### Step 7.3 — Multiple Env Files ✅
- [x] Repeatable `--env` flag for multiple env files
- [x] Right-to-left merge (later files override earlier)

### Step 7.4 — Pre-commit Hook ✅
- [x] `.pre-commit-hooks.yaml` for pre-commit framework

---

## Phase 8: Packaging & Distribution ✅

### Step 8.1 — Docker Image ✅
- [x] Dockerfile with multi-arch support
- [x] GitHub Actions workflow to publish to GHCR

### Step 8.2 — Homebrew Formula ✅
- [x] Homebrew formula for macOS/Linux

---

## Phase 9: Extended Features ✅

### Step 9.1 — More Validation Rules ✅
- [x] `base64`, `ip`, `port`, `json` format validators

### Step 9.2 — Advanced Features ✅
- [x] Secret security scanning (`envguard scan`, `--scan-secrets`)
- [x] Schema inheritance (`extends: ./base-schema.yaml`)
- [x] Public Go API (`pkg/envguard/`)

### Step 9.3 — Ecosystem ✅
- [x] VS Code extension for real-time validation

---

## Phase 10: v2.0.0 — Intelligence & Integration 🚧

### Step 10.1 — Source Code Audit (`envguard audit`) 🚧
- [ ] Create `internal/audit/` package for source code analysis
- [ ] Language extractors: Go, Node.js/TS, Python, Ruby, Rust, Java
- [ ] Detect `process.env.X`, `os.Getenv()`, `std::env::var()`, etc.
- [ ] Compare detected usage against `.env` and schema
- [ ] Report: missing (used in code, not in .env), unused (in .env, not in code), undocumented (in code, not in schema)
- [ ] `--fix` flag to auto-update `.env.example` and suggest schema additions
- [ ] JSON/text output with file locations and line numbers
- [ ] Smart severity: errors for missing required, warnings for optional with fallbacks
- [ ] `--exclude` flag for file patterns to ignore (e.g., `tests/**`)
- [ ] `--ignore-vars` for known runtime variables (e.g., `CI`, `NODE_ENV`)

### Step 10.2 — .env ↔ .env.example Sync (`envguard sync`) 🚧
- [ ] Create `internal/sync/` package for bidirectional sync
- [ ] `envguard sync --env .env --example .env.example` — update example from env
- [ ] `envguard sync --check` — fail if drift detected (CI mode)
- [ ] `--schema` flag to include schema-defined variables in example
- [ ] Preserve comments and ordering where possible
- [ ] Handle `sensitive` variables by masking values in example
- [ ] `--add-missing` to add missing keys to `.env` with empty values
- [ ] `--group-by` to organize variables by category/prefix
- [ ] Generate helpful comments showing where variables are used in code
- [ ] Smart placeholders based on variable name (e.g., `DATABASE_URL=postgresql://...`)

### Step 10.3 — Watch Mode (`envguard watch`) 🚧
- [ ] File system watcher for `.env`, `envguard.yaml`, and source files
- [ ] Debounced validation (300ms default)
- [ ] `--cmd` flag to run a command on validation success (e.g., restart server)
- [ ] `--cmd-on-fail` flag to run a command on validation failure
- [ ] Clear terminal + show results on each change
- [ ] Exit with `Ctrl+C`; show summary of last run
- [ ] `--quiet` flag to only show errors
- [ ] Support for watching multiple .env files

### Step 10.4 — Variable Interpolation 🚧
- [ ] Expand `${VAR}` syntax within .env values
- [ ] Support `${VAR:-default}` for default values
- [ ] Support `${VAR:?error}` for required variable errors
- [ ] Circular reference detection and prevention
- [ ] Cross-file interpolation when multiple `--env` files provided
- [ ] Escape syntax: `\${VAR}` to prevent expansion
- [ ] Update parser to handle recursive expansion
- [ ] Unit tests for edge cases (missing vars, self-reference, deep nesting)

### Step 10.5 — Schema Inference (`envguard init --infer`) 🚧
- [ ] Analyze existing `.env` file to infer types (string, integer, boolean, url)
- [ ] Detect patterns: port numbers, URLs, booleans, database connection strings
- [ ] Generate `envguard.yaml` with inferred types and smart defaults
- [ ] Include descriptions based on variable naming conventions
- [ ] `--src` flag to also scan code for usage patterns
- [ ] `--interactive` flag to review and confirm each inferred type
- [ ] Preserve existing schema comments and ordering if file exists

### Step 10.6 — Documentation Generation (`envguard docs`) 🚧
- [ ] Generate Markdown documentation from schema
- [ ] Include variable descriptions, types, defaults, and examples
- [ ] Group variables by category/prefix
- [ ] Generate `ENVIRONMENT.md` with setup instructions
- [ ] `--format` flag: `markdown`, `html`, `json`
- [ ] `--template` flag for custom documentation templates
- [ ] Include usage examples for each variable
- [ ] Auto-generate table of contents

### Step 10.7 — New Validation Rules 🚧
- [ ] `oneOf` / `anyOf` — alternative type/schemas for a variable
- [ ] `prefix` / `suffix` — string must start/end with given substring
- [ ] Cross-variable validation: `SSL_PORT` must be > 1024 when `HTTPS=true`
- [ ] `itemType` for arrays — validate each item against a type/schema
- [ ] `uniqueItems` for arrays — reject duplicate items
- [ ] `itemPattern` for arrays — regex for each array item
- [ ] `notEmpty` for arrays — reject empty arrays
- [ ] `multipleOf` for integers/floats — value must be multiple of given number

### Step 10.8 — More Format Validators 🚧
- [ ] `format: datetime` — ISO 8601 timestamp
- [ ] `format: date` — YYYY-MM-DD
- [ ] `format: time` — HH:MM:SS
- [ ] `format: timezone` — IANA timezone identifier
- [ ] `format: color` — hex color, rgb(), rgba(), hsl()
- [ ] `format: slug` — URL-friendly slug
- [ ] `format: filepath` — valid file path (exists or not)
- [ ] `format: directory` — valid directory path
- [ ] `format: locale` — BCP 47 language tag
- [ ] `format: jwt` — JSON Web Token structure validation
- [ ] `format: mongodb-uri` — MongoDB connection string
- [ ] `format: redis-uri` — Redis connection string

### Step 10.9 — Rule Severity Levels 🚧
- [ ] Add `severity` field to schema variables: `error` (default), `warn`, `info`
- [ ] Warnings don't fail validation (exit 0) unless `--strict-warnings`
- [ ] Text reporter shows warnings with ⚠ and info with ℹ
- [ ] JSON output includes `severity` per error
- [ ] `--fail-on-warnings` flag to treat warnings as errors
- [ ] `--show-info` flag to include info-level messages in output
- [ ] Color-coded output based on severity

### Step 10.10 — Config File Support 🚧
- [ ] Support `.envguardrc`, `.envguardrc.yaml`, `envguard.config.yaml`
- [ ] Configurable defaults: schema path, env files, format, env-name, strict
- [ ] CLI flags override config file values
- [ ] Config file discovery: cwd → git root → home directory
- [ ] `envguard init --config` to generate config file
- [ ] Config validation with helpful error messages
- [ ] Environment variable overrides: `ENGUARD_SCHEMA`, `ENGUARD_STRICT`
- [ ] Config file schema documentation

### Step 10.11 — SARIF Output Format 🚧
- [ ] Create `internal/reporter/sarif.go`
- [ ] Generate valid SARIF 2.1.0 output
- [ ] Map validation errors to SARIF `results` with locations
- [ ] Include tool info, rules, and help URLs
- [ ] Integrate with GitHub Advanced Security
- [ ] Support for fingerprints to track issues across runs
- [ ] Baseline comparison support

### Step 10.12 — Enhanced Secret Scanning 🚧
- [ ] Entropy-based detection for high-entropy strings
- [ ] 10+ new built-in rules: Azure keys, GCP keys, Telegram tokens, etc.
- [ ] Configurable severity per rule: `critical`, `high`, `medium`, `low`
- [ ] `--secret-severity` flag to filter by minimum severity
- [ ] Per-rule allowlists (e.g., allow specific test API keys)
- [ ] `envguard scan --baseline` to generate/ignore existing findings
- [ ] `--scan-secrets` integration with validate command
- [ ] JSON output with rule IDs and severity

### Step 10.13 — Schema Composition 🚧
- [ ] Multiple `extends` with merge semantics
- [ ] Remote schema URLs (`extends: https://...`) with caching
- [ ] Conditional imports based on `env-name`
- [ ] Override detection: warn when child schema overrides parent rules
- [ ] Schema versioning with migration hints
- [ ] Circular dependency detection for extends
- [ ] `--extends-timeout` for remote schema fetching

### Step 10.14 — Git Hook Integration 🚧
- [ ] `envguard install-hook` command for pre-commit hooks
- [ ] Support for pre-commit and pre-push hooks
- [ ] `--type` flag: `pre-commit` (default) or `pre-push`
- [ ] `--force` flag to overwrite existing hooks
- [ ] `envguard uninstall-hook` to remove hooks
- [ ] Hook runs `envguard validate --strict` by default
- [ ] Configurable hook command via `.envguardrc`

### Step 10.15 — Monorepo Support 🚧
- [ ] Auto-detect `.env` files in subdirectories
- [ ] `envguard validate --recursive` to validate all .env files
- [ ] Per-directory schema support (`./api/envguard.yaml`, `./web/envguard.yaml`)
- [ ] Root-level schema inheritance for shared variables
- [ ] `envguard audit --recursive` for monorepo code scanning
- [ ] Service-specific `.env.example` generation
- [ ] Workspace-aware validation (pnpm, lerna, turbo)

### Step 10.16 — Performance & Caching 🚧
- [ ] Parallel validation of independent variables
- [ ] Schema parsing cache (mtime-based)
- [ ] Incremental validation for watch mode
- [ ] Benchmark suite and performance regression tests
- [ ] Memory profiling for large .env files (>1000 variables)
- [ ] Optimized regex compilation (cache compiled patterns)

### Step 10.17 — IDE Ecosystem 🚧
- [ ] JetBrains plugin (IntelliJ, PyCharm, WebStorm)
- [ ] Enhanced VS Code extension: quick-fixes, code actions, schema autocomplete
- [ ] LSP server for real-time validation in any LSP-compatible editor
- [ ] Schema autocomplete suggestions in editors
- [ ] Hover information showing variable descriptions and types

---

## Phase 11: Future Ideas (Post-v2.0.0)

### Step 11.1 — Language Wrappers
- [ ] Java package (`envguard-java`) on Maven Central
- [ ] Rust crate (`envguard`) on crates.io
- [ ] Ruby gem (`envguard-ruby`)

### Step 11.2 — DevOps Integrations
- [ ] Terraform provider for environment validation
- [ ] Kubernetes admission controller for ConfigMap validation
- [ ] Docker Compose healthcheck integration
- [ ] GitHub App for PR-level env validation comments

### Step 11.3 — Team Features
- [ ] Schema registry for sharing schemas across teams
- [ ] Role-based schema editing permissions
- [ ] Team-wide secret scanning policies
- [ ] Slack/Teams notifications for validation failures

### Step 11.4 — GUI & Web
- [ ] Web dashboard for schema editing with live preview
- [ ] Desktop GUI (Tauri or Electron) for non-CLI users
- [ ] Schema marketplace with templates for popular frameworks

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
git tag v2.0.0
git push origin v2.0.0
```
