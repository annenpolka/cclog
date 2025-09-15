package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/annenpolka/cclog/internal/domain"
)

// ParseJSONLFile parses a single JSONL file and returns a ConversationLog
func ParseJSONLFile(filePath string) (*domain.ConversationLog, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var messages []domain.Message
	scanner := bufio.NewScanner(file)
	// Expand buffer size to handle large JSONL lines (up to 1MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

    var msg domain.Message
    if err := json.Unmarshal([]byte(line), &msg); err != nil {
        // Skip malformed JSONL lines instead of failing the entire file.
        // This makes parsing resilient to partial writes or tool errors.
        continue
    }

    messages = append(messages, msg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return &domain.ConversationLog{
		Messages: messages,
		FilePath: filePath,
	}, nil
}

// ParseJSONLDirectory parses all JSONL files in a directory
func ParseJSONLDirectory(dirPath string) ([]*domain.ConversationLog, error) {
	files, err := filepath.Glob(filepath.Join(dirPath, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob JSONL files in %s: %w", dirPath, err)
	}

	var logs []*domain.ConversationLog
	for _, file := range files {
		log, err := ParseJSONLFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", file, err)
		}

		// Skip empty files (files with no messages)
		if len(log.Messages) == 0 {
			continue
		}

		logs = append(logs, log)
	}

	return logs, nil
}
