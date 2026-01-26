package usecase_test

import (
	"strings"
	"testing"

	"github.com/annenpolka/cclog/internal/usecase"
)

// Red: minimal scenario – render markdown from a single JSONL file path.
func TestGenerateMarkdownFromPath_SingleFile_Basic(t *testing.T) {
	input := "../../testdata/sample.jsonl"

	md, err := usecase.GenerateMarkdownFromPath(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if md == "" {
		t.Fatalf("expected non-empty markdown output")
	}

	if !strings.Contains(md, "# Conversation Log") {
		t.Errorf("expected header '# Conversation Log' in output, got: %q", md[:min(120, len(md))])
	}

	if !strings.Contains(md, "**File:** `"+input+"`") {
		t.Errorf("expected file path in output, want %q to appear", "**File:** `"+input+"`")
	}
}

// helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
