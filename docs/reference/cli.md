# CLI Reference

## Commands

### validate

Validate `.env` file(s) against a schema.

```bash
envguard validate [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--env` | `-e` | `.env` | Path to `.env` file (repeatable) |
| `--format` | `-f` | `text` | Output format: `text`, `json`, `github`, `sarif` |
| `--strict` | | `false` | Fail if `.env` contains undefined keys |
| `--env-name` | | `""` | Environment name for `requiredIn`/`devOnly` |
| `--scan-secrets` | | `false` | Scan for hardcoded secrets |
| `--watch` | | `false` | Enable watch mode |
| `--discover` | | `false` | Auto-discover packages in monorepo |
| `--severity` | | `critical` | Minimum severity to treat as error |

Multiple `--env` files are merged **right-to-left** (later files override earlier ones).

### scan

Scan `.env` files for hardcoded secrets.

```bash
envguard scan [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--env` | `-e` | `.env` | Path to `.env` file (repeatable) |
| `--format` | `-f` | `text` | Output format: `text`, `json`, `sarif` |
| `--schema` | `-s` | `""` | Optional schema with custom secret rules |
| `--baseline` | | `""` | Baseline file to suppress known matches |

### lint

Lint a schema file for best practices.

```bash
envguard lint [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--format` | `-f` | `text` | Output format: `text`, `json` |

### init

Generate a starter `envguard.yaml` schema file.

```bash
envguard init [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--infer` | | `false` | Infer schema from existing `.env` file |
| `--env` | `-e` | `.env` | Path to `.env` file for inference |
| `--config` | | `false` | Generate `.envguardrc.yaml` config file |

### generate-example

Generate `.env.example` from a schema.

```bash
envguard generate-example [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |

### audit

Audit source code for environment variable usage.

```bash
envguard audit [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--src` | | `.` | Source code directory |
| `--env` | `-e` | `.env` | Path to `.env` file |
| `--format` | `-f` | `text` | Output format: `text`, `json`, `sarif` |

### sync

Sync `.env.example` with `.env`.

```bash
envguard sync [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--env` | `-e` | `.env` | Path to `.env` file |
| `--example` | | `.env.example` | Path to `.env.example` file |
| `--format` | `-f` | `text` | Output format: `text`, `json`, `sarif` |

### watch

Watch files and re-validate on changes.

```bash
envguard watch [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--env` | `-e` | `.env` | Path to `.env` file (repeatable) |
| `--schema` | `-s` | `envguard.yaml` | Path to schema YAML file |
| `--debounce` | | `300ms` | Debounce duration |
| `--command` | | `""` | Command to run after validation |
| `--clear` | | `false` | Clear terminal before each run |
| `--quiet` | | `false` | Only show errors |
| `--format` | `-f` | `text` | Output format |

### install-hook

Install a Git hook.

```bash
envguard install-hook <hook-name> [args...]
```

### uninstall-hook

Uninstall a Git hook.

```bash
envguard uninstall-hook <hook-name>
```

### lsp

Start the LSP server.

```bash
envguard lsp
```

### version

Print version information.

```bash
envguard version
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--help` | Show help for any command |
| `--verbose` | Enable verbose logging |
