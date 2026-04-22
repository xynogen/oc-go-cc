// Package transformer handles request/response transformation and token counting.
package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xynogen/ogc/pkg/types"
)

// ErrClientDisconnected is returned when the client disconnects during streaming.
var ErrClientDisconnected = fmt.Errorf("client disconnected")

// StreamHandler handles streaming SSE transformation from OpenAI to Anthropic format.
type StreamHandler struct {
	responseTransformer *ResponseTransformer
}

// NewStreamHandler creates a new stream handler.
func NewStreamHandler() *StreamHandler {
	return &StreamHandler{
		responseTransformer: NewResponseTransformer(),
	}
}

// ProxyStream takes an OpenAI streaming response and writes Anthropic-format SSE to the writer.
// It reads OpenAI ChatCompletionChunk SSE events and transforms them into Anthropic MessageEvent SSE events.
// The clientCtx is used to detect client disconnection and abort early.
//
// CRITICAL: This function reads directly from resp.Body without buffering to minimize latency.
// Per deep research: "Don't use bufio.Scanner or bufio.Reader on the response body - it adds buffering"
func (h *StreamHandler) ProxyStream(
	w http.ResponseWriter,
	openaiResp io.ReadCloser,
	originalModel string,
	clientCtx context.Context,
) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported by response writer")
	}

	// Generate a unique message ID for this stream.
	msgID := "msg_" + generateID()

	// Send message_start event with the full message envelope.
	msgStart := types.MessageEvent{
		Type: "message_start",
		Message: &types.MessageResponse{
			ID:      msgID,
			Type:    "message",
			Role:    "assistant",
			Content: []types.ContentBlock{},
			Model:   originalModel,
		},
	}
	if err := writeSSEEvent(w, msgStart); err != nil {
		return ErrClientDisconnected
	}
	flusher.Flush()

	// Read directly from response body without buffering.
	// Use a tight loop with a line buffer - no bufio.Reader.
	contentIndex := 0
	var lineBuf bytes.Buffer
	contentStarted := false

	// Read in larger chunks for efficiency, then parse lines
	readBuf := make([]byte, 4096)

	for {
		// Check if client disconnected
		select {
		case <-clientCtx.Done():
			return ErrClientDisconnected
		default:
		}

		// Read chunk from upstream
		n, err := openaiResp.Read(readBuf)
		if n > 0 {
			// Process bytes immediately
			for i := 0; i < n; i++ {
				b := readBuf[i]
				if b == '\n' {
					line := lineBuf.String()
					lineBuf.Reset()

					// Process complete line
					if err := h.processSSELine(w, flusher, line, &contentIndex, &contentStarted, originalModel); err != nil {
						return err
					}
				} else {
					lineBuf.WriteByte(b)
				}
			}
		}

		if err == io.EOF {
			// Process any remaining data in buffer
			if lineBuf.Len() > 0 {
				line := lineBuf.String()
				h.processSSELine(w, flusher, line, &contentIndex, &contentStarted, originalModel)
			}
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read stream: %w", err)
		}
	}

	// Send message_stop event to signal stream completion.
	stopEvent := types.MessageEvent{
		Type: "message_stop",
	}
	if err := writeSSEEvent(w, stopEvent); err != nil {
		return ErrClientDisconnected
	}
	flusher.Flush()

	return nil
}

// processSSELine processes a single SSE line from upstream.
// Per deep research: "Treat SSE primarily as a text protocol" - minimize JSON parsing.
func (h *StreamHandler) processSSELine(
	w http.ResponseWriter,
	flusher http.Flusher,
	line string,
	contentIndex *int,
	contentStarted *bool,
	originalModel string,
) error {
	line = strings.TrimSpace(line)

	// Skip empty lines
	if line == "" {
		return nil
	}

	// Skip non-data lines (event: lines, id: lines, etc.)
	if !strings.HasPrefix(line, "data: ") {
		return nil
	}

	data := strings.TrimPrefix(line, "data: ")
	if data == "" {
		return nil
	}

	// Handle [DONE] marker
	if data == "[DONE]" {
		return nil
	}

	// Fast path: extract content directly without full JSON parsing.
	// Only applies when content is non-empty and the chunk has no reasoning field
	// (which would require the full-JSON path below).
	// Kimi/DeepSeek chunks carry "reasoning" instead of "content" — those fall through.
	if idx := strings.Index(data, `"delta":{"content":"`); idx != -1 && !strings.Contains(data, `"reasoning":`) {
		start := idx + len(`"delta":{"content":"`)
		end := strings.Index(data[start:], `"`)
		if end != -1 {
			content := data[start : start+end]
			if content != "" {
				if !*contentStarted {
					*contentStarted = true
					startEvent := types.MessageEvent{
						Type:  "content_block_start",
						Index: contentIndex,
						Delta: &types.Delta{
							Type: "text",
						},
					}
					if err := writeSSEEvent(w, startEvent); err != nil {
						return ErrClientDisconnected
					}
				}

				delta := types.Delta{
					Type: "text_delta",
					Text: content,
				}
				event := types.MessageEvent{
					Type:  "content_block_delta",
					Index: contentIndex,
					Delta: &delta,
				}
				if err := writeSSEEvent(w, event); err != nil {
					return ErrClientDisconnected
				}
				flusher.Flush()
				return nil
			}
		}
	}

	// Check for finish_reason - need to send stop events
	if strings.Contains(data, `"finish_reason":`) && !strings.Contains(data, `"finish_reason":null`) {
		// Send content_block_stop
		stopEvent := types.MessageEvent{
			Type:  "content_block_stop",
			Index: contentIndex,
		}
		if err := writeSSEEvent(w, stopEvent); err != nil {
			return ErrClientDisconnected
		}

		// Send message_delta with stop_reason
		msgDelta := types.MessageEvent{
			Type: "message_delta",
			Delta: &types.Delta{
				StopReason: "end_turn", // Simplified - OpenAI usually sends "stop"
			},
		}
		if err := writeSSEEvent(w, msgDelta); err != nil {
			return ErrClientDisconnected
		}
		flusher.Flush()
		return nil
	}

	// For tool calls and other complex cases, fall back to full JSON parsing
	var chunk types.ChatCompletionChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		// Skip malformed chunks - don't fail the whole stream
		return nil
	}

	if len(chunk.Choices) == 0 {
		return nil
	}

	choice := chunk.Choices[0]

	// Handle text content deltas.
	// Extended-thinking models (Kimi, DeepSeek R1) stream their answer in
	// choice.Delta.Reasoning before switching to choice.Delta.Content.
	// We emit reasoning as text when content is absent so the client always
	// receives a non-empty response even if the context window is tight.
	textToken := choice.Delta.Content
	if textToken == "" {
		textToken = choice.Delta.Reasoning
	}

	if textToken != "" {
		if !*contentStarted {
			*contentStarted = true
			startEvent := types.MessageEvent{
				Type:  "content_block_start",
				Index: contentIndex,
				Delta: &types.Delta{
					Type: "text",
				},
			}
			if err := writeSSEEvent(w, startEvent); err != nil {
				return ErrClientDisconnected
			}
		}

		delta := types.Delta{
			Type: "text_delta",
			Text: textToken,
		}
		event := types.MessageEvent{
			Type:  "content_block_delta",
			Index: contentIndex,
			Delta: &delta,
		}
		if err := writeSSEEvent(w, event); err != nil {
			return ErrClientDisconnected
		}
		flusher.Flush()
	}

	// Handle tool call deltas
	if len(choice.Delta.ToolCalls) > 0 {
		for _, tc := range choice.Delta.ToolCalls {
			*contentIndex++

			startEvent := types.MessageEvent{
				Type:  "content_block_start",
				Index: contentIndex,
				Delta: &types.Delta{
					Type: "tool_use",
				},
			}
			if err := writeSSEEvent(w, startEvent); err != nil {
				return ErrClientDisconnected
			}

			if tc.Function.Arguments != "" {
				delta := types.Delta{
					Type:        "input_json_delta",
					PartialJSON: tc.Function.Arguments,
				}
				event := types.MessageEvent{
					Type:  "content_block_delta",
					Index: contentIndex,
					Delta: &delta,
				}
				if err := writeSSEEvent(w, event); err != nil {
					return ErrClientDisconnected
				}
			}
			flusher.Flush()
		}
	}

	// Handle finish reason
	if choice.FinishReason != "" {
		stopEvent := types.MessageEvent{
			Type:  "content_block_stop",
			Index: contentIndex,
		}
		if err := writeSSEEvent(w, stopEvent); err != nil {
			return ErrClientDisconnected
		}

		var usage *types.Usage
		if chunk.Usage != nil {
			usage = &types.Usage{
				InputTokens:  chunk.Usage.PromptTokens,
				OutputTokens: chunk.Usage.CompletionTokens,
			}
		}

		msgDelta := types.MessageEvent{
			Type: "message_delta",
			Delta: &types.Delta{
				StopReason: h.responseTransformer.mapFinishReason(choice.FinishReason),
			},
			Usage: usage,
		}
		if err := writeSSEEvent(w, msgDelta); err != nil {
			return ErrClientDisconnected
		}
		flusher.Flush()
	}

	return nil
}

// writeSSEEvent writes a single SSE event to the HTTP response writer.
// Format: "event: <type>\ndata: <json>\n\n"
func writeSSEEvent(w http.ResponseWriter, event types.MessageEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
	return err
}

// generateID creates a unique identifier based on current time.
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
