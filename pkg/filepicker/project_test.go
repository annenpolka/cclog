package filepicker

import (
	"os"
	"testing"
	"time"
)

func TestExtractProjectName(t *testing.T) {
	tests := []struct {
		name     string
		cwd      string
		expected string
	}{
		{
			name:     "regular project path",
			cwd:      "/Users/annenpolka/junks/cclog",
			expected: "cclog",
		},
		{
			name:     "nested project path",
			cwd:      "/Users/annenpolka/junks/hello-rails",
			expected: "hello-rails",
		},
		{
			name:     "deep nested path",
			cwd:      "/Users/annenpolka/workspace/projects/my-app",
			expected: "my-app",
		},
		{
			name:     "root path",
			cwd:      "/",
			expected: "",
		},
		{
			name:     "empty path",
			cwd:      "",
			expected: "",
		},
		{
			name:     "path with trailing slash",
			cwd:      "/Users/annenpolka/junks/cclog/",
			expected: "cclog",
		},
		{
			name:     "single level path",
			cwd:      "/cclog",
			expected: "cclog",
		},
		{
			name:     "desktop path",
			cwd:      "/Users/annenpolka/Desktop",
			expected: "Desktop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProjectName(tt.cwd)
			if result != tt.expected {
				t.Errorf("extractProjectName(%q) = %q, want %q", tt.cwd, result, tt.expected)
			}
		})
	}
}

func TestFileInfo_WithProjectName(t *testing.T) {
	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "JSONL file with project name should display date [project] title",
			file: FileInfo{
				Name:              "conversation.jsonl",
				Path:              "/path/conversation.jsonl",
				IsDir:             false,
				ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ConversationTitle: "User requested Go...",
				ProjectName:       "cclog",
			},
			expected: "2025-01-15 14:30 [cclog] User requested Go...",
		},
		{
			name: "JSONL file with project name but no title should display date [project]",
			file: FileInfo{
				Name:              "empty.jsonl",
				Path:              "/path/empty.jsonl",
				IsDir:             false,
				ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ConversationTitle: "",
				ProjectName:       "cclog",
			},
			expected: "2025-01-15 14:30 [cclog]",
		},
		{
			name: "JSONL file without project name should display date title",
			file: FileInfo{
				Name:              "conversation.jsonl",
				Path:              "/path/conversation.jsonl",
				IsDir:             false,
				ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ConversationTitle: "User requested Go...",
				ProjectName:       "",
			},
			expected: "2025-01-15 14:30 User requested Go...",
		},
		{
			name: "Non-JSONL file should display just filename",
			file: FileInfo{
				Name:              "document.txt",
				Path:              "/path/document.txt",
				IsDir:             false,
				ConversationTitle: "",
				ProjectName:       "cclog",
			},
			expected: "document.txt",
		},
		{
			name: "Directory should display with slash",
			file: FileInfo{
				Name:              "folder",
				Path:              "/path/folder",
				IsDir:             true,
				ConversationTitle: "",
				ProjectName:       "cclog",
			},
			expected: "folder/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.file.Title(); got != tt.expected {
				t.Errorf("Title() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractConversationInfo(t *testing.T) {
	tests := []struct {
		name            string
		jsonlContent    string
		expectedTitle   string
		expectedProject string
	}{
		{
			name: "should extract project name from CWD field",
			jsonlContent: `{"type":"user","message":{"role":"user","content":"Hello"},"cwd":"/Users/annenpolka/junks/cclog","uuid":"test","timestamp":"2025-07-06T05:01:44.663Z"}
{"type":"assistant","message":{"role":"assistant","content":"Hi there"},"cwd":"/Users/annenpolka/junks/cclog","uuid":"test2","timestamp":"2025-07-06T05:01:45.663Z"}`,
			expectedTitle:   "Hello",
			expectedProject: "cclog",
		},
		{
			name: "should extract different project name",
			jsonlContent: `{"type":"user","message":{"role":"user","content":"Test message"},"cwd":"/Users/annenpolka/junks/hello-rails","uuid":"test","timestamp":"2025-07-06T05:01:44.663Z"}
{"type":"assistant","message":{"role":"assistant","content":"Test response"},"cwd":"/Users/annenpolka/junks/hello-rails","uuid":"test2","timestamp":"2025-07-06T05:01:45.663Z"}`,
			expectedTitle:   "Test message",
			expectedProject: "hello-rails",
		},
		{
			name: "should handle empty CWD",
			jsonlContent: `{"type":"user","message":{"role":"user","content":"Test message"},"cwd":"","uuid":"test","timestamp":"2025-07-06T05:01:44.663Z"}
{"type":"assistant","message":{"role":"assistant","content":"Test response"},"cwd":"","uuid":"test2","timestamp":"2025-07-06T05:01:45.663Z"}`,
			expectedTitle:   "Test message",
			expectedProject: "",
		},
		{
			name: "should handle missing CWD",
			jsonlContent: `{"type":"user","message":{"role":"user","content":"Test message"},"uuid":"test","timestamp":"2025-07-06T05:01:44.663Z"}
{"type":"assistant","message":{"role":"assistant","content":"Test response"},"uuid":"test2","timestamp":"2025-07-06T05:01:45.663Z"}`,
			expectedTitle:   "Test message",
			expectedProject: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with test content
			tempFile, err := os.CreateTemp("", "test_*.jsonl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			if _, err := tempFile.WriteString(tt.jsonlContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tempFile.Close()

			// Test extractConversationInfo
			title, projectName := extractConversationInfo(tempFile.Name())

			if title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
			}

			if projectName != tt.expectedProject {
				t.Errorf("Expected project name %q, got %q", tt.expectedProject, projectName)
			}
		})
	}
}

func TestExtractConversationInfo_WithSummaryFirst(t *testing.T) {
	tests := []struct {
		name            string
		jsonlContent    string
		expectedTitle   string
		expectedProject string
	}{
		{
			name: "should extract project name from second message when first is summary",
			jsonlContent: `{"type":"summary","summary":"Git Hooks Configuration for Sound Notification","leafUuid":"4f2e29fe-f53c-4216-a5b5-7e0997265077"}
{"type":"user","message":{"role":"user","content":"Hello"},"cwd":"/Users/annenpolka/junks/cclog","uuid":"test","timestamp":"2025-07-06T05:01:44.663Z"}
{"type":"assistant","message":{"role":"assistant","content":"Hi there"},"cwd":"/Users/annenpolka/junks/cclog","uuid":"test2","timestamp":"2025-07-06T05:01:45.663Z"}`,
			expectedTitle:   "Hello",
			expectedProject: "cclog",
		},
		{
			name: "should extract from first non-summary message",
			jsonlContent: `{"type":"summary","summary":"Test summary","leafUuid":"test-uuid"}
{"type":"system","message":{"role":"system","content":"System message"},"cwd":"/Users/annenpolka/junks/hello-rails","uuid":"system","timestamp":"2025-07-06T05:01:44.663Z"}
{"type":"user","message":{"role":"user","content":"User message"},"cwd":"/Users/annenpolka/junks/hello-rails","uuid":"user","timestamp":"2025-07-06T05:01:45.663Z"}`,
			expectedTitle:   "User message",
			expectedProject: "hello-rails",
		},
		{
			name: "should handle multiple summary messages",
			jsonlContent: `{"type":"summary","summary":"First summary","leafUuid":"test-uuid1"}
{"type":"summary","summary":"Second summary","leafUuid":"test-uuid2"}
{"type":"user","message":{"role":"user","content":"Test message"},"cwd":"/Users/annenpolka/junks/tdt","uuid":"test","timestamp":"2025-07-06T05:01:44.663Z"}`,
			expectedTitle:   "Test message",
			expectedProject: "tdt",
		},
		{
			name: "should handle file with only summary messages",
			jsonlContent: `{"type":"summary","summary":"Only summary","leafUuid":"test-uuid"}
{"type":"summary","summary":"Another summary","leafUuid":"test-uuid2"}`,
			expectedTitle:   "",
			expectedProject: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with test content
			tempFile, err := os.CreateTemp("", "test_*.jsonl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			if _, err := tempFile.WriteString(tt.jsonlContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tempFile.Close()

			// Test extractConversationInfo
			title, projectName := extractConversationInfo(tempFile.Name())

			if title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
			}

			if projectName != tt.expectedProject {
				t.Errorf("Expected project name %q, got %q", tt.expectedProject, projectName)
			}
		})
	}
}
