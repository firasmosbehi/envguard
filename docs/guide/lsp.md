# LSP Server & VS Code Extension

EnvGuard includes an LSP (Language Server Protocol) server for real-time validation in editors.

## VS Code Extension

The easiest way to use LSP features is through the official VS Code extension.

### Installation

Search for **EnvGuard** in the VS Code marketplace (publisher: `firasmosbehi`).

### Features

- Real-time validation as you type `.env` files
- Diagnostics shown in the Problems panel
- Hover information for variable definitions
- Quick fixes for common issues

### Configuration

Add to `.vscode/settings.json`:

```json
{
  "envguard.schemaPath": "envguard.yaml",
  "envguard.enableValidation": true
}
```

## LSP Server

Run the LSP server standalone for other editors:

```bash
envguard lsp
```

The server communicates over stdin/stdout using the LSP protocol.

### Editor Setup

#### Neovim (nvim-lspconfig)

```lua
require('lspconfig').envguard.setup{}
```

#### Emacs (lsp-mode)

```elisp
(require 'lsp-mode)
(add-to-list 'lsp-language-id-configuration '("\\.env\\'" . "envguard"))
(lsp-register-client
 (make-lsp-client :new-connection (lsp-stdio-connection "envguard lsp")
                  :activation-fn (lsp-activate-on "envguard")
                  :server-id 'envguard))
```

#### Sublime Text (LSP)

```json
{
  "clients": {
    "envguard": {
      "enabled": true,
      "command": ["envguard", "lsp"],
      "selector": "source.env"
    }
  }
}
```

## Capabilities

The LSP server supports:
- `textDocument/diagnostic` — Publish diagnostics for `.env` files
- `textDocument/hover` — Show schema definition on hover
- `textDocument/codeAction` — Quick fixes for validation errors
