# Interpolation

EnvGuard supports variable interpolation in `.env` files, allowing values to reference other variables.

## Syntax

Use `${VAR}` or `$VAR` to reference another variable:

```bash
BASE_URL=https://api.example.com
API_URL=${BASE_URL}/v1
FULL_URL=$BASE_URL/v2/resource
```

## Default Values

Provide a default if the referenced variable is unset:

```bash
PORT=${APP_PORT:-3000}
HOST=${APP_HOST:-localhost}
```

## Nested Interpolation

```bash
PROTOCOL=https
DOMAIN=example.com
BASE_URL=${PROTOCOL}://${DOMAIN}
API_URL=${BASE_URL}/api
```

## Validation

EnvGuard resolves interpolations before validation:

```yaml
# envguard.yaml
env:
  BASE_URL:
    type: string
    format: url
    required: true

  API_URL:
    type: string
    format: url
    required: true
```

```bash
# .env
BASE_URL=https://api.example.com
API_URL=${BASE_URL}/v1
```

```bash
$ envguard validate
✓ All environment variables are valid
```

## Escaping

To include a literal `$` in a value, use `$$`:

```bash
PRICE=$$100
```

## Limitations

- Cyclic references are detected and rejected
- Interpolation only works within the same `.env` file
- Cross-file interpolation is not supported
