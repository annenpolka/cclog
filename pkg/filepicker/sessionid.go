package filepicker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/annenpolka/cclog/pkg/types"
)

// extractSessionID extracts the sessionId from the first valid message in a JSONL file
func extractSessionID(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Expand buffer size to handle large JSONL lines (up to 1MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines
		if line == "" {
			continue
		}

		var message types.Message
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			return "", fmt.Errorf("failed to parse JSONL line: %w", err)
		}

		// Check if sessionId is present and not empty
		if message.SessionID == "" {
			return "", fmt.Errorf("sessionId is empty or missing")
		}

		return message.SessionID, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return "", fmt.Errorf("no valid messages found in file")
}