package router

import (
	"fmt"
	"strings"

	"oc-go-cc/internal/config"
)

// Scenario represents the routing scenario for model selection.
type Scenario string

const (
	ScenarioDefault     Scenario = "default"
	ScenarioBackground  Scenario = "background"
	ScenarioThink       Scenario = "think"
	ScenarioLongContext Scenario = "long_context"
)

// ScenarioResult contains the detected scenario and token count.
type ScenarioResult struct {
	Scenario   Scenario
	TokenCount int
	Reason     string
}

// MessageContent represents a single message in a conversation.
type MessageContent struct {
	Role    string
	Content string
}

// DetectScenario analyzes a request to determine which model to use.
func DetectScenario(messages []MessageContent, tokenCount int, cfg *config.Config) ScenarioResult {
	// 1. Check for long context
	threshold := getLongContextThreshold(cfg)
	if tokenCount > threshold {
		return ScenarioResult{
			Scenario:   ScenarioLongContext,
			TokenCount: tokenCount,
			Reason:     fmt.Sprintf("token count %d exceeds threshold %d", tokenCount, threshold),
		}
	}

	// 2. Check for thinking/reasoning patterns
	if hasThinkingPattern(messages) {
		return ScenarioResult{
			Scenario:   ScenarioThink,
			TokenCount: tokenCount,
			Reason:     "thinking/reasoning pattern detected",
		}
	}

	// 3. Check for background task patterns
	if hasBackgroundPattern(messages) {
		return ScenarioResult{
			Scenario:   ScenarioBackground,
			TokenCount: tokenCount,
			Reason:     "background task pattern detected",
		}
	}

	// 4. Default
	return ScenarioResult{
		Scenario:   ScenarioDefault,
		TokenCount: tokenCount,
		Reason:     "default scenario",
	}
}

// hasThinkingPattern looks for system prompts mentioning "think", "plan", "reason"
// or content containing thinking/reasoning markers.
func hasThinkingPattern(messages []MessageContent) bool {
	for _, msg := range messages {
		if msg.Role == "system" {
			lower := strings.ToLower(msg.Content)
			if strings.Contains(lower, "think") ||
				strings.Contains(lower, "plan") ||
				strings.Contains(lower, "reason") {
				return true
			}
		}
		// Check for thinking content blocks
		if strings.Contains(msg.Content, "thinking") ||
			strings.Contains(msg.Content, "antThinking") {
			return true
		}
	}
	return false
}

// hasBackgroundPattern looks for patterns that suggest background tasks
// such as file reading, directory listing, or grep operations.
func hasBackgroundPattern(messages []MessageContent) bool {
	for _, msg := range messages {
		lower := strings.ToLower(msg.Content)
		if strings.Contains(lower, "read file") ||
			strings.Contains(lower, "list directory") ||
			strings.Contains(lower, "grep") {
			return true
		}
	}
	return false
}

// getLongContextThreshold returns the configured threshold or a sensible default.
func getLongContextThreshold(cfg *config.Config) int {
	if lc, ok := cfg.Models["long_context"]; ok && lc.ContextThreshold > 0 {
		return lc.ContextThreshold
	}
	return 60000 // Default threshold
}
