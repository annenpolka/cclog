package filepicker

import (
	"testing"
)

func TestCalculatePreviewHeight(t *testing.T) {
	tests := []struct {
		name             string
		terminalHeight   int
		splitRatio       float64
		minHeight        int
		expectedHeight   int
		expectedListHeight int
	}{
		{
			name:             "Standard terminal size with default split",
			terminalHeight:   40,
			splitRatio:       0.5,
			minHeight:        10,
			expectedHeight:   17, // (40 - 6) / 2 = 17
			expectedListHeight: 17,
		},
		{
			name:             "Large terminal with 70% preview",
			terminalHeight:   80,
			splitRatio:       0.7,
			minHeight:        10,
			expectedHeight:   51, // (80 - 6) * 0.7 = 51.8 -> 51
			expectedListHeight: 23, // 80 - 6 - 51 = 23
		},
		{
			name:             "Small terminal with minimum height constraint",
			terminalHeight:   20,
			splitRatio:       0.5,
			minHeight:        10,
			expectedHeight:   10, // Would be 7, but constrained to minHeight
			expectedListHeight: 4, // 20 - 6 - 10 = 4
		},
		{
			name:             "Very small terminal",
			terminalHeight:   15,
			splitRatio:       0.5,
			minHeight:        10,
			expectedHeight:   7, // Adaptive split prioritizes list
			expectedListHeight: 2, // Minimum list height ensured
		},
		{
			name:             "30% preview split",
			terminalHeight:   60,
			splitRatio:       0.3,
			minHeight:        10,
			expectedHeight:   16, // (60 - 6) * 0.3 = 16.2 -> 16
			expectedListHeight: 38, // 60 - 6 - 16 = 38
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			height, listHeight := calculatePreviewHeight(tt.terminalHeight, tt.splitRatio, tt.minHeight)
			
			if height != tt.expectedHeight {
				t.Errorf("calculatePreviewHeight() preview height = %d, expected %d", height, tt.expectedHeight)
			}
			
			if listHeight != tt.expectedListHeight {
				t.Errorf("calculatePreviewHeight() list height = %d, expected %d", listHeight, tt.expectedListHeight)
			}
		})
	}
}

func TestCalculateOptimalSplitRatio(t *testing.T) {
	tests := []struct {
		name           string
		terminalHeight int
		contentLines   int
		expected       float64
	}{
		{
			name:           "Standard content with standard terminal",
			terminalHeight: 40,
			contentLines:   20,
			expected:       0.5, // Default split for standard content
		},
		{
			name:           "Long content needs more space",
			terminalHeight: 60,
			contentLines:   80,
			expected:       0.7, // More space for long content
		},
		{
			name:           "Short content needs less space",
			terminalHeight: 40,
			contentLines:   5,
			expected:       0.3, // Less space for short content
		},
		{
			name:           "Very long content with large terminal",
			terminalHeight: 100,
			contentLines:   200,
			expected:       0.8, // 0.7 + 0.1 for large terminal = 0.8
		},
		{
			name:           "Small terminal with long content",
			terminalHeight: 20,
			contentLines:   50,
			expected:       0.5, // Balanced for small terminal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio := calculateOptimalSplitRatio(tt.terminalHeight, tt.contentLines)
			
			// Use tolerance for floating point comparison
			tolerance := 0.001
			if ratio < tt.expected-tolerance || ratio > tt.expected+tolerance {
				t.Errorf("calculateOptimalSplitRatio() = %f, expected %f", ratio, tt.expected)
			}
		})
	}
}

func TestPreviewModelDynamicHeight(t *testing.T) {
	preview := NewPreviewModel()
	
	// Test setting dynamic height
	preview.SetDynamicHeight(40, 0.6, 10)
	
	_, height := preview.GetSize()
	if height != 20 { // (40 - 6) * 0.6 = 20.4 -> 20
		t.Errorf("SetDynamicHeight() height = %d, expected %d", height, 20)
	}
	
	// Test minimum height constraint on very small screen
	preview.SetDynamicHeight(15, 0.5, 10)
	
	_, height = preview.GetSize()
	if height != 7 { // Adaptive split gives more to list on small screens
		t.Errorf("SetDynamicHeight() with minimum constraint height = %d, expected %d", height, 7)
	}
}

func TestPreviewModelSplitRatioAdjustment(t *testing.T) {
	preview := NewPreviewModel()
	
	// Test initial split ratio (now 80%)
	if preview.GetSplitRatio() != 0.8 {
		t.Errorf("Initial split ratio = %f, expected %f", preview.GetSplitRatio(), 0.8)
	}
	
	// Test adjusting split ratio (already at max)
	preview.AdjustSplitRatio(0.1)
	if preview.GetSplitRatio() != 0.8 {
		t.Errorf("After adjustment split ratio = %f, expected %f", preview.GetSplitRatio(), 0.8)
	}
	
	// Test decreasing split ratio
	preview.AdjustSplitRatio(-0.1)
	tolerance := 0.001
	expected := 0.7
	actual := preview.GetSplitRatio()
	if actual < expected-tolerance || actual > expected+tolerance {
		t.Errorf("After decrease split ratio = %f, expected %f", actual, expected)
	}
	
	// Test minimum constraint
	preview.AdjustSplitRatio(-1.0)
	if preview.GetSplitRatio() != 0.2 { // Should be capped at 0.2
		t.Errorf("Minimum constraint split ratio = %f, expected %f", preview.GetSplitRatio(), 0.2)
	}
}