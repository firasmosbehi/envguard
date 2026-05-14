# Git Hooks

EnvGuard can install and manage Git hooks for automatic validation.

## Install a Hook

```bash
# Install pre-commit hook
envguard install-hook pre-commit

# Install pre-push hook
envguard install-hook pre-push
```

This creates the appropriate hook script in `.git/hooks/`.

## What the Hook Does

The default hook runs:

```bash
envguard validate --strict
```

If validation fails, the commit/push is blocked.

## Custom Hook Arguments

```bash
envguard install-hook pre-commit -- --scan-secrets
```

## Uninstall a Hook

```bash
envguard uninstall-hook pre-commit
```

## Manual Hook Script

You can also write your own hook:

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Validate env files
envguard validate --strict --scan-secrets
if [ $? -ne 0 ]; then
    echo "Validation failed. Commit aborted."
    exit 1
fi

# Run tests
npm test
```

## Multiple Hooks

EnvGuard supports installing multiple validation hooks:

```bash
envguard install-hook pre-commit
envguard install-hook pre-push -- --format json
```
