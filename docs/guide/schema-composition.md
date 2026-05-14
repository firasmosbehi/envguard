# Schema Composition

Break large schemas into reusable pieces with inheritance and remote schemas.

## Extending Schemas

Use `extends` to inherit from a base schema:

```yaml
# base.yaml
version: "1.0"
env:
  DATABASE_URL:
    type: string
    required: true
    format: url

# production.yaml
version: "1.0"
extends: base.yaml
env:
  STRIPE_SECRET_KEY:
    type: string
    required: true
    sensitive: true
```

Variables from the base schema are merged. Child variables override parent definitions.

## Multiple Levels

```yaml
# base.yaml
version: "1.0"
env:
  PORT:
    type: integer
    default: 3000

# web.yaml
version: "1.0"
extends: base.yaml
env:
  PORT:
    type: integer
    default: 8080

# production.yaml
version: "1.0"
extends: web.yaml
env:
  SSL_CERT:
    type: string
    required: true
```

Circular inheritance is detected and rejected with an error.

## Remote Schemas

Fetch schemas from a URL:

```yaml
version: "1.0"
extends: https://raw.githubusercontent.com/org/shared-configs/main/envguard-base.yaml
env:
  MY_APP_VAR:
    type: string
```

Remote schemas are cached locally for 1 hour.

## Composition Patterns

### Environment-Specific Overrides

```
├── envguard.yaml          # base schema
├── envguard.dev.yaml      # dev overrides
├── envguard.staging.yaml  # staging overrides
└── envguard.prod.yaml     # production overrides
```

```bash
envguard validate -s envguard.prod.yaml --env-name production
```

### Shared Library Schemas

```yaml
# packages/shared/envguard.yaml
version: "1.0"
env:
  SHARED_API_URL:
    type: string
    format: url
    required: true

# apps/web/envguard.yaml
version: "1.0"
extends: ../../packages/shared/envguard.yaml
env:
  WEB_PORT:
    type: integer
    default: 3000
```
