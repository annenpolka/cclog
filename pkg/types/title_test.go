package types

import (
	"testing"
	"time"
)

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		want     string
	}{
		{
			name: "Extract title from summary type",
			messages: []Message{
				{
					Type:      "summary",
					Message:   map[string]interface{}{"type": "summary", "summary": "User requested Go CLI tool development using TDD"},
					Timestamp: time.Now(),
				},
			},
			want: "User requested Go...",
		},
		{
			name: "Extract title from first user message",
			messages: []Message{
				{
					Type:      "user",
					Message:   map[string]interface{}{"role": "user", "content": "goでこれらを人間が読みやすいmarkdownにパースするコマンドラインツールを作る"},
					Timestamp: time.Now(),
				},
			},
			want: "goでこれらを人間が読みやすいma...",
		},
		{
			name: "Truncate long title to ~20 characters",
			messages: []Message{
				{
					Type:      "user",
					Message:   map[string]interface{}{"role": "user", "content": "これは非常に長いタイトルのテストです。２０文字を超えるタイトルは適切に切り詰められるべきです。"},
					Timestamp: time.Now(),
				},
			},
			want: "これは非常に長いタイトルのテストで...",
		},
		{
			name: "Return default title when no suitable message found",
			messages: []Message{
				{
					Type:      "system",
					Message:   map[string]interface{}{"role": "system", "content": "System message"},
					Timestamp: time.Now(),
				},
			},
			want: "Claude Conversation",
		},
		{
			name: "Handle empty messages",
			messages: []Message{},
			want:     "Claude Conversation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &ConversationLog{
				Messages: tt.messages,
			}
			got := ExtractTitle(log)
			if got != tt.want {
				t.Errorf("ExtractTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "Short title unchanged",
			title: "Short title",
			want:  "Short title",
		},
		{
			name:  "Exactly 20 characters",
			title: "12345678901234567890",
			want:  "12345678901234567890",
		},
		{
			name:  "Long title truncated with ellipsis",
			title: "This is a very long title that should be truncated",
			want:  "This is a very lo...",
		},
		{
			name:  "Japanese characters handled correctly",
			title: "これは日本語の長いタイトルです",
			want:  "これは日本語の長いタイトルです",
		},
		{
			name:  "Very long Japanese title truncated",
			title: "これは非常に長い日本語のタイトルで、適切に切り詰められるべきです",
			want:  "これは非常に長い日本語のタイトルで...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateTitle(tt.title)
			if got != tt.want {
				t.Errorf("TruncateTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}