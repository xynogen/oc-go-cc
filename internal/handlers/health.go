package handlers

import (
	"encoding/json"
	"net/http"

	"oc-go-cc/internal/token"
)

// HealthHandler handles health checks and token counting endpoints.
type HealthHandler struct {
	tokenCounter *token.Counter
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(tokenCounter *token.Counter) *HealthHandler {
	return &HealthHandler{
		tokenCounter: tokenCounter,
	}
}

// HandleHealth handles GET /health.
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "oc-go-cc",
	})
}

// HandleCountTokens handles POST /v1/messages/count_tokens.
func (h *HealthHandler) HandleCountTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Count tokens.
	var messages []token.MessageContent
	for _, msg := range body.Messages {
		messages = append(messages, token.MessageContent{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	count, err := h.tokenCounter.CountMessages("", messages)
	if err != nil {
		http.Error(w, "failed to count tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{
		"token_count": count,
	})
}
