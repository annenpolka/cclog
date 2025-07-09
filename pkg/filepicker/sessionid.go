package filepicker

import (
	"fmt"
	"path/filepath"
	"strings"
)

// extractSessionID extracts the sessionId from the filename by removing the extension
func extractSessionID(filePath string) (string, error) {
	// Get the base filename without directory
	filename := filepath.Base(filePath)

	// Check if file has .jsonl extension
	if !strings.HasSuffix(strings.ToLower(filename), ".jsonl") {
		return "", fmt.Errorf("file %s is not a JSONL file", filename)
	}

	// Remove the .jsonl extension to get sessionId
	sessionId := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Check if sessionId is empty after removing extension
	if sessionId == "" {
		return "", fmt.Errorf("cannot extract sessionId from filename: %s", filename)
	}

	return sessionId, nil
}
