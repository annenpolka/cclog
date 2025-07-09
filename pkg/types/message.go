package types

import (
	"time"
)

// Message represents a single message in the JSONL conversation log
type Message struct {
	ParentUUID    *string     `json:"parentUuid"`
	IsSidechain   bool        `json:"isSidechain"`
	UserType      string      `json:"userType"`
	CWD           string      `json:"cwd"`
	SessionID     string      `json:"sessionId"`
	Version       string      `json:"version"`
	Type          string      `json:"type"`
	Message       interface{} `json:"message"`
	IsMeta        bool        `json:"isMeta,omitempty"`
	UUID          string      `json:"uuid"`
	Timestamp     time.Time   `json:"timestamp"`
	RequestID     string      `json:"requestId,omitempty"`
	ToolUseResult interface{} `json:"toolUseResult,omitempty"`
}

// ConversationLog represents a collection of messages from a JSONL file
type ConversationLog struct {
	Messages []Message `json:"messages"`
	FilePath string    `json:"filePath"`
}

// ClaudeMessage represents the structure of Claude's message content
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Summary represents conversation summary information
type Summary struct {
	Type     string `json:"type"`
	Summary  string `json:"summary"`
	LeafUUID string `json:"leafUuid"`
}
