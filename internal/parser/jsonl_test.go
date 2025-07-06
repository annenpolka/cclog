package parser

import (
	"path/filepath"
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