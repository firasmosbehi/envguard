# Schema Format Reference

Complete reference for the `envguard.yaml` schema file format.

## Top-Level Fields

```yaml
version: "1.0"           # Required. Schema version.
extends: "base.yaml"     # Optional. Inherit from another schema.
env:                     # Required. Variable definitions.
  # ...
secrets:                 # Optional. Custom secret rules.
  custom:
    # ...
```

## Variable Definition

```yaml
VARIABLE_NAME:
  type: string           # Required
  required: true
  default: "fallback"
  description: "Human-readable docs"
  message: "Custom error message"
  pattern: "^regex$"
  enum: [a, b, c]
  min: 1
  max: 100
  minLength: 1
  maxLength: 255
  format: email
  disallow: [forbidden1, forbidden2]
  requiredIn: [production, staging]
  devOnly: false
  separator: ","
  allowEmpty: true
  contains: "required-item"
  dependsOn: OTHER_VAR
  when: "specific-value"
  deprecated: "Use NEW_VAR instead"
  sensitive: true
  transform: lowercase
```

## Types

| Type | Description | Example Values |
|------|-------------|----------------|
| `string` | Any text | `hello`, `https://example.com` |
| `integer` | Whole numbers | `42`, `-1`, `0` |
| `float` | Decimal numbers | `3.14`, `-2.5`, `1.5e10` |
| `boolean` | True/false | `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off` |
| `array` | Comma-separated values | `a,b,c` |

## Formats

| Format | Regex/Validation | Example |
|--------|-----------------|---------|
| `email` | RFC 5322 | `user@example.com` |
| `url` | Standard URL | `https://api.example.com/v1` |
| `uuid` | UUID v4 | `550e8400-e29b-41d4-a716-446655440000` |
| `base64` | Base64 alphabet | `SGVsbG8gV29ybGQ=` |
| `ip` | IPv4 or IPv6 | `192.168.1.1`, `::1` |
| `port` | 1-65535 | `8080` |
| `json` | Valid JSON | `{"key": "value"}` |
| `duration` | Go duration | `5m30s`, `1h` |
| `semver` | Semantic version | `1.2.3`, `2.0.0-beta.1` |
| `hostname` | RFC 1123 | `api.example.com` |
| `hex` | Hexadecimal | `a1b2c3d4` |
| `cron` | Cron expression | `0 0 * * *` |

## Constraints

- `required: true` and `default` are mutually exclusive in practice
- `enum` values must be compatible with the variable's `type`
- `pattern` is only applied to `string` types
- Empty enums (`enum: []`) are rejected as invalid schema definitions
- Whitespace-only values fail `required` checks
- `devOnly: true` and `required` / `requiredIn` are mutually exclusive
- `dependsOn` and `when` must be used together
- `allowEmpty: false` is redundant when `required: true`
- `min` cannot be greater than `max`
- `minLength` cannot be greater than `maxLength`
- `array` type **requires** a `separator`
- `transform` can only be used with `string` type
- Circular `extends` inheritance is detected and rejected

## Custom Secret Rules

```yaml
secrets:
  custom:
    - name: "internal-api-token"
      pattern: "iat_[a-zA-Z0-9]{32}"
      message: "Internal API token detected"
      severity: "high"
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Rule identifier |
| `pattern` | Yes | Regex pattern |
| `message` | Yes | Human-readable message |
| `severity` | No | `critical`, `high`, `medium`, `low` |
