package types

import (
	"encoding/json"
	"strings"
)

const (
	maxTitleLength = 20
	ellipsis       = "..."
)

// ExtractTitle extracts a suitable title from conversation log
func ExtractTitle(log *ConversationLog) string {
	if log == nil || len(log.Messages) == 0 {
		return "Claude Conversation"
	}

	// First, try to find a summary type message
	for _, msg := range log.Messages {
		if msg.Type == "summary" {
			if title := extractTitleFromSummary(msg); title != "" {
				return title
			}
		}
	}

	// If no summary found, use the first user message
	for _, msg := range log.Messages {
		if msg.Type == "user" && !msg.IsMeta {
			if title := extractTitleFromUserMessage(msg); title != "" {
				return title
			}
		}
	}

	return "Claude Conversation"
}

// extractTitleFromSummary extracts title from summary type message
func extractTitleFromSummary(msg Message) string {
	if msg.Message == nil {
		return ""
	}

	// Try to parse as Summary struct
	if summaryMap, ok := msg.Message.(map[string]interface{}); ok {
		if summaryText, exists := summaryMap["summary"]; exists {
			if title, ok := summaryText.(string); ok {
				return title
			}
		}
	}

	// Try to parse as JSON
	if msgBytes, err := json.Marshal(msg.Message); err == nil {
		var summary Summary
		if err := json.Unmarshal(msgBytes, &summary); err == nil {
			return summary.Summary
		}
	}

	return ""
}

// extractTitleFromUserMessage extracts title from user message
func extractTitleFromUserMessage(msg Message) string {
	if msg.Message == nil {
		return ""
	}

	// Try to parse as map
	if msgMap, ok := msg.Message.(map[string]interface{}); ok {
		if content, exists := msgMap["content"]; exists {
			if title, ok := content.(string); ok {
				return title
			}
		}
	}

	// Try to parse as JSON
	if msgBytes, err := json.Marshal(msg.Message); err == nil {
		var claudeMsg ClaudeMessage
		if err := json.Unmarshal(msgBytes, &claudeMsg); err == nil {
			return claudeMsg.Content
		}
	}

	return ""
}

// TruncateTitle truncates title to appropriate length
func TruncateTitle(title string) string {
	return TruncateTitleWithWidth(title, maxTitleLength)
}

// TruncateTitleWithWidth truncates title to specified width
func TruncateTitleWithWidth(title string, width int) string {
	if title == "" || width <= 0 {
		return ""
	}

	// Remove leading/trailing whitespace
	title = strings.TrimSpace(title)

	// Count runes (not bytes) for proper Unicode handling
	runes := []rune(title)
	
	if len(runes) <= width {
		return title
	}

	// Handle case where width is smaller than ellipsis
	ellipsisRunes := []rune(ellipsis)
	if width <= len(ellipsisRunes) {
		return string(runes[:width])
	}

	// Truncate and add ellipsis
	truncated := string(runes[:width-len(ellipsisRunes)])
	return truncated + ellipsis
}