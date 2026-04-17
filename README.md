# oc-go-cc

A Go CLI proxy tool that allows you to use your OpenCode Go subscription with Claude Code.

## Overview

`oc-go-cc` intercepts Claude Code's Anthropic API requests, transforms them to OpenAI format, and forwards them to OpenCode Go. This enables you to use Claude Code's interface with the affordable models available through OpenCode Go ($5/month).

## Features

- **Model Routing**: Automatically routes requests to different models based on context (default, thinking, long context, background tasks)
- **Fallback Support**: Configurable fallback chains — if one model fails, automatically tries the next
- **Token Counting**: Uses tiktoken for accurate token counting
- **Streaming Support**: Full SSE streaming for real-time responses
- **Configurable**: JSON-based configuration with environment variable overrides

## Installation

```bash
# Build from source
git clone https://github.com/user/oc-go-cc.git
cd oc-go-cc
make build

# Or install directly
go install github.com/user/oc-go-cc/cmd/oc-go-cc@latest
```

## Quick Start

### 1. Initialize Configuration

```bash
oc-go-cc init
```

This creates a default config at `~/.config/oc-go-cc/config.json`.

### 2. Set Your API Key

```bash
export OC_GO_CC_API_KEY=sk-opencode-your-key-here
```

### 3. Start the Proxy

```bash
oc-go-cc serve
```

### 4. Configure Claude Code

In your terminal (before running `claude`):

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:3456
export ANTHROPIC_AUTH_TOKEN=unused
```

### 5. Run Claude Code

```bash
claude
```

## Configuration

### Config File Location

`~/.config/oc-go-cc/config.json`

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OC_GO_CC_API_KEY` | OpenCode Go API key (required) | - |
| `OC_GO_CC_CONFIG` | Custom config file path | `~/.config/oc-go-cc/config.json` |
| `OC_GO_CC_HOST` | Proxy host | `127.0.0.1` |
| `OC_GO_CC_PORT` | Proxy port | `3456` |
| `OC_GO_CC_OPENCODE_URL` | OpenCode Go API URL | `https://opencode.ai/zen/go/v1/chat/completions` |
| `OC_GO_CC_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

### Model Routing

The proxy automatically routes requests to different models based on the request characteristics:

| Scenario | Trigger | Config Key |
|----------|---------|------------|
| **Default** | Standard chat | `models.default` |
| **Think** | Thinking/reasoning patterns detected | `models.think` |
| **Long Context** | Token count exceeds threshold | `models.long_context` |
| **Background** | File operations, simple tasks | `models.background` |

### Fallback Configuration

If a model request fails, the proxy will automatically try fallback models in order:

```json
{
  "fallbacks": {
    "default": [
      { "provider": "opencode-go", "model_id": "glm-5" },
      { "provider": "opencode-go", "model_id": "qwen3.6-plus" }
    ]
  }
}
```

### Available Models

| Model ID | Description |
|----------|-------------|
| `glm-5.1` | Latest GLM model |
| `glm-5` | GLM model |
| `kimi-k2.5` | Kimi K2.5 |
| `mimo-v2-pro` | MiMo V2 Pro |
| `mimo-v2-omni` | MiMo V2 Omni |
| `minimax-m2.7` | MiniMax M2.7 |
| `minimax-m2.5` | MiniMax M2.5 |
| `qwen3.6-plus` | Qwen 3.6 Plus |
| `qwen3.5-plus` | Qwen 3.5 Plus |

## CLI Commands

| Command | Description |
|---------|-------------|
| `oc-go-cc serve` | Start the proxy server |
| `oc-go-cc stop` | Stop the proxy server |
| `oc-go-cc status` | Check server status |
| `oc-go-cc init` | Create default configuration |
| `oc-go-cc validate` | Validate configuration file |
| `oc-go-cc models` | List available models |

## Project Structure

```
oc-go-cc/
├── cmd/oc-go-cc/          # CLI entry point
├── internal/
│   ├── config/            # Configuration loading
│   ├── router/            # Model routing logic
│   ├── server/            # HTTP server
│   ├── handlers/          # Request handlers
│   ├── transformer/       # Request/response transformers
│   ├── client/            # OpenCode Go HTTP client
│   ├── fallback/          # Fallback system
│   └── token/             # Token counting
├── pkg/types/             # API type definitions
├── configs/               # Example configurations
├── go.mod
├── Makefile
└── README.md
```

## License

MIT
