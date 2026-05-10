# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

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
