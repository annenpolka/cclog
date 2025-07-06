package filepicker

import (
	"os"
	"testing"
)

func TestPreviewModel_SetContent(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedContent string
	}{
		{
			name:            "Empty content",
			content:         "",
			expectedContent: "",
		},
		{
			name:            "Simple markdown content",
			content:         "# Title\n\nThis is a test.",
			expectedContent: "# Title\n\nThis is a test.",
		},
		{
			name:            "Multi-line content",
			content:         "Line 1\nLine 2\nLine 3",
			expectedContent: "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := NewPreviewModel()
			_ = preview.SetContent(tt.content)
			
			if preview.GetContent() != tt.expectedContent {
				t.Errorf("SetContent() = %v, want %v", preview.GetContent(), tt.expectedContent)
			}
		})
	}
}

func TestPreviewModel_SetVisible(t *testing.T) {
	tests := []struct {
		name    string
		visible bool
	}{
		{
			name:    "Set visible to true",
			visible: true,
		},
		{
			name:    "Set visible to false",
			visible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := NewPreviewModel()
			preview.SetVisible(tt.visible)
			
			if preview.IsVisible() != tt.visible {
				t.Errorf("SetVisible(%v) = %v, want %v", tt.visible, preview.IsVisible(), tt.visible)
			}
		})
	}
}

func TestPreviewModel_SetSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "Set size to 80x24",
			width:  80,
			height: 24,
		},
		{
			name:   "Set size to 120x40",
			width:  120,
			height: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := NewPreviewModel()
			preview.SetSize(tt.width, tt.height)
			
			width, height := preview.GetSize()
			if width != tt.width || height != tt.height {
				t.Errorf("SetSize(%d, %d) = (%d, %d), want (%d, %d)", 
					tt.width, tt.height, width, height, tt.width, tt.height)
			}
		})
	}
}

func TestGeneratePreview(t *testing.T) {
	tests := []struct {
		name         string
		jsonlPath    string
		shouldError  bool
		expectedEmpty bool
	}{
		{
			name:         "Valid JSONL file",
			jsonlPath:    "../../testdata/sample.jsonl",
			shouldError:  false,
			expectedEmpty: false,
		},
		{
			name:         "Non-existent file",
			jsonlPath:    "non-existent-file.jsonl",
			shouldError:  true,
			expectedEmpty: true,
		},
		{
			name:         "Empty path",
			jsonlPath:    "",
			shouldError:  false,
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := GeneratePreview(tt.jsonlPath)
			
			if tt.shouldError && err == nil {
				t.Errorf("GeneratePreview(%s) expected error but got none", tt.jsonlPath)
			}
			
			if !tt.shouldError && err != nil {
				t.Errorf("GeneratePreview(%s) unexpected error: %v", tt.jsonlPath, err)
			}
			
			if tt.expectedEmpty && content != "" {
				t.Errorf("GeneratePreview(%s) expected empty content but got: %s", tt.jsonlPath, content)
			}
			
			if !tt.expectedEmpty && !tt.shouldError && content == "" {
				t.Errorf("GeneratePreview(%s) expected non-empty content but got empty", tt.jsonlPath)
			}
		})
	}
}

func TestPreviewModel_DefaultState(t *testing.T) {
	preview := NewPreviewModel()
	
	if !preview.IsVisible() {
		t.Errorf("NewPreviewModel() should start with visible=true")
	}
	
	if preview.GetContent() != "" {
		t.Errorf("NewPreviewModel() should start with empty content")
	}
	
	width, height := preview.GetSize()
	if width != 0 || height != 0 {
		t.Errorf("NewPreviewModel() should start with size (0, 0), got (%d, %d)", width, height)
	}
}

func TestPreviewModel_Cleanup(t *testing.T) {
	preview := NewPreviewModel()
	
	// Set some content to create temp file
	_ = preview.SetContent("# Test Content\n\nThis is a test.")
	
	// Check that temp file was created
	if preview.tempFile == "" {
		t.Errorf("SetContent should create a temp file")
	}
	
	// Check temp file exists
	if _, err := os.Stat(preview.tempFile); os.IsNotExist(err) {
		t.Errorf("Temp file should exist after SetContent")
	}
	
	// Cleanup should remove temp file
	preview.Cleanup()
	
	// Check temp file is removed
	if preview.tempFile != "" {
		t.Errorf("Cleanup should clear tempFile path")
	}
}