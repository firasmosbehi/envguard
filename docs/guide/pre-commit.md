# Pre-commit Hook

EnvGuard provides an official pre-commit hook for validating `.env` files before commits.

## Setup

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/firasmosbehi/envguard
    rev: v2.0.0
    hooks:
      - id: envguard-validate
```

## Configuration

The hook validates against `envguard.yaml` and `.env` in the repository root. Use `pass_filenames: false` and `always_run: true` so it always validates the default paths.

### Strict Mode

```yaml
repos:
  - repo: https://github.com/firasmosbehi/envguard
    rev: v2.0.0
    hooks:
      - id: envguard-validate
        args: ['--strict']
```

### Multiple Files

```yaml
repos:
  - repo: https://github.com/firasmosbehi/envguard
    rev: v2.0.0
    hooks:
      - id: envguard-validate
        args: ['-e', '.env', '-e', '.env.local']
```

### Secret Scanning

```yaml
repos:
  - repo: https://github.com/firasmosbehi/envguard
    rev: v2.0.0
    hooks:
      - id: envguard-validate
        args: ['--scan-secrets']
```

## Manual Hook Installation

If you prefer not to use pre-commit, install the hook manually:

```bash
envguard install-hook pre-commit
```

This creates `.git/hooks/pre-commit`:

```bash
#!/bin/sh
envguard validate --strict
```

Remove it later:

```bash
envguard uninstall-hook pre-commit
```

## Supported Hooks

- `pre-commit` — Validate before committing
- `pre-push` — Validate before pushing
- `commit-msg` — Validate in commit message context
