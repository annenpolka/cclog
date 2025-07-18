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
			want: "User requested Go CLI tool development using TDD",
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
			want: "goでこれらを人間が読みやすいmarkdownにパースするコマンドラインツールを作る",
		},
		{
			name: "Return full title without truncation",
			messages: []Message{
				{
					Type:      "user",
					Message:   map[string]interface{}{"role": "user", "content": "これは非常に長いタイトルのテストです。２０文字を超えるタイトルは適切に切り詰められるべきです。"},
					Timestamp: time.Now(),
				},
			},
			want: "これは非常に長いタイトルのテストです。２０文字を超えるタイトルは適切に切り詰められるべきです。",
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
			name:     "Handle empty messages",
			messages: []Message{},
			want:     "(empty)",
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

func TestExtractTitle_NoTruncation(t *testing.T) {
	// Test that ExtractTitle returns full title without truncation
	// This allows the dynamic character limit system to work properly
	tests := []struct {
		name     string
		messages []Message
		want     string
	}{
		{
			name: "Long title should be returned in full",
			messages: []Message{
				{
					Type:      "user",
					Message:   map[string]interface{}{"role": "user", "content": "これは非常に長いタイトルのテストです。２０文字を超えるタイトルでも完全に返されるべきです。動的文字制限システムが適切に動作するように。"},
					Timestamp: time.Now(),
				},
			},
			want: "これは非常に長いタイトルのテストです。２０文字を超えるタイトルでも完全に返されるべきです。動的文字制限システムが適切に動作するように。",
		},
		{
			name: "Summary with long title should be returned in full",
			messages: []Message{
				{
					Type:      "summary",
					Message:   map[string]interface{}{"type": "summary", "summary": "User requested Go CLI tool development using Test-Driven Development methodology with comprehensive test coverage"},
					Timestamp: time.Now(),
				},
			},
			want: "User requested Go CLI tool development using Test-Driven Development methodology with comprehensive test coverage",
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

func TestExtractTitle_NewlineReplacement(t *testing.T) {
	// Test that ExtractTitle replaces newlines with spaces
	tests := []struct {
		name     string
		messages []Message
		want     string
	}{
		{
			name: "Title with newline should be replaced with space",
			messages: []Message{
				{
					Type:      "user",
					Message:   map[string]interface{}{"role": "user", "content": "これは改行を含む\nタイトルのテストです"},
					Timestamp: time.Now(),
				},
			},
			want: "これは改行を含む タイトルのテストです",
		},
		{
			name: "Title with multiple newlines should be replaced with spaces",
			messages: []Message{
				{
					Type:      "user",
					Message:   map[string]interface{}{"role": "user", "content": "複数の\n改行が\n含まれる\nタイトル"},
					Timestamp: time.Now(),
				},
			},
			want: "複数の 改行が 含まれる タイトル",
		},
		{
			name: "Title with carriage return and newline should be replaced",
			messages: []Message{
				{
					Type:      "user",
					Message:   map[string]interface{}{"role": "user", "content": "CRLF改行\r\nテスト"},
					Timestamp: time.Now(),
				},
			},
			want: "CRLF改行 テスト",
		},
		{
			name: "Summary with newlines should be replaced",
			messages: []Message{
				{
					Type:      "summary",
					Message:   map[string]interface{}{"type": "summary", "summary": "改行を含む\nサマリーの\nテスト"},
					Timestamp: time.Now(),
				},
			},
			want: "改行を含む サマリーの テスト",
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

func TestExtractTitle_ArrayContent(t *testing.T) {
	// Test that ExtractTitle handles array-based content in user messages
	tests := []struct {
		name     string
		messages []Message
		want     string
	}{
		{
			name: "User message with array content containing text",
			messages: []Message{
				{
					Type: "user",
					Message: map[string]interface{}{
						"role": "user",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "goでこれらを人間が読みやすいmarkdownにパースするコマンドラインツールを作る",
							},
						},
					},
					Timestamp: time.Now(),
				},
			},
			want: "goでこれらを人間が読みやすいmarkdownにパースするコマンドラインツールを作る",
		},
		{
			name: "User message with array content containing multiple text blocks",
			messages: []Message{
				{
					Type: "user",
					Message: map[string]interface{}{
						"role": "user",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "最初のテキストブロック",
							},
							map[string]interface{}{
								"type": "text",
								"text": "二番目のテキストブロック",
							},
						},
					},
					Timestamp: time.Now(),
				},
			},
			want: "最初のテキストブロック",
		},
		{
			name: "User message with array content containing tool result",
			messages: []Message{
				{
					Type: "user",
					Message: map[string]interface{}{
						"role": "user",
						"content": []interface{}{
							map[string]interface{}{
								"tool_use_id": "toolu_012QmdQYhr9RbGGztxjfpYmd",
								"type":        "tool_result",
								"content":     "- /Users/annenpolka/junks/cclog/",
							},
						},
					},
					Timestamp: time.Now(),
				},
			},
			want: "- /Users/annenpolka/junks/cclog/",
		},
		{
			name: "User message with array content but no text blocks should fallback",
			messages: []Message{
				{
					Type: "user",
					Message: map[string]interface{}{
						"role": "user",
						"content": []interface{}{
							map[string]interface{}{
								"type": "tool_use",
								"name": "LS",
							},
						},
					},
					Timestamp: time.Now(),
				},
			},
			want: "Claude Conversation",
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

func TestTruncateTitleWithSpecificWidth(t *testing.T) {
	tests := []struct {
		name  string
		title string
		width int
		want  string
	}{
		{
			name:  "Short title with wide width",
			title: "Short title",
			width: 30,
			want:  "Short title",
		},
		{
			name:  "Title fits exactly in width",
			title: "12345678901234567890",
			width: 20,
			want:  "12345678901234567890",
		},
		{
			name:  "Long title truncated with custom width",
			title: "This is a very long title that should be truncated",
			width: 15,
			want:  "This is a ve...",
		},
		{
			name:  "Japanese title with custom width",
			title: "これは日本語の長いタイトルです",
			width: 10,
			want:  "これは日本語の...",
		},
		{
			name:  "Very narrow width",
			title: "Any title",
			width: 5,
			want:  "An...",
		},
		{
			name:  "Width smaller than ellipsis",
			title: "Any title",
			width: 2,
			want:  "An",
		},
		{
			name:  "Zero width",
			title: "Any title",
			width: 0,
			want:  "",
		},
		{
			name:  "Wide width allows full Japanese title",
			title: "これは日本語の長いタイトルです",
			width: 50,
			want:  "これは日本語の長いタイトルです",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateTitle(tt.title, tt.width)
			if got != tt.want {
				t.Errorf("TruncateTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
