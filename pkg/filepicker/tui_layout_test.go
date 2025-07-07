package filepicker

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelUpdatePreviewSize(t *testing.T) {
	model := NewModel("/tmp", false)
	
	tests := []struct {
		name             string
		terminalWidth    int
		terminalHeight   int
		expectedPreviewWidth int
		expectedPreviewHeight int
	}{
		{
			name:             "Standard terminal size",
			terminalWidth:    80,
			terminalHeight:   40,
			expectedPreviewWidth: 76, // 80 - 4
			expectedPreviewHeight: 27, // (40 - 6) * 0.8 = 27.2 -> 27
		},
		{
			name:             "Large terminal",
			terminalWidth:    120,
			terminalHeight:   60,
			expectedPreviewWidth: 116, // 120 - 4
			expectedPreviewHeight: 43, // (60 - 6) * 0.8 = 43.2 -> 43
		},
		{
			name:             "Small terminal",
			terminalWidth:    40,
			terminalHeight:   20,
			expectedPreviewWidth: 36, // 40 - 4
			expectedPreviewHeight: 11, // (20 - 6) * 0.8 = 11.2 -> 11
		},
		{
			name:             "Very small terminal",
			terminalWidth:    10,
			terminalHeight:   8,
			expectedPreviewWidth: 6, // 10 - 4
			expectedPreviewHeight: 10, // Minimum height constraint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.terminalWidth = tt.terminalWidth
			model.terminalHeight = tt.terminalHeight
			
			model.updatePreviewSize()
			
			width, height := model.preview.GetSize()
			
			if width != tt.expectedPreviewWidth {
				t.Errorf("updatePreviewSize() width = %d, expected %d", width, tt.expectedPreviewWidth)
			}
			
			if height != tt.expectedPreviewHeight {
				t.Errorf("updatePreviewSize() height = %d, expected %d", height, tt.expectedPreviewHeight)
			}
		})
	}
}

func TestModelDynamicLayoutAdjustment(t *testing.T) {
	model := NewModel("/tmp", false)
	
	tests := []struct {
		name           string
		terminalWidth  int
		terminalHeight int
		splitRatio     float64
		expectedListHeight int
		expectedPreviewHeight int
	}{
		{
			name:           "50/50 split with medium terminal",
			terminalWidth:  80,
			terminalHeight: 40,
			splitRatio:     0.5,
			expectedListHeight: 17, // (40 - 6) / 2
			expectedPreviewHeight: 17,
		},
		{
			name:           "30/70 split favoring preview",
			terminalWidth:  80,
			terminalHeight: 60,
			splitRatio:     0.7,
			expectedListHeight: 17, // 60 - 6 - 37 = 17
			expectedPreviewHeight: 37, // (60 - 6) * 0.7 = 37.8 -> 37
		},
		{
			name:           "70/30 split favoring list",
			terminalWidth:  80,
			terminalHeight: 40,
			splitRatio:     0.3,
			expectedListHeight: 24, // (40 - 6) * 0.7
			expectedPreviewHeight: 10, // (40 - 6) * 0.3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.terminalWidth = tt.terminalWidth
			model.terminalHeight = tt.terminalHeight
			
			model.updateDynamicLayout(tt.splitRatio)
			
			listHeight := model.getListHeight()
			_, previewHeight := model.preview.GetSize()
			
			if listHeight != tt.expectedListHeight {
				t.Errorf("updateDynamicLayout() list height = %d, expected %d", listHeight, tt.expectedListHeight)
			}
			
			if previewHeight != tt.expectedPreviewHeight {
				t.Errorf("updateDynamicLayout() preview height = %d, expected %d", previewHeight, tt.expectedPreviewHeight)
			}
		})
	}
}

func TestModelWindowSizeMessage(t *testing.T) {
	model := NewModel("/tmp", false)
	
	// Test window size update
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	
	m := updatedModel.(Model)
	
	if m.terminalWidth != 100 {
		t.Errorf("WindowSizeMsg update: width = %d, expected %d", m.terminalWidth, 100)
	}
	
	if m.terminalHeight != 50 {
		t.Errorf("WindowSizeMsg update: height = %d, expected %d", m.terminalHeight, 50)
	}
	
	// Check if preview size was updated
	width, height := m.preview.GetSize()
	expectedWidth := 96 // 100 - 4
	expectedHeight := 35 // (50 - 6) * 0.8 = 35.2 -> 35
	
	if width != expectedWidth {
		t.Errorf("WindowSizeMsg preview width = %d, expected %d", width, expectedWidth)
	}
	
	if height != expectedHeight {
		t.Errorf("WindowSizeMsg preview height = %d, expected %d", height, expectedHeight)
	}
}

func TestModelResponsiveLayoutThresholds(t *testing.T) {
	model := NewModel("/tmp", false)
	
	tests := []struct {
		name           string
		terminalWidth  int
		terminalHeight int
		expectedCompact bool
		expectedMaxTitleChars int
	}{
		{
			name:           "Large terminal - full layout",
			terminalWidth:  120,
			terminalHeight: 60,
			expectedCompact: false,
			expectedMaxTitleChars: 108, // 120 - 17 - 3 - 2 + (120-80)/4 = 108
		},
		{
			name:           "Medium terminal - full layout",
			terminalWidth:  80,
			terminalHeight: 40,
			expectedCompact: false,
			expectedMaxTitleChars: 58, // 80 - 17 - 3 - 2 = 58
		},
		{
			name:           "Small terminal - compact layout",
			terminalWidth:  50,
			terminalHeight: 25,
			expectedCompact: true,
			expectedMaxTitleChars: 28, // 50 - 17 - 3 - 2 = 28
		},
		{
			name:           "Very small terminal - compact layout",
			terminalWidth:  30,
			terminalHeight: 15,
			expectedCompact: true,
			expectedMaxTitleChars: 20, // Minimum threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.terminalWidth = tt.terminalWidth
			model.terminalHeight = tt.terminalHeight
			
			model.updateDisplaySettings()
			
			if model.useCompactLayout != tt.expectedCompact {
				t.Errorf("updateDisplaySettings() compact layout = %v, expected %v", model.useCompactLayout, tt.expectedCompact)
			}
			
			if model.maxTitleChars != tt.expectedMaxTitleChars {
				t.Errorf("updateDisplaySettings() max title chars = %d, expected %d", model.maxTitleChars, tt.expectedMaxTitleChars)
			}
		})
	}
}

func TestModelMinimumSizeConstraints(t *testing.T) {
	model := NewModel("/tmp", false)
	
	// Test with extremely small terminal
	model.terminalWidth = 5
	model.terminalHeight = 5
	
	model.updatePreviewSize()
	
	width, height := model.preview.GetSize()
	
	// Should handle negative sizes gracefully
	if width < 0 {
		t.Errorf("Preview width should not be negative: %d", width)
	}
	
	if height < 0 {
		t.Errorf("Preview height should not be negative: %d", height)
	}
}

