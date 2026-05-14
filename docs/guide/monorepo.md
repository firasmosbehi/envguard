# Monorepo Support

EnvGuard supports monorepos with multiple packages, each with its own schema and `.env` files.

## Discovery

EnvGuard can auto-discover packages containing `envguard.yaml`:

```bash
envguard validate --discover
```

This searches the current directory and subdirectories for `envguard.yaml` files.

## Manual Configuration

Define packages in `.envguardrc.yaml`:

```yaml
packages:
  - apps/web
  - apps/api
  - packages/shared
```

EnvGuard will validate each package in sequence:

```bash
$ envguard validate
[apps/web] ✓ All environment variables are valid
[apps/api] ✓ All environment variables are valid
[packages/shared] ✗ PORT: expected integer, got "abc"
```

## Per-Package Schema

Each package directory should contain its own `envguard.yaml` and `.env`:

```
├── apps/
│   ├── web/
│   │   ├── envguard.yaml
│   │   └── .env
│   └── api/
│       ├── envguard.yaml
│       └── .env
├── packages/
│   └── shared/
│       ├── envguard.yaml
│       └── .env
└── .envguardrc.yaml
```

## Inheritance

Packages can extend a base schema:

```yaml
# packages/shared/envguard.yaml
version: "1.0"
env:
  SHARED_VAR:
    type: string
    required: true

# apps/web/envguard.yaml
version: "1.0"
extends: ../../packages/shared/envguard.yaml
env:
  WEB_VAR:
    type: string
```

## CI Integration

For monorepo CI pipelines, validate all packages at once:

```yaml
- name: Validate Env
  run: envguard validate --discover --format sarif
```
