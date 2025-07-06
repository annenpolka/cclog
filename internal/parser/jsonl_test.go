package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJSONLFile(t *testing.T) {
	testFile := filepath.Join("..", "..", "testdata", "sample.jsonl")
	
	log, err := ParseJSONLFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse JSONL file: %v", err)
	}

	if len(log.Messages) != 11 {
		t.Errorf("Expected 11 messages, got %d", len(log.Messages))
	}

	// Test first message (meta message)
	firstMsg := log.Messages[0]
	if firstMsg.Type != "user" {
		t.Errorf("Expected first message type 'user', got '%s'", firstMsg.Type)
	}

	if firstMsg.SessionID != "41eb70c6-2cac-4420-834b-ceaea98a7494" {
		t.Errorf("Expected sessionId '41eb70c6-2cac-4420-834b-ceaea98a7494', got '%s'", firstMsg.SessionID)
	}

	if !firstMsg.IsMeta {
		t.Errorf("Expected first message to be meta")
	}

	// Test real user message
	userMsg := log.Messages[3]
	if userMsg.Type != "user" {
		t.Errorf("Expected user message type 'user', got '%s'", userMsg.Type)
	}

	// Test assistant message
	assistantMsg := log.Messages[4]
	if assistantMsg.Type != "assistant" {
		t.Errorf("Expected assistant message type 'assistant', got '%s'", assistantMsg.Type)
	}

	// Test summary message
	summaryMsg := log.Messages[9]
	if summaryMsg.Type != "summary" {
		t.Errorf("Expected summary message type 'summary', got '%s'", summaryMsg.Type)
	}

	// Test system message
	systemMsg := log.Messages[10]
	if systemMsg.Type != "system" {
		t.Errorf("Expected system message type 'system', got '%s'", systemMsg.Type)
	}
}

func TestParseJSONLFileNotFound(t *testing.T) {
	_, err := ParseJSONLFile("nonexistent.jsonl")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseJSONLDirectory(t *testing.T) {
	testDir := filepath.Join("..", "..", "testdata")
	
	logs, err := ParseJSONLDirectory(testDir)
	if err != nil {
		t.Fatalf("Failed to parse JSONL directory: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log file, got %d", len(logs))
	}

	if len(logs[0].Messages) != 11 {
		t.Errorf("Expected 11 messages in first log, got %d", len(logs[0].Messages))
	}
}

func TestParseJSONLFileLargeLines(t *testing.T) {
	// Create a temporary file with a large line (80KB)
	tmpFile := filepath.Join(t.TempDir(), "large_line.jsonl")
	
	// Generate a large content string (80KB)
	largeContent := strings.Repeat("A", 80*1024)
	
	// Create a valid JSONL message with large content
	largeMessage := `{"parentUuid":"test-uuid","isSidechain":false,"userType":"external","cwd":"/test","sessionId":"test-session","version":"1.0.0","type":"user","message":{"role":"user","content":"` + largeContent + `"},"uuid":"large-uuid","timestamp":"2025-07-06T05:01:44.663Z"}`
	
	// Write test data
	content := largeMessage + "\n" + `{"type":"user","message":{"role":"user","content":"normal message"},"uuid":"normal-uuid","timestamp":"2025-07-06T05:01:45.663Z"}`
	
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test parsing
	log, err := ParseJSONLFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse JSONL file with large lines: %v", err)
	}
	
	if len(log.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(log.Messages))
	}
	
	// Verify the large message was parsed correctly
	// Message.Message is interface{}, need to cast to map for content access
	if msg, ok := log.Messages[0].Message.(map[string]interface{}); ok {
		if content, ok := msg["content"].(string); ok {
			if len(content) != 80*1024 {
				t.Errorf("Expected large message content length %d, got %d", 80*1024, len(content))
			}
		} else {
			t.Error("Failed to extract content from large message")
		}
	} else {
		t.Error("Failed to cast large message to map")
	}
	
	// Verify the normal message was also parsed
	if msg, ok := log.Messages[1].Message.(map[string]interface{}); ok {
		if content, ok := msg["content"].(string); ok {
			if content != "normal message" {
				t.Errorf("Expected normal message content 'normal message', got '%s'", content)
			}
		} else {
			t.Error("Failed to extract content from normal message")
		}
	} else {
		t.Error("Failed to cast normal message to map")
	}
}

func TestParseJSONLFileEmpty(t *testing.T) {
	// Create a temporary empty file
	tmpFile := filepath.Join(t.TempDir(), "empty.jsonl")
	
	err := os.WriteFile(tmpFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	
	// Test parsing empty file
	log, err := ParseJSONLFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse empty JSONL file: %v", err)
	}
	
	if len(log.Messages) != 0 {
		t.Errorf("Expected 0 messages for empty file, got %d", len(log.Messages))
	}
}

func TestParseJSONLDirectoryWithEmptyFiles(t *testing.T) {
	// Create a temporary directory with mixed files
	tmpDir := t.TempDir()
	
	// Create an empty file
	emptyFile := filepath.Join(tmpDir, "empty.jsonl")
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	
	// Create a file with only whitespace
	whitespaceFile := filepath.Join(tmpDir, "whitespace.jsonl")
	err = os.WriteFile(whitespaceFile, []byte("   \n  \n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create whitespace test file: %v", err)
	}
	
	// Create a valid file
	validFile := filepath.Join(tmpDir, "valid.jsonl")
	validContent := `{"type":"user","message":{"role":"user","content":"test"},"uuid":"test-uuid","timestamp":"2025-07-06T05:01:44.663Z"}`
	err = os.WriteFile(validFile, []byte(validContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid test file: %v", err)
	}
	
	// Test parsing directory
	logs, err := ParseJSONLDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Failed to parse JSONL directory: %v", err)
	}
	
	// Debug: print what files were found
	t.Logf("Found %d log files", len(logs))
	for i, log := range logs {
		t.Logf("Log %d: %s with %d messages", i, log.FilePath, len(log.Messages))
	}
	
	// Should only return the valid file, empty files should be excluded
	if len(logs) != 1 {
		t.Errorf("Expected 1 log file (empty files should be excluded), got %d", len(logs))
	}
	
	if len(logs[0].Messages) != 1 {
		t.Errorf("Expected 1 message in valid log, got %d", len(logs[0].Messages))
	}
}