package formatter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cclog/pkg/types"
)

// FormatConversationToMarkdown converts a single conversation log to markdown
func FormatConversationToMarkdown(log *types.ConversationLog) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Conversation Log\n\n")
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", log.FilePath))
	sb.WriteString(fmt.Sprintf("**Messages:** %d\n\n", len(log.Messages)))

	// Sort messages by timestamp for chronological order
	messages := make([]types.Message, len(log.Messages))
	copy(messages, log.Messages)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	// Process messages
	for _, msg := range messages {
		if msg.Type == "summary" {
			continue // Skip summary messages for now
		}

		sb.WriteString(formatMessage(msg))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatMultipleConversationsToMarkdown converts multiple conversation logs to markdown
func FormatMultipleConversationsToMarkdown(logs []*types.ConversationLog) string {
	var sb strings.Builder

	// Main header
	sb.WriteString("# Claude Conversation Logs\n\n")
	sb.WriteString(fmt.Sprintf("**Total Conversations:** %d\n\n", len(logs)))

	// Table of contents
	sb.WriteString("## Table of Contents\n\n")
	for i, log := range logs {
		filename := filepath.Base(log.FilePath)
		sb.WriteString(fmt.Sprintf("%d. [%s](#%s)\n", i+1, filename, 
			strings.ToLower(strings.ReplaceAll(filename, ".", ""))))
	}
	sb.WriteString("\n")

	// Individual conversations
	for _, log := range logs {
		filename := filepath.Base(log.FilePath)
		sb.WriteString(fmt.Sprintf("## %s\n\n", filename))
		
		// Sort messages by timestamp
		messages := make([]types.Message, len(log.Messages))
		copy(messages, log.Messages)
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Timestamp.Before(messages[j].Timestamp)
		})

		for _, msg := range messages {
			if msg.Type == "summary" {
				continue
			}
			sb.WriteString(formatMessage(msg))
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// formatMessage formats a single message to markdown
func formatMessage(msg types.Message) string {
	var sb strings.Builder

	// Determine message type and format accordingly
	switch msg.Type {
	case "user":
		sb.WriteString("### User\n\n")
	case "assistant":
		sb.WriteString("### Assistant\n\n")
	default:
		sb.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(msg.Type)))
	}

	// Add timestamp
	jstTime := msg.Timestamp.In(time.FixedZone("JST", 9*60*60))
	sb.WriteString(fmt.Sprintf("**Time:** %s\n\n", jstTime.Format("2006-01-02 15:04:05")))

	// Extract and format message content
	content := extractMessageContent(msg.Message)
	if content != "" {
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	// Add metadata if present
	if msg.UUID != "" {
		sb.WriteString(fmt.Sprintf("*UUID: %s*\n\n", msg.UUID))
	}

	return sb.String()
}

// extractMessageContent extracts readable content from the message field
func extractMessageContent(message interface{}) string {
	if message == nil {
		return ""
	}

	// Try to convert to map
	msgMap, ok := message.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%v", message)
	}

	// Extract content field
	content, exists := msgMap["content"]
	if !exists {
		return ""
	}

	// Handle string content
	if str, ok := content.(string); ok {
		return str
	}

	// Handle array content (Claude's complex message format)
	if contentArray, ok := content.([]interface{}); ok {
		var parts []string
		for _, item := range contentArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, exists := itemMap["type"]; exists && itemType == "text" {
					if text, exists := itemMap["text"]; exists {
						if textStr, ok := text.(string); ok {
							parts = append(parts, textStr)
						}
					}
				}
			}
		}
		return strings.Join(parts, "\n")
	}

	return fmt.Sprintf("%v", content)
}