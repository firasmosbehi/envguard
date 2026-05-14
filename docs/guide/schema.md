# Schema

The EnvGuard schema is a YAML file that declaratively defines the expected shape of your environment variables.

## File Name

By default, EnvGuard looks for `envguard.yaml` in the current directory. You can specify a different file with `--schema`.

## Top-Level Structure

```yaml
version: "1.0"           # Schema version (required)
extends: "base.yaml"     # Optional: inherit from another schema file
env:                     # Map of variable names to definitions (required)
  VARIABLE_NAME:
    type: string
    required: true
    # ... more rules

secrets:                 # Optional: custom secret detection rules
  custom:
    - name: "my-token"
      pattern: "tk_[a-zA-Z0-9]{32}"
      message: "Custom token detected"
```

## Variable Definition

Every variable under `env:` supports the following fields:

| Field | Types | Description |
|-------|-------|-------------|
| `type` | all | **Required.** Data type: `string`, `integer`, `float`, `boolean`, `array` |
| `required` | all | If `true`, variable must be present and non-empty |
| `default` | all | Fallback value injected when variable is absent |
| `description` | all | Human-readable documentation |
| `message` | all | Custom error message on validation failure |
| `pattern` | `string` | Regex the value must match |
| `enum` | `string`, `integer`, `float`, `array` | Array of allowed values |
| `min` | `integer`, `float` | Minimum numeric value (inclusive) |
| `max` | `integer`, `float` | Maximum numeric value (inclusive) |
| `minLength` | `string`, `array` | Minimum length |
| `maxLength` | `string`, `array` | Maximum length |
| `format` | `string` | Built-in format validator (see below) |
| `disallow` | `string` | Array of forbidden string values |
| `requiredIn` | all | Environments where the variable is required |
| `devOnly` | all | Variable only allowed in development |
| `separator` | `array` | Delimiter for splitting array items |
| `allowEmpty` | all | If `false`, reject empty strings |
| `contains` | `array` | Require array to contain this item |
| `dependsOn` | all | Name of another variable that triggers conditional requirement |
| `when` | all | Value the `dependsOn` variable must have |
| `deprecated` | all | Warning message shown when variable is present |
| `sensitive` | all | Redact value in output |
| `transform` | `string` | Pre-validation transform: `lowercase`, `uppercase`, `trim` |

## Supported Types

### string

```yaml
API_KEY:
  type: string
  required: true
  pattern: "^[A-Za-z0-9_-]{32}$"
```

### integer

```yaml
PORT:
  type: integer
  default: 3000
  min: 1
  max: 65535
```

### float

```yaml
RATE_LIMIT:
  type: float
  default: 1.5
  min: 0.1
  max: 100.0
```

### boolean

```yaml
DEBUG:
  type: boolean
  default: false
```

Accepted values (case-insensitive): `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off`.

### array

```yaml
ALLOWED_HOSTS:
  type: array
  separator: ","
  default: "localhost,127.0.0.1"
```

## Built-in Formats

| Format | Description | Example |
|--------|-------------|---------|
| `email` | Valid email address | `user@example.com` |
| `url` | Valid URL | `https://api.example.com` |
| `uuid` | UUID v4 | `550e8400-e29b-41d4-a716-446655440000` |
| `base64` | Base64-encoded string | `SGVsbG8gV29ybGQ=` |
| `ip` | IPv4 or IPv6 address | `192.168.1.1` |
| `port` | Valid TCP/UDP port (1-65535) | `8080` |
| `json` | Valid JSON | `{"key": "value"}` |
| `duration` | Go duration string | `5m30s` |
| `semver` | Semantic version | `1.2.3` |
| `hostname` | Valid hostname | `api.example.com` |
| `hex` | Hexadecimal string | `a1b2c3` |
| `cron` | Cron expression | `0 0 * * *` |

## Example Schema

```yaml
version: "1.0"

env:
  DATABASE_URL:
    type: string
    required: true
    format: url
    description: "PostgreSQL connection string"

  PORT:
    type: integer
    default: 3000
    min: 1024
    max: 65535
    description: "HTTP server port"

  DEBUG:
    type: boolean
    default: false
    description: "Enable debug mode"

  LOG_LEVEL:
    type: string
    enum: [debug, info, warn, error]
    default: "info"
    description: "Logging verbosity"

  API_KEY:
    type: string
    required: true
    sensitive: true
    description: "External API key"

  ALLOWED_HOSTS:
    type: array
    separator: ","
    default: "localhost"
    description: "Comma-separated allowed hostnames"
```
