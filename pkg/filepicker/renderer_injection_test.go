package filepicker

import (
	"testing"
)

// Red: When a renderer is injected, GeneratePreview should delegate to it.
func TestGeneratePreview_UsesInjectedRenderer(t *testing.T) {
	called := false
	SetRenderer(func(path string, includeAll bool) (string, error) {
		called = true
		if path == "" {
			t.Fatalf("expected non-empty path in renderer")
		}
		// includeAll should be the inverse of enableFiltering
		// We won't assert here, just ensure delegation happens
		return "INJECTED", nil
	})
	defer SetRenderer(nil)

	out, err := GeneratePreview("../../testdata/sample.jsonl", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected injected renderer to be called")
	}
	if out != "INJECTED" {
		t.Fatalf("expected delegated content, got: %q", out)
	}
}

