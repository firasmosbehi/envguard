# envguard-cli

> Node.js wrapper for EnvGuard — validate `.env` files against a declarative YAML schema.

## Install

```bash
npm install envguard-cli
```

The correct EnvGuard binary for your platform is downloaded automatically via `postinstall`.

## Quick Start

```typescript
import { validate } from "envguard-cli";

const result = await validate({
  schemaPath: "envguard.yaml",
  envPath: ".env",
});

if (!result.valid) {
  for (const error of result.errors) {
    console.log(`${error.key}: ${error.message}`);
  }
  process.exit(1);
}

console.log("✓ Environment validated!");
```

## Synchronous API

```typescript
import { validateSync } from "@envguard/node";

const result = validateSync({ schemaPath: "envguard.yaml", envPath: ".env" });
```

## CLI

```bash
npx envguard-cli validate --schema envguard.yaml --env .env
```

## API

### `validate(options?)` → `Promise<ValidationResult>`

**Options:**
- `schemaPath` — path to schema YAML (default: `envguard.yaml`)
- `envPath` — path to `.env` file (default: `.env`)
- `strict` — fail on unknown keys in `.env` (default: `false`)

**Returns:**
- `ValidationResult.valid` — `boolean`
- `ValidationResult.errors` — `ValidationError[]`
- `ValidationResult.warnings` — `ValidationError[]`

Each `ValidationError` has:
- `key` — variable name
- `message` — human-readable error
- `rule` — rule that failed (`required`, `type`, `pattern`, `enum`, `strict`)

## License

MIT
