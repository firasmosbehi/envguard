# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [1.0.0] - 2026-05-13

### Added
- **Comprehensive linter integration** — `golangci-lint` (Go), `ESLint` (TypeScript), and `Ruff` (Python) with `make lint` / `make lint-fix`
- **Expanded test coverage** — Go statement coverage increased from 67% → 88.4%
  - `validator_rules_and_internals_test.go`: schemaToInt64/Float64, all 12 format validators, RedactSensitive, transforms
  - `cli_scan_generate_validate_test.go`: scan, generate-example, validate edge cases
  - `schema_parse_and_structural_test.go`: ParseLenient, validateEnumValue, toFloat64
  - `envguard_api_coverage_test.go`: ValidateFile strict mode, error paths
  - `e2e_commands_and_validators_test.go`: all format validators, array rules, devOnly, dependsOn, deprecated, sensitive, transform, strict mode
  - Node.js `install_platform_and_binary.test.ts`: platform detection, binary naming
  - Python `test_install.py` & `test_validator_dataclasses.py`: platform detection, dataclass validation
- **CI lint gates** — Node.js and Python lint/test jobs added to `ci.yml`

### Changed
- **Strict pre-commit hook** — `.pre-commit-hooks.yaml` now runs `envguard validate --strict`
- **Go lint fixes** — addressed `errcheck`, `errorlint`, `ineffassign`, `revive`, `gocritic`, `staticcheck`, `gofmt` across the entire codebase
- **TypeScript lint fixes** — resolved `curly` rule violations and eliminated all `any` types in catch blocks
- **Python lint fixes** — resolved D103/D101 docstrings, E501 line length, I001 import sorting; applied `ruff format`

## [0.1.8] - 2026-05-11

### Added
- **Secret scanning** — `envguard scan` command detects hardcoded credentials (AWS keys, GitHub tokens, private keys, JWTs, Stripe/Slack tokens)
- `--scan-secrets` flag for `envguard validate` to scan while validating
- **Schema inheritance** — `extends: ./base-schema.yaml` for composing schemas
- **Public Go API** — `pkg/envguard/` with `Validate()`, `ValidateFile()`, `ParseSchema()`, `ParseEnv()`
- **New format validators** — `base64`, `ip`, `port`, `json`
- **VS Code extension** — Real-time diagnostics for `.env` files via `envguard validate`

## [0.1.7] - 2026-05-10

### Added
- `allowEmpty: false` — reject empty strings even for optional variables
- `contains` for array type — require array to contain a specific item
- `dependsOn` + `when` — conditional required validation (e.g. `SSL_CERT` required when `HTTPS=true`)
- Dockerfile for containerized CI usage
- Homebrew formula for macOS/Linux installation
- GitHub Actions workflow to publish Docker images to GHCR

## [0.1.6] - 2026-05-10

### Added
- `type: array` with configurable `separator` (default `,`) for validating comma-separated values
- `minLength`/`maxLength` support for array type (validates number of items)
- `enum` support for array type (validates each item against allowed values)
- Custom error messages via `message` field on schema variables
- Multiple `--env` file support (merged right-to-left, later overrides earlier)
- `.pre-commit-hooks.yaml` for pre-commit framework integration

### Changed
- `minLength`/`maxLength` now supported for both `string` and `array` types
- `enum` validation now supports `string`, `integer`, `float`, and `array` types
- Node.js and Python wrappers updated to support multiple env paths

## [0.1.5] - 2026-05-10

### Added
- `min`/`max` validation for integers and floats
- `minLength`/`maxLength` validation for strings
- `format` validator with built-in checks: `email`, `url`, `uuid`
- `disallow` list for rejecting specific string values
- Environment-specific rules: `requiredIn` and `devOnly`
- `--env-name` CLI flag for environment-specific validation
- `envguard generate-example` command to create `.env.example` from schema
- `--output` flag for `generate-example` command

### Changed
- `validator.Validate` signature now accepts environment name as 4th parameter

## [0.1.4] - 2026-05-10

### Added
- GitHub Action (`action.yml`) for CI/CD integration
- Automated npm publish workflow (`.github/workflows/publish-npm.yml`)
- Automated PyPI publish workflow (`.github/workflows/publish-pypi.yml`)
- Cross-platform binary releases via GitHub Actions
- Test workflow for GitHub Action on Ubuntu and macOS

### Changed
- Renamed npm package from `@envguard/node` to `envguard-validator`
- Renamed PyPI package from `envguard` to `envguard-validator`
- Updated README with comprehensive documentation for all features

## [0.1.1] - 2026-05-10

### Fixed
- Scanner crash on `.env` values larger than 64KB (increased buffer to 1MB)
- Empty enum `[]` being silently ignored — now rejected as invalid schema
- Whitespace-only values (e.g., `"   "`) incorrectly passing `required` checks
- JSON output polluted with human-readable stderr text — now JSON goes to stdout, text to stderr
- CI `make test` failing on clean runners due to missing `covdata` files
- `.gitignore` blocking `cmd/envguard/` directory

### Added
- Node.js wrapper package with `validate()` and `validateSync()` APIs
- Python wrapper package with `validate()` API
- `envguard-node` and `envguard-py` CLI wrappers
- Auto-download of platform-specific binaries from GitHub releases

## [0.1.0] - 2026-05-10

### Added
- Initial release of EnvGuard CLI
- `validate` command with `--schema`, `--env`, `--format`, `--strict` flags
- `init` command to generate starter `envguard.yaml`
- `version` command
- Schema types: `string`, `integer`, `float`, `boolean`
- Validation rules: `required`, `default`, `pattern`, `enum`
- Strict mode for detecting unknown keys in `.env`
- Text and JSON output formats
- Colored human-readable error output
- Exit codes: 0 (success), 1 (validation failure), 2 (I/O error)
- 90+ unit tests with race detector
- 21 end-to-end tests
- GitHub Actions CI workflow for build, test, and vet
