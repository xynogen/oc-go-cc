# Models Guide

Guide to configuring models for ogc with any OpenAI-compatible API provider.

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

## See Also

- [ogc Configuration](../configs/config.example.json)
- [README.md](../README.md) for setup instructions
- Your API provider's documentation