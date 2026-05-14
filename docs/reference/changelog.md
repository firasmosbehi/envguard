# Changelog

All notable changes to EnvGuard are documented in this file.

## [2.0.0] - 2024-XX-XX

### Added
- **Source Code Audit** — Audit source code for `process.env` / `os.Getenv` usage and detect missing variables
- **Sync Command** — Sync `.env.example` with `.env` automatically
- **Watch Mode** — File system watcher with debounced validation and command execution
- **Interpolation** — Variable interpolation in `.env` files (`${VAR}` and `$VAR` syntax)
- **Schema Inference** — Auto-generate schemas from existing `.env` files with type/format detection
- **SARIF Output** — Full SARIF 2.1.0 support for validate, scan, audit, lint, and sync commands
- **Enhanced Secrets** — 18 built-in secret rules (added Azure, GCP, Telegram, SendGrid, Twilio, npm, Docker, Firebase, Anthropic, OpenAI)
- **Entropy Detection** — High-entropy string detection with baseline support
- **Config Files** — `.envguardrc.yaml` support with environment variable overrides
- **Severity Levels** — `critical`, `high`, `medium`, `low`, `info` severity levels
- **New Validation Rules** — `contains`, `dependsOn`/`when`, `deprecated`, `sensitive`, `transform`
- **More Format Validators** — `duration`, `semver`, `hostname`, `hex`, `cron`
- **Docs Generation** — Generate Markdown, HTML, and JSON documentation from schemas
- **Git Hooks** — `install-hook` and `uninstall-hook` commands
- **Monorepo Support** — Package discovery and per-package validation
- **Performance/Caching** — Regex compilation cache for faster repeated validations
- **LSP Server** — Language Server Protocol support for real-time editor validation
- **Schema Composition** — `extends` keyword for schema inheritance and remote schema fetching

### Changed
- Improved error messages with actionable suggestions
- Enhanced parallel validation performance
- Better JSON output stability for machine parsing

## [1.0.0] - 2024-XX-XX

### Added
- Initial release
- Schema validation with 5 types (string, integer, float, boolean, array)
- 8 built-in secret detection rules
- CLI with validate, scan, lint, init, generate-example, version commands
- Node.js and Python wrappers
- GitHub Action
- Docker image
- Homebrew formula
- Pre-commit hook
- VS Code extension
