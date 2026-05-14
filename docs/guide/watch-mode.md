# Watch Mode

Watch mode automatically re-validates your `.env` files whenever they change. Perfect for development workflows.

## Basic Usage

```bash
envguard validate --watch
```

EnvGuard will:
1. Run an initial validation
2. Watch `.env` files and the schema for changes
3. Re-validate after a debounce period
4. Print results to the terminal

## With Custom Options

```bash
envguard validate --watch -s envguard.yaml -e .env -e .env.local --strict
```

## Watch Command

The dedicated `watch` command provides additional control:

```bash
envguard watch -e .env -s envguard.yaml
```

### Options

| Flag | Description |
|------|-------------|
| `--debounce` | Debounce duration (default: `300ms`) |
| `--command` | Run a shell command after each validation |
| `--clear` | Clear the terminal before each run |
| `--quiet` | Only show errors |

## Running Commands on Change

Trigger your test suite or application restart on validation:

```bash
envguard watch --command "npm test"
```

```bash
envguard watch --clear --command "go run ./cmd/app"
```

## Output Formats

Watch mode supports all standard output formats:

```bash
# JSON output (useful for editor integration)
envguard watch -f json

# GitHub Actions format
envguard watch -f github
```

## How It Works

EnvGuard uses [fsnotify](https://github.com/fsnotify/fsnotify) for efficient filesystem watching. Changes are debounced to avoid excessive re-validation during rapid file edits.

The watcher handles:
- File writes and renames
- File creation (for new `.env` files)
- Directory watching for monorepo setups
- Graceful shutdown on `Ctrl+C`
