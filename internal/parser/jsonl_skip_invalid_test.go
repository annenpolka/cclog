package parser

import (
    "path/filepath"
    "testing"
    "os"

    "github.com/annenpolka/cclog/internal/testutil"
)

// Red: invalid JSONL lines should be skipped, not fail parsing.
func TestParseJSONLFileSkipsInvalidLines(t *testing.T) {
    tmpFile := filepath.Join(t.TempDir(), "mixed.jsonl")

    valid1 := `{"type":"user","message":{"role":"user","content":"ok1"},"uuid":"u1","timestamp":"2025-07-06T05:01:44.663Z"}`
    invalid := `not a json line at all`
    valid2 := `{"type":"assistant","message":{"role":"assistant","content":"ok2"},"uuid":"u2","timestamp":"2025-07-06T05:01:45.663Z"}`

    content := valid1 + "\n" + invalid + "\n" + valid2 + "\n"
    if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
        t.Fatalf("failed to write temp jsonl: %v", err)
    }

    log, err := ParseJSONLFile(tmpFile)
    if err != nil {
        t.Fatalf("parser should not error on invalid lines, got: %v", err)
    }

    testutil.Diff(t, 2, len(log.Messages))
    testutil.Diff(t, "user", log.Messages[0].Type)
    testutil.Diff(t, "assistant", log.Messages[1].Type)
}

