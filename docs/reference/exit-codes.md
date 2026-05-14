# Exit Codes

EnvGuard uses a consistent exit code scheme across all commands.

## Codes

| Code | Meaning | Commands |
|------|---------|----------|
| `0` | Success | All |
| `1` | Validation failed / Secrets found | `validate`, `scan`, `sync`, `audit` |
| `2` | I/O or parsing error | All |

## Usage in Scripts

### Bash

```bash
envguard validate
if [ $? -eq 0 ]; then
    echo "Valid!"
elif [ $? -eq 1 ]; then
    echo "Validation failed"
else
    echo "I/O error"
fi
```

### Makefile

```makefile
validate:
	envguard validate --strict

test: validate
	go test ./...
```

### CI Pipeline

```yaml
- name: Validate
  run: envguard validate --strict
  continue-on-error: false
```

## Strict Mode

With `--strict`, exit code `1` is also returned when the `.env` file contains keys not defined in the schema.

## Secret Scanning

When `--scan-secrets` is used, exit code `1` is returned if any secrets are detected, even if validation otherwise passes.
