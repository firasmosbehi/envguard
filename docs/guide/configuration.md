# Configuration

EnvGuard can read default options from a `.envguardrc.yaml` file in your project root. This lets you run `envguard validate` without repeating flags.

## Config File Location

EnvGuard searches for `.envguardrc.yaml` in the current working directory. You can also use `.envguardrc.yml` or `.envguardrc.json`.

## Generate a Config File

```bash
envguard init --config
```

This creates `.envguardrc.yaml` with sensible defaults:

```yaml
schema: envguard.yaml
env:
  - .env
format: text
strict: false
```

## Config Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `schema` | string | `envguard.yaml` | Path to schema file |
| `env` | string[] | `[".env"]` | Environment files to validate |
| `format` | string | `text` | Output format: `text`, `json`, `github`, `sarif` |
| `strict` | boolean | `false` | Fail if `.env` contains undefined keys |
| `envName` | string | `""` | Environment name for `requiredIn`/`devOnly` |
| `scanSecrets` | boolean | `false` | Scan for hardcoded secrets |
| `watch` | boolean | `false` | Enable watch mode |

## Example

```yaml
schema: config/envguard.yaml
env:
  - .env
  - .env.local
  - .env.production
format: sarif
strict: true
envName: production
scanSecrets: true
```

## Environment Variable Overrides

Any config option can be overridden with an environment variable:

```bash
ENVGUARD_SCHEMA=other.yaml
ENVGUARD_FORMAT=json
ENVGUARD_STRICT=true
ENVGUARD_SCAN_SECRETS=true
ENVGUARD_ENV_NAME=staging
```

## Monorepo Config

For monorepos, you can define multiple packages in the config:

```yaml
packages:
  - apps/web
  - apps/api
  - packages/shared
```

EnvGuard will look for `envguard.yaml` and `.env` in each package directory.
