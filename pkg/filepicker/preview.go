package filepicker

import (
	"github.com/annenpolka/cclog/internal/formatter"
	"github.com/annenpolka/cclog/internal/parser"
	"os"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/philistino/teacup/markdown"
)

type PreviewModel struct {
	markdownBubble markdown.Bubble
	content        string
	visible        bool
	width          int
	height         int
	tempFile       string // Store temporary markdown file path
	splitRatio     float64 // Split ratio for preview height (0.2 to 0.8)
	minHeight      int     // Minimum preview height
	maxHeight      int     // Maximum preview height
}

func NewPreviewModel() *PreviewModel {
	borderColor := lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#444444"}
	markdownBubble := markdown.New(true, false, borderColor)
	return &PreviewModel{
		markdownBubble: markdownBubble,
		content:        "",
		visible:        true,
		width:          0,
		height:         0,
		tempFile:       "",
		splitRatio:     0.8, // Default 80% for preview
		minHeight:      10,  // Minimum 10 lines
		maxHeight:      0,   // No maximum by default
	}
}

func (p *PreviewModel) SetContent(content string) tea.Cmd {
	p.content = content
	
	// Clean up previous temp file
	if p.tempFile != "" {
		os.Remove(p.tempFile)
		p.tempFile = ""
	}
	
	if content == "" {
		return nil
	}
	
	// Create temporary markdown file
	tempFile, err := os.CreateTemp("", "cclog_preview_*.md")
	if err != nil {
		return nil
	}
	
	// Write markdown content to temp file
	if _, err := tempFile.Write([]byte(content)); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil
	}
	tempFile.Close()
	
	p.tempFile = tempFile.Name()
	// Reset scroll position to top when loading new content
	p.markdownBubble.GotoTop()
	return p.markdownBubble.SetFileName(p.tempFile)
}

func (p *PreviewModel) GetContent() string {
	return p.content
}

func (p *PreviewModel) SetVisible(visible bool) {
	p.visible = visible
}

func (p *PreviewModel) IsVisible() bool {
	return p.visible
}

func (p *PreviewModel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.markdownBubble.SetSize(width, height)
}

func (p *PreviewModel) GetSize() (int, int) {
	return p.width, p.height
}

// SetDynamicHeight sets the height based on terminal dimensions and split ratio
func (p *PreviewModel) SetDynamicHeight(terminalHeight int, splitRatio float64, minHeight int) {
	p.splitRatio = splitRatio
	p.minHeight = minHeight
	
	height, _ := calculatePreviewHeight(terminalHeight, splitRatio, minHeight)
	p.height = height
	p.markdownBubble.SetSize(p.width, p.height)
}

// GetSplitRatio returns the current split ratio
func (p *PreviewModel) GetSplitRatio() float64 {
	return p.splitRatio
}

// AdjustSplitRatio adjusts the split ratio by the given delta
func (p *PreviewModel) AdjustSplitRatio(delta float64) {
	p.splitRatio += delta
	
	// Constrain to 0.2 - 0.8 range
	if p.splitRatio < 0.2 {
		p.splitRatio = 0.2
	} else if p.splitRatio > 0.8 {
		p.splitRatio = 0.8
	}
}

// Cleanup removes temporary files
func (p *PreviewModel) Cleanup() {
	if p.tempFile != "" {
		os.Remove(p.tempFile)
		p.tempFile = ""
	}
}

func (p *PreviewModel) Update(msg tea.Msg) (*PreviewModel, tea.Cmd) {
	var cmd tea.Cmd
	
	// Handle scroll keys for markdown preview
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d", "pgdn":
			// Scroll down using viewport
			p.markdownBubble.Viewport.ScrollDown(3)
		case "u", "pgup":
			// Scroll up using viewport
			p.markdownBubble.Viewport.ScrollUp(3)
		case "g":
			// Go to top
			p.markdownBubble.GotoTop()
		case "G":
			// Go to bottom by setting YOffset to maximum value
			p.markdownBubble.Viewport.YOffset = p.markdownBubble.Viewport.TotalLineCount() - p.markdownBubble.Viewport.Height
			if p.markdownBubble.Viewport.YOffset < 0 {
				p.markdownBubble.Viewport.YOffset = 0
			}
		}
	}
	
	p.markdownBubble, cmd = p.markdownBubble.Update(msg)
	return p, cmd
}

func (p *PreviewModel) View() string {
	if !p.visible {
		return ""
	}
	
	if p.content == "" {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)
		return style.Render("No preview available")
	}
	
	return p.markdownBubble.View()
}

func GeneratePreview(jsonlPath string, enableFiltering bool) (string, error) {
	if jsonlPath == "" {
		return "", nil
	}
	
	// Parse JSONL file
	log, err := parser.ParseJSONLFile(jsonlPath)
	if err != nil {
		return "", err
	}
	
	// Apply filtering based on enableFiltering parameter
	filteredLog := formatter.FilterConversationLog(log, enableFiltering)
	
	// Convert to markdown
	markdown := formatter.FormatConversationToMarkdownWithOptions(filteredLog, formatter.FormatOptions{
		ShowUUID:         false,
		ShowPlaceholders: !enableFiltering, // Show placeholders when filtering is disabled (--include-all equivalent)
	})
	
	return markdown, nil
}

// calculatePreviewHeight calculates preview and list heights based on terminal dimensions
func calculatePreviewHeight(terminalHeight int, splitRatio float64, minHeight int) (int, int) {
	// Reserve space for header, borders, and help text
	availableHeight := terminalHeight - 6
	
	if availableHeight < minHeight {
		return minHeight, availableHeight - minHeight
	}
	
	previewHeight := int(float64(availableHeight) * splitRatio)
	
	// Apply minimum height constraint
	if previewHeight < minHeight {
		previewHeight = minHeight
	}
	
	listHeight := availableHeight - previewHeight
	
	return previewHeight, listHeight
}

// calculateOptimalSplitRatio determines the optimal split ratio based on content and terminal size
func calculateOptimalSplitRatio(terminalHeight int, contentLines int) float64 {
	// Base split ratio
	baseRatio := 0.5
	
	// Adjust based on content length
	if contentLines > terminalHeight {
		// Long content needs more space
		baseRatio = 0.7
	} else if contentLines < terminalHeight/3 {
		// Short content needs less space
		baseRatio = 0.3
	}
	
	// Adjust based on terminal size
	if terminalHeight > 80 {
		// Large terminal can accommodate more preview
		baseRatio += 0.1
	} else if terminalHeight < 30 {
		// Small terminal needs balanced split
		baseRatio = 0.5
	}
	
	// Constrain to valid range
	if baseRatio < 0.2 {
		baseRatio = 0.2
	} else if baseRatio > 0.8 {
		baseRatio = 0.8
	}
	
	return baseRatio
}

// CountContentLines counts the number of lines in the content
func (p *PreviewModel) CountContentLines() int {
	if p.content == "" {
		return 0
	}
	
	lines := strings.Split(p.content, "\n")
	return len(lines)
}