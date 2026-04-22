# Models Guide

Guide to configuring models for ogc with any OpenAI-compatible API provider.

**Source:** Your API provider's documentation (OpenCode Go, LM Studio, Groq, Cloudflare, etc.)

## Quick Configuration

```json
{
  "models": {
    "gpt-4o": {
      "provider": "openai",
      "model_id": "gpt-4o",
      "temperature": 0.7,
      "max_tokens": 4096
    }
  },
  "model_mapping": {
    "claude-opus": "gpt-4o",
    "claude-sonnet": "gpt-4o-mini",
    "claude-haiku": "gpt-4o-mini"
  }
}
```

## Model Mapping

The `model_mapping` section maps Claude model names to your backend models:

| Claude Model | Maps To | Notes |
| ------------ | ------- | ----- |
| `claude-opus` | Your preferred model | Primary model |
| `claude-sonnet` | Your preferred model | Balanced choice |
| `claude-haiku` | Your preferred model | Fast/cheap option |

## Provider-Specific Notes

### OpenCode Go

Models and endpoints:
- **OpenAI endpoint:** `https://opencode.ai/zen/go/v1/chat/completions`
- **Anthropic endpoint:** `https://opencode.ai/zen/go/v1/messages`

MiniMax models use the Anthropic endpoint natively.

### LM Studio

- Use your local server URL (e.g., `http://localhost:11434/v1/chat/completions`)
- Model IDs depend on your loaded model

### Groq

- **Base URL:** `https://api.groq.com/openai/v1/chat/completions`
- Use their model IDs (e.g., `llama-3.1-70b-versatile`)

## See Also

- [ogc Configuration](../configs/config.example.json)
- [README.md](../README.md) for setup instructions
- Your API provider's documentation