package filepicker

import (
	"cclog/internal/formatter"
	"cclog/internal/parser"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PreviewModel struct {
	viewport viewport.Model
	content  string
	visible  bool
	width    int
	height   int
}

func NewPreviewModel() *PreviewModel {
	vp := viewport.New(0, 0)
	return &PreviewModel{
		viewport: vp,
		content:  "",
		visible:  true,
		width:    0,
		height:   0,
	}
}

func (p *PreviewModel) SetContent(content string) {
	p.content = content
	p.viewport.SetContent(content)
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
	p.viewport.Width = width
	p.viewport.Height = height
}

func (p *PreviewModel) GetSize() (int, int) {
	return p.width, p.height
}

func (p *PreviewModel) Update(msg tea.Msg) (*PreviewModel, tea.Cmd) {
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return p, cmd
}

func (p *PreviewModel) View() string {
	if !p.visible {
		return ""
	}
	
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1)
	
	if p.content == "" {
		return style.Render("No preview available")
	}
	
	return style.Render(p.viewport.View())
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