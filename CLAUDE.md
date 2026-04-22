# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build   # Build binary to bin/ogc
make run     # Run without building
make test    # Run tests with race detector
make lint    # go vet + test
make clean   # Remove build artifacts
make install # Build and install to $GOPATH/bin
make dist    # Cross-compile for all platforms
```

Run a single test: `go test ./internal/router/ -v`

## Architecture

**Purpose:** ogc is a proxy server that sits between Claude Code and any OpenAI-compatible API endpoint. It intercepts Anthropic API requests, transforms them to OpenAI Chat Completions format, forwards them to your configured endpoint, and transforms responses back to Anthropic SSE.

**Config-driven model routing.** You define models and their endpoints in `~/.config/ogc/config.json`. The `model_mapping` section maps Claude model names to your backend models. No code changes needed to add/change models.

**Key features:**
- Supports ANY OpenAI-compatible API (OpenCode Go, LM Studio, Groq, Cloudflare, etc.)
- Configurable endpoints per model
- Per-model fallback chains for reliability
- Circuit breaker per model (3 failures = 30s skip)

**Two API endpoints:**
- OpenAI endpoint (`/v1/chat/completions`) — standard for most models
- Anthropic endpoint (`/v1/messages`) — for models expecting Anthropic format (e.g., MiniMax)

`internal/client/opencode.go` routes by model ID via `IsAnthropicModel()` to determine which endpoint to use.

**Polymorphic field handling:** Anthropic's `system` and `content` fields accept both strings and arrays. `pkg/types/` uses `json.RawMessage` with accessor methods (`SystemText()`, `ContentBlocks()`) to handle both formats.

## Key Files

- `cmd/ogc/main.go` — CLI entry point (cobra). Default config template is generated here.
- `internal/config/` — Config types and JSON loader with `${VAR}` env interpolation.
- `internal/transformer/` — Request/response format conversion (Anthropic ↔ OpenAI).
- `internal/router/fallback.go` — Circuit breaker per model (3 failures = 30s skip).
- `configs/config.example.json` — Reference config with all options documented.