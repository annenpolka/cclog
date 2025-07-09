package filepicker

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestExtractSessionID(t *testing.T) {
	tests := []struct {
		name        string
		jsonlContent string
		expected    string
		wantErr     bool
	}{
		{
			name: "valid_jsonl_with_sessionid",
			jsonlContent: `{"sessionId": "session-123", "type": "human", "message": "Hello"}
{"sessionId": "session-123", "type": "assistant", "message": "Hi there"}`,
			expected: "session-123",
			wantErr:  false,
		},
		{
			name: "empty_file",
			jsonlContent: "",
			expected: "",
			wantErr:  true,
		},
		{
			name: "invalid_json",
			jsonlContent: `{"sessionId": "session-123", "type": "human", "message": "Hello"`,
			expected: "",
			wantErr:  true,
		},
		{
			name: "missing_sessionid",
			jsonlContent: `{"type": "human", "message": "Hello"}`,
			expected: "",
			wantErr:  true,
		},
		{
			name: "empty_sessionid",
			jsonlContent: `{"sessionId": "", "type": "human", "message": "Hello"}`,
			expected: "",
			wantErr:  true,
		},
		{
			name: "whitespace_only_lines",
			jsonlContent: `
   
{"sessionId": "session-456", "type": "human", "message": "Hello"}
   `,
			expected: "session-456",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := ioutil.TempFile("", "test_*.jsonl")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write test content
			if _, err := tmpFile.WriteString(tt.jsonlContent); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}
			tmpFile.Close()

			// Test extraction
			result, err := extractSessionID(tmpFile.Name())

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

func TestExtractSessionIDFromNonJSONLFile(t *testing.T) {
	// Create a non-JSONL file
	tmpFile, err := ioutil.TempFile("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "This is not a JSONL file"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test extraction should fail
	_, err = extractSessionID(tmpFile.Name())
	if err == nil {
		t.Error("extractSessionID() expected error for non-JSONL file, got nil")
	}
}

func TestExtractSessionIDFromNonExistentFile(t *testing.T) {
	// Test with non-existent file
	_, err := extractSessionID("/non/existent/file.jsonl")
	if err == nil {
		t.Error("extractSessionID() expected error for non-existent file, got nil")
	}
}