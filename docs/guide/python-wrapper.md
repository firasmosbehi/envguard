# Python Wrapper

The `envguard-validator` PyPI package provides a Python API around the EnvGuard CLI.

## Installation

```bash
pip install envguard-validator
```

## API

### validate(schema_path, env_paths, **options)

```python
from envguard import validate

result = validate(
    'envguard.yaml',
    ['.env'],
    strict=True,
    env_name='production',
    scan_secrets=True,
)

if result.valid:
    print("✓ All valid")
else:
    print("✗ Validation failed")
    for err in result.errors:
        print(f"  {err.variable}: {err.message}")
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `strict` | bool | `False` | Fail on undefined keys |
| `env_name` | str | `""` | Environment name |
| `scan_secrets` | bool | `False` | Scan for secrets |
| `format` | str | `"json"` | Output format |

## Result Object

```python
@dataclass
class ValidationResult:
    valid: bool
    errors: list[ValidationError]
    warnings: list[ValidationWarning]

@dataclass
class ValidationError:
    variable: str
    message: str
    severity: str

@dataclass
class ValidationWarning:
    variable: str
    message: str
```

## CLI

The package also provides a CLI:

```bash
envguard-py validate -s envguard.yaml -e .env
```

## Binary Management

The Python wrapper lazily downloads the correct platform binary to `~/.envguard/bin/` on first use. No manual installation needed.
