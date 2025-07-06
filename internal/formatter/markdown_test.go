package formatter

import (
	"strings"
	"testing"
	"time"

	"cclog/pkg/types"
)

func TestFormatConversationToMarkdownWithoutUUID(t *testing.T) {
	// Test default behavior (no UUID)
	timestamp1, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")

	log := &types.ConversationLog{
		FilePath: "/test/path/sample.jsonl",
		Messages: []types.Message{
			{
				Type:      "user",
				UUID:      "user-uuid-1",
				Timestamp: timestamp1,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
		},
	}

	markdown := FormatConversationToMarkdown(log)

	// Check that UUID is NOT included by default
	if strings.Contains(markdown, "UUID:") {
		t.Error("Markdown should not contain UUID by default")
	}

	if !strings.Contains(markdown, "Hello, how are you?") {
		t.Error("Markdown should contain user message content")
	}
}

func TestFormatConversationToMarkdownWithUUID(t *testing.T) {
	// Test with UUID enabled
	timestamp1, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")

	log := &types.ConversationLog{
		FilePath: "/test/path/sample.jsonl",
		Messages: []types.Message{
			{
				Type:      "user",
				UUID:      "user-uuid-1",
				Timestamp: timestamp1,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
		},
	}

	markdown := FormatConversationToMarkdownWithOptions(log, FormatOptions{ShowUUID: true})

	// Check that UUID IS included when enabled
	if !strings.Contains(markdown, "UUID: user-uuid-1") {
		t.Error("Markdown should contain UUID when enabled")
	}
}

func TestFormatConversationToMarkdown(t *testing.T) {
	// Create test data
	timestamp1, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")
	timestamp2, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:44.663Z")

	log := &types.ConversationLog{
		FilePath: "/test/path/sample.jsonl",
		Messages: []types.Message{
			{
				Type:      "user",
				UUID:      "user-uuid-1",
				Timestamp: timestamp1,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
			{
				Type:      "assistant",
				UUID:      "assistant-uuid-1",
				Timestamp: timestamp2,
				Message: map[string]interface{}{
					"role": "assistant",
					"content": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": "I'm doing well, thank you!",
						},
					},
				},
			},
		},
	}

	markdown := FormatConversationToMarkdown(log)

	// Check if markdown contains expected elements
	if !strings.Contains(markdown, "# Conversation Log") {
		t.Error("Markdown should contain main title")
	}

	if !strings.Contains(markdown, "**File:** `/test/path/sample.jsonl`") {
		t.Error("Markdown should contain file path")
	}

	if !strings.Contains(markdown, "## User") {
		t.Error("Markdown should contain user section")
	}

	if !strings.Contains(markdown, "## Assistant") {
		t.Error("Markdown should contain assistant section")
	}

	if !strings.Contains(markdown, "Hello, how are you?") {
		t.Error("Markdown should contain user message content")
	}

	if !strings.Contains(markdown, "I'm doing well, thank you!") {
		t.Error("Markdown should contain assistant message content")
	}

	if !strings.Contains(markdown, "2025-07-06 14:01:29") {
		t.Error("Markdown should contain formatted timestamp")
	}
}

func TestFormatMultipleConversationsToMarkdown(t *testing.T) {
	timestamp1, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")

	logs := []*types.ConversationLog{
		{
			FilePath: "/test/log1.jsonl",
			Messages: []types.Message{
				{
					Type:      "user",
					UUID:      "user-uuid-1",
					Timestamp: timestamp1,
					Message: map[string]interface{}{
						"role":    "user",
						"content": "First conversation",
					},
				},
			},
		},
		{
			FilePath: "/test/log2.jsonl",
			Messages: []types.Message{
				{
					Type:      "user",
					UUID:      "user-uuid-2",
					Timestamp: timestamp1,
					Message: map[string]interface{}{
						"role":    "user",
						"content": "Second conversation",
					},
				},
			},
		},
	}

	markdown := FormatMultipleConversationsToMarkdown(logs)

	if !strings.Contains(markdown, "# Claude Conversation Logs") {
		t.Error("Markdown should contain main title for multiple conversations")
	}

	if !strings.Contains(markdown, "First conversation") {
		t.Error("Markdown should contain first conversation content")
	}

	if !strings.Contains(markdown, "Second conversation") {
		t.Error("Markdown should contain second conversation content")
	}

	if !strings.Contains(markdown, "log1.jsonl") {
		t.Error("Markdown should contain first log filename")
	}

	if !strings.Contains(markdown, "log2.jsonl") {
		t.Error("Markdown should contain second log filename")
	}
}

func TestExtractMessageContent(t *testing.T) {
	tests := []struct {
		name     string
		message  interface{}
		expected string
	}{
		{
			name: "simple string content",
			message: map[string]interface{}{
				"role":    "user",
				"content": "Hello world",
			},
			expected: "Hello world",
		},
		{
			name: "complex content array",
			message: map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Response text",
					},
				},
			},
			expected: "Response text",
		},
		{
			name:     "nil message",
			message:  nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMessageContent(tt.message)
			if result != tt.expected {
				t.Errorf("extractMessageContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}