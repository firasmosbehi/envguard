# Validation Rules

EnvGuard collects **all** validation errors before returning. No short-circuiting means you fix everything in one pass.

## Validation Order

1. Check `devOnly` / `requiredIn` / `dependsOn` to determine requiredness
2. Warn if `deprecated` and variable is present
3. Check `required` (presence + non-empty after trim)
4. Check `allowEmpty`
5. Apply `default` if missing
6. Apply `transform` if specified
7. Coerce to `type`
8. Check `enum`, `pattern`, `min`/`max`, `minLength`/`maxLength`, `format`, `disallow`, `contains`

## Presence Rules

### required

```yaml
DATABASE_URL:
  type: string
  required: true
```

Fails if the variable is missing or contains only whitespace.

### allowEmpty

```yaml
OPTIONAL_NOTES:
  type: string
  allowEmpty: false
```

Rejects empty strings even when `required` is `false`.

### default

```yaml
PORT:
  type: integer
  default: 3000
```

Injects the default value when the variable is absent. Mutually exclusive with `required: true` in practice.

### requiredIn

```yaml
STRIPE_SECRET_KEY:
  type: string
  requiredIn: [production, staging]
```

Only required in specified environments. Use `--env-name` to set the environment.

### devOnly

```yaml
DEBUG:
  type: boolean
  devOnly: true
```

Only allowed in development. Skipped in other environments.

### dependsOn + when

```yaml
SMTP_HOST:
  type: string
  required: true

SMTP_PASSWORD:
  type: string
  required: true
  dependsOn: SMTP_HOST
```

`SMTP_PASSWORD` is only required when `SMTP_HOST` is present. Use `when` to check for a specific value:

```yaml
AWS_REGION:
  type: string
  required: true

AWS_ENDPOINT:
  type: string
  required: true
  dependsOn: AWS_REGION
  when: "us-east-1"
```

## Type Rules

### enum

```yaml
LOG_LEVEL:
  type: string
  enum: [debug, info, warn, error]
```

Restricts values to a fixed set. Empty enums are rejected as invalid schema definitions.

### pattern

```yaml
API_KEY:
  type: string
  pattern: "^[A-Za-z0-9]{32}$"
```

Only applies to `string` types.

### min / max

```yaml
PORT:
  type: integer
  min: 1
  max: 65535
```

`min` cannot be greater than `max`.

### minLength / maxLength

```yaml
PASSWORD:
  type: string
  minLength: 8
  maxLength: 128
```

For strings: character count. For arrays: item count.

### disallow

```yaml
FORBIDDEN_VALUE:
  type: string
  disallow: ["admin", "root"]
```

Rejects specific string values.

### contains

```yaml
ROLES:
  type: array
  separator: ","
  contains: "admin"
```

Requires the array to contain a specific item.

### transform

```yaml
USERNAME:
  type: string
  transform: lowercase
```

Transforms the value before validation. Options: `lowercase`, `uppercase`, `trim`. Only for `string` type.

## Severity Levels

Validation errors can have severity levels when used with the `--severity` flag:

- `critical` — Hard failures (exit code 1)
- `high` — Significant issues
- `medium` — Warnings that should be addressed
- `low` — Minor suggestions
- `info` — Informational notes
