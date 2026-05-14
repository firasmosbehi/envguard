# Type Coercion

EnvGuard automatically converts string values from `.env` files to the target types defined in your schema.

## Coercion Rules

| Type | Accepted Input | Rejected Input |
|------|---------------|----------------|
| `string` | Any text | — |
| `integer` | `42`, `-3`, `0` | `3.14`, `abc`, `12.0` |
| `float` | `3.14`, `-2.5`, `10`, `1.5e10` | `abc` |
| `boolean` | `true`, `false`, `1`, `0`, `yes`, `no`, `on`, `off` (case-insensitive) | `2`, `maybe`, empty string |
| `array` | `"a,b,c"` | `""` (empty string) |

## Examples

### Boolean Coercion

```yaml
env:
  DEBUG:
    type: boolean
```

```bash
DEBUG=true      # ✓ true
DEBUG=1         # ✓ true
DEBUG=yes       # ✓ true
DEBUG=FALSE     # ✓ false
DEBUG=0         # ✓ false
DEBUG=off       # ✓ false
DEBUG=maybe     # ✗ rejected
```

### Integer Coercion

```yaml
env:
  PORT:
    type: integer
```

```bash
PORT=3000       # ✓ 3000
PORT=-1         # ✓ -1
PORT=3.14       # ✗ not an integer
PORT=12.0       # ✗ not an integer
PORT=abc        # ✗ not a number
```

### Float Coercion

```yaml
env:
  RATE:
    type: float
```

```bash
RATE=3.14       # ✓ 3.14
RATE=10         # ✓ 10.0
RATE=1.5e10     # ✓ 15000000000.0
RATE=abc        # ✗ not a number
```

### Array Coercion

```yaml
env:
  ALLOWED_HOSTS:
    type: array
    separator: ","
```

```bash
ALLOWED_HOSTS=localhost,example.com    # ✓ ["localhost", "example.com"]
ALLOWED_HOSTS="a, b, c"               # ✓ ["a", " b", " c"]
ALLOWED_HOSTS=                         # ✗ empty string
```

## Strict Mode

With `--strict`, EnvGuard also checks that your `.env` file does not contain keys that are not defined in the schema:

```bash
envguard validate --strict
```

This catches typos and outdated variables.
