package formatter

import (
	"testing"
	"time"

	"github.com/annenpolka/cclog/pkg/types"
)

func TestIsContentfulMessage(t *testing.T) {
	timestamp, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")

	tests := []struct {
		name     string
		message  types.Message
		expected bool
	}{
		{
			name: "normal user message",
			message: types.Message{
				Type:      "user",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
			expected: true,
		},
		{
			name: "normal assistant message",
			message: types.Message{
				Type:      "assistant",
				Timestamp: timestamp,
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
			expected: true,
		},
		{
			name: "system message should be filtered",
			message: types.Message{
				Type:      "system",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "system",
					"content": "System reminder",
				},
			},
			expected: false,
		},
		{
			name: "empty message should be filtered",
			message: types.Message{
				Type:      "assistant",
				Timestamp: timestamp,
				Message:   map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "API error message should be filtered",
			message: types.Message{
				Type:      "assistant",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "assistant",
					"content": "API Error: Request was aborted.",
				},
			},
			expected: false,
		},
		{
			name: "interrupted request should be filtered",
			message: types.Message{
				Type:      "user",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "[Request interrupted by user]",
				},
			},
			expected: false,
		},
		{
			name: "command message should be filtered",
			message: types.Message{
				Type:      "user",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "<command-name>/add-dir</command-name>",
				},
			},
			expected: false,
		},
		{
			name: "bash input should be filtered",
			message: types.Message{
				Type:      "user",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "<bash-input>git status</bash-input>",
				},
			},
			expected: false,
		},
		{
			name: "meta message should be filtered",
			message: types.Message{
				Type:      "user",
				Timestamp: timestamp,
				IsMeta:    true,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "Some meta content",
				},
			},
			expected: false,
		},
		{
			name: "summary message should be filtered",
			message: types.Message{
				Type:      "summary",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"summary": "Test conversation summary",
				},
			},
			expected: false,
		},
		{
			name: "message with only UUID should be filtered",
			message: types.Message{
				Type:      "assistant",
				Timestamp: timestamp,
				UUID:      "some-uuid",
				Message:   nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsContentfulMessage(tt.message)
			if result != tt.expected {
				t.Errorf("IsContentfulMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterMessages(t *testing.T) {
	timestamp, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")

	messages := []types.Message{
		{
			Type:      "user",
			Timestamp: timestamp,
			Message: map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		},
		{
			Type:      "system",
			Timestamp: timestamp,
			Message: map[string]interface{}{
				"role":    "system",
				"content": "System message",
			},
		},
		{
			Type:      "assistant",
			Timestamp: timestamp,
			Message: map[string]interface{}{
				"role":    "assistant",
				"content": "Response",
			},
		},
		{
			Type:      "assistant",
			Timestamp: timestamp,
			Message:   nil,
		},
	}

	filtered := FilterMessages(messages, true)
	
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered messages, got %d", len(filtered))
	}

	// Test with filtering disabled
	unfiltered := FilterMessages(messages, false)
	
	if len(unfiltered) != 4 {
		t.Errorf("Expected 4 unfiltered messages, got %d", len(unfiltered))
	}
}

func TestFilterConversationLog(t *testing.T) {
	timestamp, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")

	log := &types.ConversationLog{
		FilePath: "/test/path.jsonl",
		Messages: []types.Message{
			{
				Type:      "user",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "user",
					"content": "Test message",
				},
			},
			{
				Type:      "system",
				Timestamp: timestamp,
				Message: map[string]interface{}{
					"role":    "system",
					"content": "System message",
				},
			},
		},
	}

	filtered := FilterConversationLog(log, true)
	
	if len(filtered.Messages) != 1 {
		t.Errorf("Expected 1 filtered message, got %d", len(filtered.Messages))
	}

	if filtered.FilePath != log.FilePath {
		t.Errorf("Expected filepath to be preserved")
	}
}