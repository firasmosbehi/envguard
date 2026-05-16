---
layout: home

hero:
  name: "EnvGuard"
  text: "Environment Variable Validator"
  tagline: Validate .env files against a declarative YAML schema. Catch misconfigurations before deployment.
  image:
    src: /logo.svg
    alt: EnvGuard
  actions:
    - theme: brand
      text: Get Started
      link: /guide/quickstart
    - theme: alt
      text: View on GitHub
      link: https://github.com/firasmosbehi/envguard

features:
  - icon: 🛡️
    title: Schema Validation
    details: Define rules once in YAML and validate everywhere. Types, patterns, enums, ranges, formats, and more.
  - icon: 🔍
    title: Secret Detection
    details: Built-in scanning for AWS keys, GitHub tokens, JWTs, Stripe keys, and 13 more patterns plus entropy-based detection.
  - icon: ⚡
    title: Blazing Fast
    details: Written in Go as a single static binary. Zero runtime dependencies. Parallel validation support.
  - icon: 🔧
    title: CI-First
    details: Native GitHub Action, pre-commit hook, SARIF output, and wrappers for Node.js and Python.
  - icon: 👁️
    title: Watch Mode
    details: Automatically re-validate on file changes with debounced filesystem watching.
  - icon: 🧠
    title: Schema Inference
    details: Auto-generate schemas from existing .env files with smart type and format detection.
---

## Demo

<video autoplay loop muted playsinline width="100%" style="border-radius: 8px; margin: 1rem 0;">
  <source src="/videos/demo.mp4" type="video/mp4">
  <img src="/videos/demo.gif" alt="EnvGuard CLI demo showing version, init, validate, scan, and generate-example commands">
</video>

## Full Feature Demo

See every major feature in action — from schema generation and validation to secret scanning, linting, schema composition, and source-code auditing.

<video controls width="100%" style="border-radius: 8px; margin: 1rem 0;">
  <source src="/videos/demo-detailed.webm" type="video/webm">
  <source src="/videos/demo-detailed.mp4" type="video/mp4">
  <img src="/videos/demo-detailed.gif" alt="Detailed EnvGuard CLI demo showing all features">
</video>

## Quick Validation

Define your schema in `envguard.yaml`:

```yaml
version: "1.0"

env:
  DATABASE_URL:
    type: string
    required: true
    format: url

  PORT:
    type: integer
    default: 3000
    min: 1024
    max: 65535

  DEBUG:
    type: boolean
    default: false
```

Run the validator:

```bash
$ envguard validate
✓ All environment variables are valid
```

## Install in 10 Seconds

::: code-group

```bash [macOS/Linux]
curl -sSL https://github.com/firasmosbehi/envguard/releases/latest/download/envguard-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/') -o /usr/local/bin/envguard
chmod +x /usr/local/bin/envguard
```

```bash [Homebrew]
brew install --formula https://raw.githubusercontent.com/firasmosbehi/envguard/main/homebrew/envguard.rb
```

```bash [Docker]
docker run --rm -v $(pwd):/workspace ghcr.io/firasmosbehi/envguard:latest validate
```

```bash [Node.js]
npm install -g envguard-validator
```

```bash [Python]
pip install envguard-validator
```

:::

## License

[MIT](https://github.com/firasmosbehi/envguard/blob/main/LICENSE)
