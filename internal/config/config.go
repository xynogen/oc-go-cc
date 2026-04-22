// Package config handles application configuration loading and validation.
package config

// Config holds the complete application configuration.
type Config struct {
	APIKey       string                   `json:"api_key"`
	Host         string                   `json:"host"`
	Port         int                      `json:"port"`
	Models       map[string]ModelConfig   `json:"models"`
	Fallbacks    map[string][]ModelConfig `json:"fallbacks"`
	OpenCodeGo   OpenCodeGoConfig         `json:"opencode_go"`
	Logging      LoggingConfig            `json:"logging"`
	ModelMapping map[string]string        `json:"model_mapping"`
}

// ModelMapping maps Claude Code model patterns to OpenCode Go model IDs.
// Example: "claude-sonnet-*" -> "kimi-k2.6"
// The target model ID must exist in Models config.

// ModelConfig defines routing rules for a specific model.
type ModelConfig struct {
	Provider         string  `json:"provider"`
	ModelID          string  `json:"model_id"`
	Temperature      float64 `json:"temperature"`
	MaxTokens        int     `json:"max_tokens"`
	ContextThreshold int     `json:"context_threshold"`
}

// OpenCodeGoConfig holds the upstream OpenCode Go API settings.
type OpenCodeGoConfig struct {
	BaseURL          string `json:"base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url"`
	TimeoutMs        int    `json:"timeout_ms"`
}

// LoggingConfig controls application logging behavior.
type LoggingConfig struct {
	Level    string `json:"level"`
	Requests bool   `json:"requests"`
}
