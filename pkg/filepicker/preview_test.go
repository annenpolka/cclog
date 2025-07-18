package filepicker

import (
	tea "github.com/charmbracelet/bubbletea"
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
		name          string
		jsonlPath     string
		shouldError   bool
		expectedEmpty bool
	}{
		{
			name:          "Valid JSONL file",
			jsonlPath:     "../../testdata/sample.jsonl",
			shouldError:   false,
			expectedEmpty: false,
		},
		{
			name:          "Non-existent file",
			jsonlPath:     "non-existent-file.jsonl",
			shouldError:   true,
			expectedEmpty: true,
		},
		{
			name:          "Empty path",
			jsonlPath:     "",
			shouldError:   false,
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := GeneratePreview(tt.jsonlPath, true)

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

func TestPreviewModel_KeyBindings_GoToTop(t *testing.T) {
	preview := NewPreviewModel()

	// Set some content and scroll position
	cmd := preview.SetContent("# Test Content\n\nLine 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6")
	if cmd != nil {
		// Execute the command to load content
		_ = cmd()
	}
	preview.SetSize(80, 10)

	// Simulate scrolling down first (so we have somewhere to scroll back to)
	preview.markdownBubble.Viewport.ScrollDown(5)
	initialOffset := preview.markdownBubble.Viewport.YOffset

	// Simulate 'g' key press
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	preview.Update(keyMsg)

	// Check that we're now at the top
	if preview.markdownBubble.Viewport.YOffset != 0 {
		t.Errorf("After 'g' key press, should be at top (YOffset=0), got YOffset=%d, initial was %d", preview.markdownBubble.Viewport.YOffset, initialOffset)
	}
}

func TestPreviewModel_KeyBindings_GoToBottom(t *testing.T) {
	preview := NewPreviewModel()

	// Set some content that will be longer than the viewport
	longContent := "# Test Content\n\n"
	for i := 0; i < 20; i++ {
		longContent += "Line " + string(rune('A'+i)) + "\n"
	}
	cmd := preview.SetContent(longContent)
	if cmd != nil {
		// Execute the command to load content
		_ = cmd()
	}
	preview.SetSize(80, 10)

	// Initially should be at top
	if preview.markdownBubble.Viewport.YOffset != 0 {
		t.Errorf("Should start at top")
	}

	// Simulate 'G' key press (shift+g)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	preview.Update(keyMsg)

	// Check that we're now at the bottom
	// The bottom position should be greater than 0 for content longer than viewport
	finalOffset := preview.markdownBubble.Viewport.YOffset
	totalLines := preview.markdownBubble.Viewport.TotalLineCount()
	height := preview.markdownBubble.Viewport.Height

	// For content that's longer than viewport, we should have scrolled down
	if totalLines > height && finalOffset == 0 {
		t.Errorf("After 'G' key press, should be at bottom (YOffset>0), got YOffset=%d, totalLines=%d, height=%d", finalOffset, totalLines, height)
	}
}
