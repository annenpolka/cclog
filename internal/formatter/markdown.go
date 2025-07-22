package formatter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/annenpolka/cclog/pkg/types"
)

// FormatOptions controls how messages are formatted
type FormatOptions struct {
	ShowUUID         bool
	ShowPlaceholders bool
}

// FormatConversationToMarkdown converts a single conversation log to markdown with optional FormatOptions
func FormatConversationToMarkdown(log *types.ConversationLog, options ...FormatOptions) string {
	opt := FormatOptions{ShowUUID: false}
	if len(options) > 0 {
		opt = options[0]
	}
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

		sb.WriteString(formatMessage(msg, opt))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatMultipleConversationsToMarkdown converts multiple conversation logs to markdown with optional FormatOptions
func FormatMultipleConversationsToMarkdown(logs []*types.ConversationLog, options ...FormatOptions) string {
	opt := FormatOptions{ShowUUID: false}
	if len(options) > 0 {
		opt = options[0]
	}
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
			sb.WriteString(formatMessage(msg, opt))
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// formatMessage formats a single message to markdown with optional FormatOptions
func formatMessage(msg types.Message, options ...FormatOptions) string {
	opt := FormatOptions{ShowUUID: false}
	if len(options) > 0 {
		opt = options[0]
	}
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

	// Add timestamp using system timezone
	localTime := msg.Timestamp.In(GetSystemTimezone())
	sb.WriteString(fmt.Sprintf("**Time:** %s\n\n", localTime.Format("2006-01-02 15:04:05")))

	// Extract and format message content
	content := ExtractMessageContent(msg.Message, opt.ShowPlaceholders)
	if content != "" {
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	// Add metadata if present and enabled
	if opt.ShowUUID && msg.UUID != "" {
		sb.WriteString(fmt.Sprintf("*UUID: %s*\n\n", msg.UUID))
	}

	return sb.String()
}

// ExtractMessageContent extracts readable content from the message field with optional informative placeholders
func ExtractMessageContent(message interface{}, showPlaceholders ...bool) string {
	showPlaceholdersBool := false
	if len(showPlaceholders) > 0 {
		showPlaceholdersBool = showPlaceholders[0]
	}
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
		if showPlaceholdersBool {
			return generatePlaceholderForContent(str, msgMap)
		}
		return str
	}

	// Handle array content (Claude's complex message format)
	if contentArray, ok := content.([]interface{}); ok {
		var parts []string
		var hasToolUse bool
		var hasToolResult bool
		var toolNames []string
		var toolOperations []string

		for _, item := range contentArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, exists := itemMap["type"]; exists {
					switch itemType {
					case "text":
						if text, exists := itemMap["text"]; exists {
							if textStr, ok := text.(string); ok {
								parts = append(parts, textStr)
							}
						}
					case "tool_use":
						hasToolUse = true
						if toolName, exists := itemMap["name"]; exists {
							if toolNameStr, ok := toolName.(string); ok {
								toolNames = append(toolNames, toolNameStr)
							}
						}
					case "tool_result":
						hasToolResult = true
						if toolUseID, exists := itemMap["tool_use_id"]; exists {
							if toolID, ok := toolUseID.(string); ok {
								toolOperations = append(toolOperations, toolID)
							}
						}
					}
				}
			}
		}

		result := strings.Join(parts, "\n")
		if showPlaceholdersBool {
			if result == "" && (hasToolUse || hasToolResult) {
				// Generate more specific placeholder for tool operations
				return generatePlaceholderForToolOperation(msgMap, hasToolUse, hasToolResult, toolNames, toolOperations)
			}
			return generatePlaceholderForContent(result, msgMap)
		}
		return result
	}

	return fmt.Sprintf("%v", content)
}

// generatePlaceholderForContent generates informative placeholders for filtered content
func generatePlaceholderForContent(content string, msgMap map[string]interface{}) string {
	if content == "" {
		// Check for tool use result metadata for empty content
		if toolUseResult, exists := msgMap["toolUseResult"]; exists {
			if turMap, ok := toolUseResult.(map[string]interface{}); ok {
				return generatePlaceholderForToolUseResult(turMap)
			}
		}
		return "*[Empty message content]*"
	}

	// Check for system warning messages
	if strings.HasPrefix(content, "Caveat:") {
		return "*[System warning message - contains caveats about local commands]*"
	}

	// Check for command execution
	if strings.Contains(content, "<command-name>") && strings.Contains(content, "</command-name>") {
		// Extract command name
		start := strings.Index(content, "<command-name>") + len("<command-name>")
		end := strings.Index(content, "</command-name>")
		if start < end {
			commandName := content[start:end]
			return fmt.Sprintf("*[Command executed: %s]*", commandName)
		}
		return "*[Command executed]*"
	}

	// Check for command output
	if strings.Contains(content, "<local-command-stdout>") && strings.Contains(content, "</local-command-stdout>") {
		// Extract output content
		start := strings.Index(content, "<local-command-stdout>") + len("<local-command-stdout>")
		end := strings.Index(content, "</local-command-stdout>")
		if start < end {
			output := content[start:end]
			return fmt.Sprintf("*[Command output: %s]*", output)
		}
		return "*[Command output]*"
	}

	// Return original content for normal messages
	return content
}

// generatePlaceholderForToolOperation generates placeholders for tool use/result operations with empty content
func generatePlaceholderForToolOperation(msgMap map[string]interface{}, hasToolUse, hasToolResult bool, toolNames, toolOperations []string) string {
	if hasToolUse && len(toolNames) > 0 {
		if len(toolNames) == 1 {
			return fmt.Sprintf("*[Tool used: %s (no output)]*", toolNames[0])
		}
		return fmt.Sprintf("*[Tools used: %s (no output)]*", strings.Join(toolNames, ", "))
	}

	if hasToolResult {
		// Check for tool use result metadata
		if toolUseResult, exists := msgMap["toolUseResult"]; exists {
			if turMap, ok := toolUseResult.(map[string]interface{}); ok {
				return generatePlaceholderForToolUseResult(turMap)
			}
		}
		return "*[Tool operation completed (no output)]*"
	}

	return "*[Empty message content]*"
}

// generatePlaceholderForToolUseResult generates specific placeholders based on tool use result metadata
func generatePlaceholderForToolUseResult(turMap map[string]interface{}) string {
	// Check for file operations
	if opType, exists := turMap["type"]; exists {
		if typeStr, ok := opType.(string); ok {
			switch typeStr {
			case "create":
				if filePath, exists := turMap["filePath"]; exists {
					if pathStr, ok := filePath.(string); ok {
						return fmt.Sprintf("*[File created: %s (empty)]*", pathStr)
					}
				}
				return "*[File created (empty)]*"
			case "modify":
				if filePath, exists := turMap["filePath"]; exists {
					if pathStr, ok := filePath.(string); ok {
						return fmt.Sprintf("*[File modified: %s (no output)]*", pathStr)
					}
				}
				return "*[File modified (no output)]*"
			case "delete":
				if filePath, exists := turMap["filePath"]; exists {
					if pathStr, ok := filePath.(string); ok {
						return fmt.Sprintf("*[File deleted: %s]*", pathStr)
					}
				}
				return "*[File deleted]*"
			}
		}
	}

	// Check for command execution results
	if stdout, hasStdout := turMap["stdout"]; hasStdout {
		if stderr, hasStderr := turMap["stderr"]; hasStderr {
			if stdoutStr, ok := stdout.(string); ok {
				if stderrStr, ok := stderr.(string); ok {
					if stdoutStr == "" && stderrStr == "" {
						return "*[Command executed successfully (no output)]*"
					}
				}
			}
		}
	}

	// Check for interrupted status
	if interrupted, exists := turMap["interrupted"]; exists {
		if interruptedBool, ok := interrupted.(bool); ok && interruptedBool {
			return "*[Tool operation interrupted]*"
		}
	}

	return "*[Tool operation completed (no output)]*"
}
