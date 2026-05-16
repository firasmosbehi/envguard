# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [2.1.0] - 2026-05-16

### Added
- **Brand identity** — new logo, wordmark, icon set, and social banner in `docs/public/`
- **Documentation visuals** — terminal screenshots for validate, scan, lint, and init commands
- **Demo videos** — animated GIF/MP4/WebM demos for docs (`demo-detailed`) and LinkedIn (`demo-linkedin`)
- **Open Graph & favicon** — social preview images and multi-resolution favicons for the docs site

### Changed
- **Homepage redesign** — added "Full Feature Demo" section with detailed video walkthrough
- **Quick Start guide** — embedded real CLI output screenshots for success/failure flows
- **Secrets guide** — added screenshot showing detected secrets (AWS, GitHub tokens)
- **README branding** — main and wrapper package READMEs now display the EnvGuard logo
- **VitePress config** — updated navbar logo, OG meta tags, and Apple touch icon

## [2.0.1] - 2026-05-14

### Added
- **Audit command** — `envguard audit` scans source code for env var usage vs schema
- **Watch mode** — `envguard watch` re-validates on file changes with debounced fsnotify
- **LSP server** — `envguard lsp` for real-time diagnostics in editors
- **Sync command** — `envguard sync` keeps `.env` and `.env.example` in sync
- **Docs generation** — `envguard docs` generates Markdown/HTML/JSON from schema
- **Git hooks** — `envguard install-hook` / `uninstall-hook` for pre-commit validation
- **Monorepo support** — multi-project `.env` / schema discovery
- **Variable interpolation** — `${VAR}`, `${VAR:-default}`, `${VAR:?error}` with circular-ref detection
- **Schema cache** — RWMutex-backed cache with mtime invalidation
- **Remote schemas** — HTTP schema fetcher with `$TMPDIR` caching
- **Config file** — `.envguardrc.yaml` / `envguard.config.yaml` support with env var overrides

### Changed
- **Secret scanning** — expanded from 8 to 18 built-in rules plus entropy heuristic
- **Severity levels** — `critical` / `high` / `medium` / `low` for secret findings
- **SARIF output** — `envguard validate --format sarif` and `envguard scan --format sarif`
- **Schema composition** — `extends` supports both local files and HTTP URLs
- **Public Go API** — `ValidateParallel()` and severity-aware `Result` types

## [2.0.0] - 2026-05-13

### Added
- **Linter command** — `envguard lint` checks schemas for best practices
- **Generate-example command** — `envguard generate-example` creates `.env.example` from schema
- **New format validators** — `datetime`, `date`, `time`, `timezone`, `color`, `slug`, `filepath`, `directory`, `locale`, `jwt`, `mongodb-uri`, `redis-uri`
- **Schema fields** — `deprecated`, `sensitive`, `transform` (`lowercase` / `uppercase` / `trim`)
- **Custom secret rules** — define regex patterns in `envguard.yaml` under `secrets.custom`
- **GitHub format output** — `envguard validate --format github` for Actions annotations

### Changed
- **Node.js wrapper** — TypeScript rewrite with `validateSync()` and platform-aware binary downloader
- **Python wrapper** — dataclass-based `ValidationResult` with typed errors
- **CI/CD** — added `test-action.yml`, `publish-npm.yml`, `publish-pypi.yml`, `docker.yml`, `pages.yml`

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
