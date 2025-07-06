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

	if len(log.Messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(log.Messages))
	}

	// Test first message
	firstMsg := log.Messages[0]
	if firstMsg.Type != "user" {
		t.Errorf("Expected first message type 'user', got '%s'", firstMsg.Type)
	}

	if firstMsg.SessionID != "41eb70c6-2cac-4420-834b-ceaea98a7494" {
		t.Errorf("Expected sessionId '41eb70c6-2cac-4420-834b-ceaea98a7494', got '%s'", firstMsg.SessionID)
	}

	// Test second message
	secondMsg := log.Messages[1]
	if secondMsg.Type != "assistant" {
		t.Errorf("Expected second message type 'assistant', got '%s'", secondMsg.Type)
	}

	if secondMsg.ParentUUID == nil || *secondMsg.ParentUUID != "ccd7ef0b-5e81-4881-bda9-d55a7131ca63" {
		t.Errorf("Expected parentUUID 'ccd7ef0b-5e81-4881-bda9-d55a7131ca63', got %v", secondMsg.ParentUUID)
	}

	// Test summary message
	summaryMsg := log.Messages[2]
	if summaryMsg.Type != "summary" {
		t.Errorf("Expected third message type 'summary', got '%s'", summaryMsg.Type)
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

	if len(logs[0].Messages) != 3 {
		t.Errorf("Expected 3 messages in first log, got %d", len(logs[0].Messages))
	}
}