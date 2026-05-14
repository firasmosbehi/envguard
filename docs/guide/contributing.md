# Contributing

Thank you for your interest in contributing to EnvGuard!

## Development Setup

1. **Clone the repository**

```bash
git clone https://github.com/firasmosbehi/envguard.git
cd envguard
```

2. **Install Go** (1.22+)

3. **Build the project**

```bash
make build
```

4. **Run tests**

```bash
make test
```

## Project Structure

```
envguard/
├── cmd/envguard/          # CLI entrypoint
├── internal/              # Private implementation
│   ├── cli/               # Cobra commands
│   ├── schema/            # Schema parsing
│   ├── dotenv/            # .env file parser
│   ├── validator/         # Validation engine
│   ├── reporter/          # Output formatters
│   ├── secrets/           # Secret scanner
│   ├── watch/             # File watcher
│   ├── lsp/               # LSP server
│   └── ...
├── pkg/envguard/          # Public Go API
├── e2e/                   # End-to-end tests
├── packages/              # Language wrappers
│   ├── node/              # npm package
│   └── python/            # PyPI package
└── docs/                  # Documentation
```

## Coding Conventions

- Follow **Effective Go** and run `gofmt` / `goimports`
- Prefer explicit error handling over panics
- Keep functions small and focused
- Every package in `internal/` must have `*_test.go` files
- Target ≥80% code coverage for validator and parser packages

## Running Tests

```bash
# All tests with coverage
make test

# E2E tests
go test -v ./e2e/...

# Specific package
go test -v ./internal/validator/...
```

## Linting

```bash
make lint        # All linters
make lint-go     # Go only
make lint-fix    # Auto-fix issues
```

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Ensure all tests pass (`make test`)
5. Commit with clear messages
6. Push and open a Pull Request

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include your EnvGuard version (`envguard version`)
- Provide a minimal reproduction case
- Include schema and `.env` files if relevant

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
