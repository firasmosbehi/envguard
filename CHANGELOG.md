# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2025-05-10

### Fixed
- `.gitignore` no longer ignores `cmd/envguard/` directory
- `dotenv` parser handles values up to 1MB (was crashing at 64KB)
- Empty `enum: []` is now rejected instead of being ignored
- Whitespace-only values now fail `required: true` checks
- CI test command avoids `covdata` tool incompatibility

### Added
- 77+ new unit and e2e tests covering edge cases

## [0.1.0] - 2025-05-10

### Added
- Initial CLI tool written in Go
- YAML schema definition for environment variables
- Type coercion: `string`, `integer`, `float`, `boolean`
- Validation rules: `required`, `default`, `pattern`, `enum`
- `.env` file parser supporting comments, quotes, and escape sequences
- Human-readable text output and JSON output for CI/CD
- Strict mode to warn on unknown variables in `.env`
- `envguard validate`, `envguard init`, and `envguard version` commands
- GitHub Actions CI and release workflows
- Cross-platform builds: Linux, macOS (amd64/arm64), Windows
