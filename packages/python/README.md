<p align="center">
  <img src="https://raw.githubusercontent.com/firasmosbehi/envguard/main/docs/public/logo-icon.png" alt="EnvGuard" width="120">
</p>

# envguard-validator (Python)

> Python wrapper for EnvGuard — validate `.env` files against a declarative YAML schema.

## Install

```bash
pip install envguard-validator
```

The correct EnvGuard binary for your platform is downloaded automatically on first use.

## Quick Start

```python
from envguard import validate

result = validate(schema_path="envguard.yaml", env_path=".env")

if not result.valid:
    for error in result.errors:
        print(f"{error.key}: {error.message}")
    exit(1)

print("✓ Environment validated!")
```

## CLI

```bash
envguard-py validate --schema envguard.yaml --env .env
```

## API

### `validate(schema_path=None, env_path=None, strict=False)` → `ValidationResult`

Validates a `.env` file against a schema.

**Returns:**
- `ValidationResult.valid` — `bool`
- `ValidationResult.errors` — `list[ValidationError]`
- `ValidationResult.warnings` — `list[ValidationError]`

Each `ValidationError` has:
- `key` — variable name
- `message` — human-readable error
- `rule` — rule that failed (`required`, `type`, `pattern`, `enum`, `strict`)

## License

MIT
