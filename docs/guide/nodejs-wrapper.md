# Node.js Wrapper

The `envguard-validator` npm package provides a JavaScript/TypeScript API around the EnvGuard CLI.

## Installation

```bash
npm install envguard-validator
```

## API

### validate(schemaPath, envPaths, options)

```javascript
const { validate } = require('envguard-validator');

async function main() {
  const result = await validate('envguard.yaml', ['.env'], {
    strict: true,
    envName: 'production',
    scanSecrets: true,
  });

  if (result.valid) {
    console.log('✓ All valid');
  } else {
    console.log('✗ Validation failed');
    for (const err of result.errors) {
      console.log(`  ${err.variable}: ${err.message}`);
    }
  }
}
```

### validateSync(schemaPath, envPaths, options)

```javascript
const { validateSync } = require('envguard-validator');

const result = validateSync('envguard.yaml', ['.env']);
console.log(result.valid);
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `strict` | boolean | `false` | Fail on undefined keys |
| `envName` | string | `""` | Environment name |
| `scanSecrets` | boolean | `false` | Scan for secrets |
| `format` | string | `"json"` | Output format |

## Result Object

```typescript
interface ValidationResult {
  valid: boolean;
  errors: Array<{
    variable: string;
    message: string;
    severity: string;
  }>;
  warnings: Array<{
    variable: string;
    message: string;
  }>;
}
```

## CLI

The package also provides an `npx` CLI:

```bash
npx envguard-validator validate -s envguard.yaml -e .env
```
