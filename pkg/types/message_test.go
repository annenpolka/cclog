package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageUnmarshal(t *testing.T) {
	// Test data based on actual JSONL structure
	jsonData := `{
		"parentUuid": null,
		"isSidechain": false,
		"userType": "external",
		"cwd": "/Users/annenpolka/junks/cclog",
		"sessionId": "41eb70c6-2cac-4420-834b-ceaea98a7494",
		"version": "1.0.43",
		"type": "user",
		"message": {
			"role": "user",
			"content": "test message"
		},
		"isMeta": true,
		"uuid": "ccd7ef0b-5e81-4881-bda9-d55a7131ca63",
		"timestamp": "2025-07-06T05:01:29.618Z"
	}`

	var msg Message
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if msg.SessionID != "41eb70c6-2cac-4420-834b-ceaea98a7494" {
		t.Errorf("Expected sessionId '41eb70c6-2cac-4420-834b-ceaea98a7494', got '%s'", msg.SessionID)
	}

	if msg.Type != "user" {
		t.Errorf("Expected type 'user', got '%s'", msg.Type)
	}

	if !msg.IsMeta {
		t.Errorf("Expected isMeta to be true")
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2025-07-06T05:01:29.618Z")
	if !msg.Timestamp.Equal(expectedTime) {
		t.Errorf("Expected timestamp %v, got %v", expectedTime, msg.Timestamp)
	}
}

func TestConversationLogCreation(t *testing.T) {
	log := ConversationLog{
		Messages: []Message{},
		FilePath: "/test/path.jsonl",
	}

	if log.FilePath != "/test/path.jsonl" {
		t.Errorf("Expected filepath '/test/path.jsonl', got '%s'", log.FilePath)
	}

	if len(log.Messages) != 0 {
		t.Errorf("Expected empty messages slice, got length %d", len(log.Messages))
	}
}
