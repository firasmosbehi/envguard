# Schema Inference

Auto-generate an EnvGuard schema from an existing `.env` file.

## Usage

```bash
envguard init --infer
```

This reads `.env`, detects types and formats, and writes `envguard.yaml`.

## Detection Heuristics

| Value Pattern | Inferred Type | Format |
|---------------|---------------|--------|
| `true`, `false` | `boolean` | — |
| `42`, `-1` | `integer` | — |
| `3.14`, `-2.5` | `float` | — |
| `a,b,c` | `array` | — |
| `user@example.com` | `string` | `email` |
| `https://...` | `string` | `url` |
| `550e8400-...` | `string` | `uuid` |
| `192.168.1.1` | `string` | `ip` |
| `eyJ...` | `string` | `jwt` |
| `1.2.3` | `string` | `semver` |
| `SG.xxx` | `string` | `sendgrid` |
| `AKIA...` | `string` | `aws-key` |
| `ghp_...` | `string` | `github-token` |

## Example

Given this `.env`:

```bash
DATABASE_URL=postgres://localhost:5432/myapp
PORT=3000
DEBUG=true
API_KEY=sk-abc123
ALLOWED_HOSTS=localhost,example.com
```

The inferred schema will be:

```yaml
version: "1.0"

env:
  DATABASE_URL:
    type: string
    format: url

  PORT:
    type: integer

  DEBUG:
    type: boolean

  API_KEY:
    type: string

  ALLOWED_HOSTS:
    type: array
    separator: ","
```

## From a Specific File

```bash
envguard init --infer --env .env.production
```

## Tips

- Review the inferred schema and add `required`, `description`, and constraints manually
- Inference is a starting point, not a finished schema
- Sensitive values are not flagged automatically — add `sensitive: true` manually
