package filepicker

import (
	"github.com/annenpolka/cclog/internal/formatter"
	"github.com/annenpolka/cclog/internal/parser"
	"os"
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

func GeneratePreview(jsonlPath string) (string, error) {
	if jsonlPath == "" {
		return "", nil
	}
	
	// Parse JSONL file
	log, err := parser.ParseJSONLFile(jsonlPath)
	if err != nil {
		return "", err
	}
	
	// Apply filtering (remove system messages)
	filteredLog := formatter.FilterConversationLog(log, true)
	
	// Convert to markdown
	markdown := formatter.FormatConversationToMarkdownWithOptions(filteredLog, formatter.FormatOptions{ShowUUID: false})
	
	return markdown, nil
}