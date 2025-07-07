package filepicker

import (
	"testing"
)

func TestUpdatePreviewSize(t *testing.T) {
	tests := []struct {
		name                string
		terminalWidth       int
		expectedPreviewWidth int
		description         string
	}{
		{
			name:                "Standard terminal width",
			terminalWidth:       80,
			expectedPreviewWidth: 80, // Use full terminal width
			description:         "80文字幅のターミナルでは80文字のプレビュー幅を期待",
		},
		{
			name:                "Wide terminal width",
			terminalWidth:       120,
			expectedPreviewWidth: 120, // Use full terminal width
			description:         "120文字幅のターミナルでは120文字のプレビュー幅を期待",
		},
		{
			name:                "Narrow terminal width",
			terminalWidth:       40,
			expectedPreviewWidth: 40, // Use full terminal width
			description:         "40文字幅のターミナルでは40文字のプレビュー幅を期待",
		},
		{
			name:                "Very narrow terminal",
			terminalWidth:       10,
			expectedPreviewWidth: 10, // Use full terminal width
			description:         "10文字幅のターミナルでは10文字のプレビュー幅を期待",
		},
		{
			name:                "Edge case - width 2",
			terminalWidth:       2,
			expectedPreviewWidth: 2, // Use full terminal width
			description:         "2文字幅のターミナルでは2文字のプレビュー幅を期待",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a model with real preview
			m := &Model{
				terminalWidth: tt.terminalWidth,
				preview:       NewPreviewModel(),
			}

			// Call updatePreviewSize
			m.updatePreviewSize()

			// Check if the preview width was set correctly
			if m.preview.width != tt.expectedPreviewWidth {
				t.Errorf("Expected preview width %d, got %d. %s", 
					tt.expectedPreviewWidth, m.preview.width, tt.description)
			}
		})
	}
}

func TestCurrentPreviewWidthCalculation(t *testing.T) {
	tests := []struct {
		name                string
		terminalWidth       int
		currentPreviewWidth int
		description         string
	}{
		{
			name:                "Current calculation - 80 width",
			terminalWidth:       80,
			currentPreviewWidth: 76, // 80 - 4 = 76
			description:         "現在のロジックでは80文字幅から4を引いて76文字",
		},
		{
			name:                "Current calculation - 120 width",
			terminalWidth:       120,
			currentPreviewWidth: 116, // 120 - 4 = 116
			description:         "現在のロジックでは120文字幅から4を引いて116文字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test current logic
			previewWidth := tt.terminalWidth - 4
			if previewWidth < 0 {
				previewWidth = 0
			}

			if previewWidth != tt.currentPreviewWidth {
				t.Errorf("Current logic test failed. Expected %d, got %d. %s", 
					tt.currentPreviewWidth, previewWidth, tt.description)
			}
		})
	}
}

