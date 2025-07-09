package filepicker

import (
	"testing"
)

func TestExtractSessionID(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid_jsonl_filename",
			filePath: "/path/to/session-123.jsonl",
			expected: "session-123",
			wantErr:  false,
		},
		{
			name:     "valid_jsonl_filename_with_complex_sessionid",
			filePath: "/path/to/conv-2024-01-15-abc123.jsonl",
			expected: "conv-2024-01-15-abc123",
			wantErr:  false,
		},
		{
			name:     "valid_jsonl_filename_uppercase_extension",
			filePath: "/path/to/session-456.JSONL",
			expected: "session-456",
			wantErr:  false,
		},
		{
			name:     "valid_jsonl_filename_mixed_case_extension",
			filePath: "/path/to/session-789.JsonL",
			expected: "session-789",
			wantErr:  false,
		},
		{
			name:     "filename_with_dots_in_sessionid",
			filePath: "/path/to/session.with.dots.jsonl",
			expected: "session.with.dots",
			wantErr:  false,
		},
		{
			name:     "simple_filename_only",
			filePath: "simple-session.jsonl",
			expected: "simple-session",
			wantErr:  false,
		},
		{
			name:     "non_jsonl_file_txt",
			filePath: "/path/to/session-123.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "non_jsonl_file_json",
			filePath: "/path/to/session-123.json",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "file_without_extension",
			filePath: "/path/to/session-123",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty_filename_with_jsonl_extension",
			filePath: "/path/to/.jsonl",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "only_extension",
			filePath: ".jsonl",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "filename_with_multiple_extensions",
			filePath: "/path/to/backup.session-123.jsonl",
			expected: "backup.session-123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test extraction
			result, err := extractSessionID(tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("extractSessionID() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("extractSessionID() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("extractSessionID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractSessionIDEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		description string
	}{
		{
			name:     "empty_path",
			filePath: "",
			wantErr:  true,
			description: "空のパスではエラーが発生する",
		},
		{
			name:     "path_with_spaces",
			filePath: "/path/to/session with spaces.jsonl",
			wantErr:  false,
			description: "スペースを含むファイル名でも動作する",
		},
		{
			name:     "path_with_special_chars",
			filePath: "/path/to/session-123_test!@#.jsonl",
			wantErr:  false,
			description: "特殊文字を含むファイル名でも動作する",
		},
		{
			name:     "very_long_sessionid",
			filePath: "/path/to/very-long-session-id-with-many-characters-and-numbers-123456789.jsonl",
			wantErr:  false,
			description: "非常に長いセッションIDでも動作する",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractSessionID(tt.filePath)
			
			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
			}
		})
	}
}