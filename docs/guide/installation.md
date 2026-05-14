# Installation

EnvGuard is distributed as a single static binary with zero runtime dependencies. Choose the method that fits your workflow.

## Binary (macOS / Linux)

Download the latest release directly:

```bash
curl -sSL https://github.com/firasmosbehi/envguard/releases/latest/download/envguard-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/') -o /usr/local/bin/envguard
chmod +x /usr/local/bin/envguard
envguard version
```

Supported platforms: `linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`.

## Homebrew

```bash
brew install --formula https://raw.githubusercontent.com/firasmosbehi/envguard/main/homebrew/envguard.rb
```

## Docker

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/firasmosbehi/envguard:latest validate
```

Available tags: `latest`, `v2.0.0`, and all SemVer tags.

## Node.js

```bash
npm install -g envguard-validator
# or
npx envguard-validator validate
```

The npm package automatically downloads the correct platform binary on install.

## Python

```bash
pip install envguard-validator
```

The PyPI package lazily downloads the Go binary to `~/.envguard/bin/` on first use.

## Build from Source

Requires Go 1.22+:

```bash
git clone https://github.com/firasmosbehi/envguard.git
cd envguard
make build
./bin/envguard version
```

## VS Code Extension

Install from the marketplace: search for **EnvGuard** by `firasmosbehi`.

Or install from source:

```bash
cd vscode-extension
npm install
npm run package
```
