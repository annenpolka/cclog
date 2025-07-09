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
		return "(empty)"
	}

	// First, try to find a summary type message
	for _, msg := range log.Messages {
		if msg.Type == "summary" {
			if title := extractTitleFromSummary(msg); title != "" {
				return replaceNewlinesWithSpaces(title)
			}
		}
	}

	// If no summary found, use the first user message
	for _, msg := range log.Messages {
		if msg.Type == "user" && !msg.IsMeta {
			if title := extractTitleFromUserMessage(msg); title != "" {
				return replaceNewlinesWithSpaces(title)
			}
		}
	}

	return "Claude Conversation"
}

// replaceNewlinesWithSpaces replaces all newline characters with spaces
func replaceNewlinesWithSpaces(title string) string {
	// Replace various newline combinations with spaces
	title = strings.ReplaceAll(title, "\r\n", " ") // CRLF
	title = strings.ReplaceAll(title, "\n", " ")   // LF
	title = strings.ReplaceAll(title, "\r", " ")   // CR
	return title
}

// extractTitleFromSummary extracts title from summary type message
func extractTitleFromSummary(msg Message) string {
	if msg.Message == nil {
		return ""
	}

	// Try to parse as Summary struct
	if summaryMap, ok := msg.Message.(map[string]any); ok {
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
	if msgMap, ok := msg.Message.(map[string]any); ok {
		if content, exists := msgMap["content"]; exists {
			// Handle string content
			if title, ok := content.(string); ok {
				return title
			}
			
			// Handle array content
			if contentArray, ok := content.([]interface{}); ok {
				return extractTitleFromArrayContent(contentArray)
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

// extractTitleFromArrayContent extracts title from array-based content
func extractTitleFromArrayContent(contentArray []interface{}) string {
	for _, item := range contentArray {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// Handle text type blocks
			if itemType, exists := itemMap["type"]; exists {
				if itemType == "text" {
					if text, exists := itemMap["text"]; exists {
						if textStr, ok := text.(string); ok && textStr != "" {
							return textStr
						}
					}
				}
			}
			
			// Handle tool_result type blocks
			if itemType, exists := itemMap["type"]; exists {
				if itemType == "tool_result" {
					if content, exists := itemMap["content"]; exists {
						if contentStr, ok := content.(string); ok && contentStr != "" {
							return contentStr
						}
					}
				}
			}
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
