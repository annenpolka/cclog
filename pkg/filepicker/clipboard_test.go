package filepicker

import (
	"testing"
)

func TestCopySessionID(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "valid jsonl file",
			filePath: "../../testdata/sample.jsonl",
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			filePath: "non-existent.jsonl",
			wantErr:  false, // extractSessionID doesn't check file existence, only filename format
		},
		{
			name:     "empty file path",
			filePath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the copySessionID command
			cmd := copySessionID(tt.filePath)
			msg := cmd()

			// Check if the result is of the expected type
			result, ok := msg.(copySessionIDMsg)
			if !ok {
				t.Errorf("Expected copySessionIDMsg, got %T", msg)
				return
			}

			// Check if error expectation matches
			if tt.wantErr && result.error == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && result.error != nil {
				t.Errorf("Expected no error but got: %v", result.error)
			}
		})
	}
}

func TestCopySessionIDIntegration(t *testing.T) {
	// This test verifies the integration with the actual clipboard library
	// It should fail initially with the current golang.design/x/clipboard
	// and pass after switching to atotto/clipboard
	
	filePath := "../../testdata/sample.jsonl"
	
	// Execute the copySessionID command
	cmd := copySessionID(filePath)
	msg := cmd()
	
	// Check if the result is of the expected type
	result, ok := msg.(copySessionIDMsg)
	if !ok {
		t.Errorf("Expected copySessionIDMsg, got %T", msg)
		return
	}
	
	// For a valid file, we should get a successful result
	if result.error != nil {
		t.Errorf("Expected no error but got: %v", result.error)
	}
	
	if !result.success {
		t.Errorf("Expected success=true but got success=false")
	}
}

func TestCopySessionIDErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "invalid file extension",
			filePath: "test.txt",
			wantErr:  true,
		},
		{
			name:     "file without extension",
			filePath: "test",
			wantErr:  true,
		},
		{
			name:     "only extension",
			filePath: ".jsonl",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the copySessionID command
			cmd := copySessionID(tt.filePath)
			msg := cmd()

			// Check if the result is of the expected type
			result, ok := msg.(copySessionIDMsg)
			if !ok {
				t.Errorf("Expected copySessionIDMsg, got %T", msg)
				return
			}

			// Check if error expectation matches
			if tt.wantErr && result.error == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && result.error != nil {
				t.Errorf("Expected no error but got: %v", result.error)
			}
		})
	}
}