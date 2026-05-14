# GitHub Action

The official EnvGuard GitHub Action validates environment variables in your workflows.

## Usage

```yaml
name: Validate Env

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Validate Environment Variables
        uses: firasmosbehi/envguard@v2
        with:
          schema: envguard.yaml
          env: .env
          strict: true
```

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `schema` | No | `envguard.yaml` | Path to schema file |
| `env` | No | `.env` | Path to `.env` file(s), comma-separated |
| `strict` | No | `false` | Fail on undefined keys |
| `format` | No | `github` | Output format: `text`, `json`, `github`, `sarif` |
| `env-name` | No | `""` | Environment name for `requiredIn`/`devOnly` |
| `scan-secrets` | No | `false` | Scan for hardcoded secrets |
| `version` | No | `latest` | EnvGuard version to use |

## Examples

### Strict Mode

```yaml
- uses: firasmosbehi/envguard@v2
  with:
    strict: true
```

### Multiple Environment Files

```yaml
- uses: firasmosbehi/envguard@v2
  with:
    env: .env,.env.local
```

### Secret Scanning

```yaml
- uses: firasmosbehi/envguard@v2
  with:
    scan-secrets: true
    format: sarif
```

### Production Validation

```yaml
- uses: firasmosbehi/envguard@v2
  with:
    env-name: production
    strict: true
```

### Matrix Builds

```yaml
strategy:
  matrix:
    env: [dev, staging, prod]
steps:
  - uses: firasmosbehi/envguard@v2
    with:
      schema: envguard.${{ matrix.env }}.yaml
      env-name: ${{ matrix.env }}
```

## SARIF Upload

```yaml
- uses: firasmosbehi/envguard@v2
  with:
    format: sarif
    scan-secrets: true

- uses: github/codeql-action/upload-sarif@v3
  if: always()
  with:
    sarif_file: envguard-results.sarif
```
