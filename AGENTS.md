# AGENTS.md — EnvGuard

> Agent-focused guidance for the EnvGuard project. Read this before modifying code.

---

## 1. What is EnvGuard?

EnvGuard is a **language-agnostic CLI tool** that validates `.env` files against a declarative YAML schema. It catches missing, mistyped, or malformed environment variables before deployment. The CLI is the universal core; future language-specific packages (Node.js, Python, Java) will wrap it.

**Motto:** Define once in YAML. Validate everywhere.

---

## 2. Tech Stack

- **Language:** Go 1.22+
- **CLI Framework:** `github.com/spf13/cobra`
- **YAML Parser:** `gopkg.in/yaml.v3`
- **Testing:** Standard `testing` package + `github.com/stretchr/testify` (optional)
- **Linting:** `golangci-lint` (target: zero warnings)

---

## 3. Directory Structure

```
envguard/
├── cmd/envguard/              # CLI entrypoint only
│   └── main.go
├── internal/                  # Private implementation
│   ├── cli/                   # Cobra command wiring
│   ├── schema/                # YAML schema parsing & model
│   ├── dotenv/                # .env file parser
│   ├── validator/             # Validation engine
│   └── reporter/              # Output formatters (text, json)
├── pkg/envguard/              # PUBLIC API — for future language packages
│   └── validator.go
├── schemas/
│   └── env-schema-v1.json     # JSON Schema for YAML meta-validation
├── examples/                  # Sample files for manual testing
│   ├── envguard.yaml
│   ├── .env
│   └── .env.invalid
├── Makefile
├── go.mod
├── README.md
└── AGENTS.md                  # This file
```

**Rule of thumb:**
- Put CLI-specific code in `cmd/` and `internal/cli/`
- Put reusable business logic in `internal/<domain>/`
- Put the public Go API in `pkg/envguard/` so other Go projects / future FFI wrappers can import it

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

All features must respect the `envguard.yaml` schema. Here is the canonical MVP structure:

```yaml
version: "1.0"

env:
  VARIABLE_NAME:
    type: string | integer | float | boolean
    required: true | false       # default: false
    default: <any>               # used if variable is missing (ignored if required: true)
    description: "human text"
    pattern: "regex"             # only for string type
    enum: [val1, val2]           # only for string/integer/float
```

### Type Coercion Rules (hard requirements)
| Type | Accepted Input Examples | Rejected Input |
|------|------------------------|----------------|
| `string` | any text | — |
| `integer` | `42`, `-3`, `0` | `3.14`, `abc`, `12.0` |
| `float` | `3.14`, `-2.5`, `10` | `abc` |
| `boolean` | `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off` (case-insensitive) | `2`, `maybe`, empty string |

### Validation Order
1. Check `required` (presence + non-empty)
2. Apply `default` if missing
3. Coerce to `type`
4. Check `enum` (if present)
5. Check `pattern` (if present, string only)

**Never short-circuit.** Collect ALL errors before returning.

---

## 6. CLI Behavior

### Commands
| Command | Purpose |
|---------|---------|
| `envguard validate` | Validate `.env` against schema |
| `envguard init` | Generate a starter `envguard.yaml` |
| `envguard version` | Print version |

### Flags
| Flag | Default | Description |
|------|---------|-------------|
| `--schema` / `-s` | `envguard.yaml` | Path to schema YAML |
| `--env` / `-e` | `.env` | Path to .env file |
| `--format` / `-f` | `text` | `text` or `json` |
| `--strict` | `false` | Fail on unknown keys in `.env` |

### Exit Codes
| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Validation failed |
| `2` | I/O or parse error |

**Do not change exit codes** — they are part of the public contract for CI pipelines.

---

## 7. Testing Rules

- Every package in `internal/` must have corresponding `*_test.go` files
- Target **≥80% code coverage** for the validator and parser packages
- Use table-driven tests for validation rules
- Keep test data in `testdata/` subdirectories when files are needed
- E2E tests: run the compiled binary against `examples/` and assert exit codes

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

# Run all tests
make test

# Run linter
make lint

# Clean build artifacts
make clean

# Cross-compile for all platforms
make build-all

# Quick manual validation during dev
make build && ./bin/envguard validate -s examples/envguard.yaml -e examples/.env
```

---

## 9. Design Principles

1. **Fail fast, but report everything.** Don't stop at the first error; collect all validation failures so the user can fix them in one pass.
2. **No magic.** The schema is explicit YAML. No inference from `.env.example`, no guessing types.
3. **CLI is the source of truth.** Future language packages wrap the CLI and share the same schema format. Don't add language-specific schema extensions.
4. **Zero runtime dependencies for users.** The CLI is a single static binary. Users don't need Go, Node, Python, or anything else installed.
5. **CI-first JSON output.** The `--format json` output must be stable and machine-parseable; treat it as a public API.

---

## 10. Versioning & Releases

- Follow **SemVer**: `vMAJOR.MINOR.PATCH`
- MVP target: `v0.1.0`
- Tag releases on GitHub; attach compiled binaries for:
  - `linux/amd64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- Update `CHANGELOG.md` with every release

---

## 11. When to Update This File

Update `AGENTS.md` when you:
- Add a new CLI command or flag
- Change the schema format
- Modify exit codes or JSON output structure
- Add/remove a top-level directory
- Change build tools or Go version requirements
