package filepicker

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
	"github.com/annenpolka/cclog/internal/formatter"
	"github.com/annenpolka/cclog/internal/parser"
	"github.com/annenpolka/cclog/pkg/types"
)

type Model struct {
	dir             string
	files           []FileInfo
	cursor          int
	selected        string
	recursive       bool
	maxDisplayFiles int
	scrollOffset    int
	terminalWidth   int
	terminalHeight  int
	useCompactLayout bool
	contentAlignment string
	maxTitleChars   int
	preview         *PreviewModel
	enableFiltering bool
}

func NewModel(dir string, recursive bool) Model {
	return Model{
		dir:              dir,
		files:            []FileInfo{},
		cursor:           0,
		recursive:        recursive,
		maxDisplayFiles:  20, // Default limit
		scrollOffset:     0,
		terminalWidth:    80, // Default terminal width
		terminalHeight:   24, // Default terminal height
		useCompactLayout: false, // Default to full layout
		contentAlignment: "left", // Default alignment
		maxTitleChars:    40, // Default title character limit
		preview:          NewPreviewModel(),
		enableFiltering:  true, // Default to filtering enabled
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadFiles(m.dir, m.recursive),
		GetInitialWindowSize(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	// Update preview
	m.preview, cmd = m.preview.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.updateDisplaySettings()
		// Update preview size
		m.updatePreviewSize()
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "p":
			// Toggle preview
			m.preview.SetVisible(!m.preview.IsVisible())
			// Update preview content if visible
			if m.preview.IsVisible() {
				if cmd := m.updatePreviewContent(); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
			return m, tea.Batch(cmds...)
		case "s":
			// Toggle filtering
			m.enableFiltering = !m.enableFiltering
			// Update preview content with new filtering state
			if m.preview.IsVisible() {
				if cmd := m.updatePreviewContent(); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
			return m, tea.Batch(cmds...)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Scroll up if cursor goes above visible range
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
				// Update preview if visible
				if m.preview.IsVisible() {
					if cmd := m.updatePreviewContent(); cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				// Scroll down if cursor goes below visible range
				if m.cursor >= m.scrollOffset+m.maxDisplayFiles {
					m.scrollOffset = m.cursor - m.maxDisplayFiles + 1
				}
				// Update preview if visible
				if m.preview.IsVisible() {
					if cmd := m.updatePreviewContent(); cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
			}
		case "enter":
			if len(m.files) > 0 {
				selectedItem := m.files[m.cursor]
				if selectedItem.IsDir {
					// Navigate into directory
					m.dir = selectedItem.Path
					m.cursor = 0
					m.scrollOffset = 0
					return m, loadFiles(m.dir, m.recursive)
				} else {
					// Convert to markdown and open in editor with current filtering state
					return m, convertAndOpenInEditor(selectedItem.Path, m.enableFiltering)
				}
			}
		}
	case filesLoadedMsg:
		m.files = msg.files
		// Reset cursor and scroll when loading new files
		if m.cursor >= len(m.files) {
			m.cursor = 0
		}
		m.scrollOffset = 0
		// Initialize preview size and content if visible
		if m.preview.IsVisible() {
			m.updatePreviewSize()
			if cmd := m.updatePreviewContent(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var s strings.Builder
	
	// Show current directory with mode indicator
	modeStr := ""
	if m.recursive {
		modeStr = " [RECURSIVE]"
	}
	if m.enableFiltering {
		modeStr += " [FILTERED]"
	} else {
		modeStr += " [UNFILTERED]"
	}
	
	// Truncate directory path for narrow terminals
	dirPath := m.dir + modeStr
	if m.terminalWidth > 0 && len(dirPath) > m.terminalWidth-4 { // Reserve space for emoji and spaces
		availableWidth := m.terminalWidth - 7 // "📁 " + "..."
		if availableWidth > 0 {
			dirPath = types.TruncateTitleWithWidth(dirPath, availableWidth)
		}
	}
	
	s.WriteString("📁 " + dirPath + "\n\n")
	
	// Calculate available space for file list
	listHeight := m.terminalHeight - 5 // Reserve space for header and help
	if m.preview.IsVisible() {
		listHeight = listHeight / 2 // Split screen when preview is visible
	}
	
	// Adjust maxDisplayFiles based on available space
	originalMaxDisplay := m.maxDisplayFiles
	if listHeight > 0 && listHeight < m.maxDisplayFiles {
		m.maxDisplayFiles = listHeight
	}
	
	// Calculate display range with scrolling
	totalFiles := len(m.files)
	displayStart := m.scrollOffset
	displayEnd := m.scrollOffset + m.maxDisplayFiles
	
	if displayEnd > totalFiles {
		displayEnd = totalFiles
	}
	
	// Show scroll indicators
	if totalFiles > m.maxDisplayFiles {
		// Removed "more above" display
	}
	
	// Show files list with scrolling
	for i := displayStart; i < displayEnd; i++ {
		file := m.files[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		
		// Get base title and apply responsive formatting
		title := file.Title()
		
		// Calculate available width for content
		prefixWidth := 3 // cursor + spaces
		availableWidth := m.terminalWidth - prefixWidth
		
		// Create responsive content line
		displayLine := m.formatResponsiveLine(cursor, title, availableWidth)
		s.WriteString(displayLine + "\n")
	}
	
	// Show bottom scroll indicator
	if totalFiles > m.maxDisplayFiles {
		remainingBelow := totalFiles - displayEnd
		if remainingBelow > 0 {
			s.WriteString("↓ " + strconv.Itoa(remainingBelow) + " more below\n")
		}
	}
	
	// Restore original maxDisplayFiles
	m.maxDisplayFiles = originalMaxDisplay
	
	// Show preview if visible
	if m.preview.IsVisible() {
		s.WriteString("\n" + strings.Repeat("─", m.terminalWidth) + "\n")
		s.WriteString(m.preview.View())
	}
	
	// Show help text based on layout
	if !m.useCompactLayout {
		s.WriteString("\n")
		s.WriteString("Controls:\n")
		s.WriteString("  ↑/↓, j/k: Navigate\n")
		s.WriteString("  Enter: Open folder / Open file in editor\n")
		s.WriteString("  p: Toggle preview\n")
		s.WriteString("  s: Toggle filter\n")
		if m.preview.IsVisible() {
			s.WriteString("  d/u: Scroll preview down/up\n")
		}
		s.WriteString("  q: Quit\n")
	} else if m.terminalWidth < 40 {
		// Very narrow: minimal help
		if m.preview.IsVisible() {
			s.WriteString("\nj/k:Nav d/u:Scroll p:Preview s:Filter q:Quit")
		} else {
			s.WriteString("\nj/k:Nav Enter:Open p:Preview s:Filter q:Quit")
		}
	} else {
		// Compact: abbreviated help
		if m.preview.IsVisible() {
			s.WriteString("\nNav:↑↓/jk Open:Enter Preview:p Filter:s Scroll:d/u Quit:q")
		} else {
			s.WriteString("\nNav:↑↓/jk Open:Enter Preview:p Filter:s Quit:q")
		}
	}
	
	return s.String()
}

// GetSelectedFile returns the path of the selected file, or empty string if none selected
func (m Model) GetSelectedFile() string {
	return m.selected
}

type filesLoadedMsg struct {
	files []FileInfo
}

func loadFiles(dir string, recursive bool) tea.Cmd {
	return func() tea.Msg {
		var files []FileInfo
		var err error
		
		if recursive {
			files, err = GetFilesRecursive(dir)
		} else {
			files, err = GetFiles(dir)
		}
		
		if err != nil {
			return filesLoadedMsg{files: []FileInfo{}}
		}
		return filesLoadedMsg{files: files}
	}
}

// openInEditor opens the specified file in the default editor
func openInEditor(filepath string) tea.Cmd {
	return tea.ExecProcess(getEditorCommand(filepath), func(err error) tea.Msg {
		// Return to TUI after editor exits
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
	})
}

// getEditorCommand returns the command to open a file in the default editor
func getEditorCommand(filepath string) *exec.Cmd {
	// Get editor from environment variables
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Default editors to try
		editors := []string{"nano", "vim", "vi", "emacs"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	
	if editor == "" {
		return nil // No editor found
	}
	
	// Create command to open file in editor
	cmd := exec.Command(editor, filepath)
	return cmd
}

// convertAndOpenInEditor converts JSONL file to markdown and opens it in editor
func convertAndOpenInEditor(jsonlPath string, enableFiltering bool) tea.Cmd {
	return func() tea.Msg {
		// Convert JSONL to markdown
		markdownContent, err := convertJSONLToMarkdown(jsonlPath, enableFiltering)
		if err != nil {
			// If conversion fails, fall back to opening original file
			return openInEditor(jsonlPath)()
		}
		
		// Create temporary markdown file
		tempFile, err := os.CreateTemp("", "cclog_*.md")
		if err != nil {
			// If temp file creation fails, fall back to opening original file
			return openInEditor(jsonlPath)()
		}
		
		// Write markdown content to temp file
		if _, err := tempFile.Write([]byte(markdownContent)); err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return openInEditor(jsonlPath)()
		}
		tempFile.Close()
		
		// Open temp file in editor with cleanup
		return openMarkdownInEditor(tempFile.Name())()
	}
}

// convertJSONLToMarkdown converts a JSONL file to markdown format
func convertJSONLToMarkdown(jsonlPath string, enableFiltering bool) (string, error) {
	// Parse JSONL file
	log, err := parser.ParseJSONLFile(jsonlPath)
	if err != nil {
		return "", err
	}
	
	// Apply filtering based on enableFiltering parameter
	filteredLog := formatter.FilterConversationLog(log, enableFiltering)
	
	// Convert to markdown
	markdown := formatter.FormatConversationToMarkdownWithOptions(filteredLog, formatter.FormatOptions{ShowUUID: false})
	
	return markdown, nil
}

// openMarkdownInEditor opens a markdown file in editor and cleans up after
func openMarkdownInEditor(markdownPath string) tea.Cmd {
	return func() tea.Msg {
		cmd := getEditorCommand(markdownPath)
		if cmd == nil {
			os.Remove(markdownPath)
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
		}
		
		// Check if the editor is VS Code or other background editors
		editorName := cmd.Args[0]
		if isBackgroundEditor(editorName) {
			// For background editors, use --wait flag and don't use ExecProcess
			cmd.Args = append(cmd.Args[:1], append([]string{"--wait"}, cmd.Args[1:]...)...)
			
			// Run the command and wait for it to complete
			if err := cmd.Run(); err != nil {
				// If command fails, clean up and return
				os.Remove(markdownPath)
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
			}
			
			// Clean up after editor closes
			os.Remove(markdownPath)
			return tea.Quit
		}
		
		// For terminal editors, use ExecProcess
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			// Clean up temporary file after editor closes
			os.Remove(markdownPath)
			// Return to TUI after editor exits
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
		})()
	}
}

// isBackgroundEditor checks if the editor runs in background
func isBackgroundEditor(editorPath string) bool {
	// Extract basename from path
	editorName := editorPath
	if lastSlash := strings.LastIndex(editorPath, "/"); lastSlash >= 0 {
		editorName = editorPath[lastSlash+1:]
	}
	
	// Known background editors
	backgroundEditors := []string{"code", "codium", "subl", "atom"}
	for _, bg := range backgroundEditors {
		if editorName == bg {
			return true
		}
	}
	return false
}

// GetInitialWindowSize gets the current terminal size
func GetInitialWindowSize() tea.Cmd {
	return func() tea.Msg {
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			// Fallback to default size if unable to get terminal size
			return tea.WindowSizeMsg{Width: 80, Height: 24}
		}
		return tea.WindowSizeMsg{Width: width, Height: height}
	}
}

// updateDisplaySettings adjusts display settings based on terminal size
func (m *Model) updateDisplaySettings() {
	// Determine layout based on width
	m.useCompactLayout = m.terminalWidth < 60
	
	// Calculate dynamic title character limit based on terminal width
	// Base calculation: terminal width - prefix (date/time + cursor + spaces)
	dateTimeWidth := 17 // "2025-01-15 14:30 "
	prefixWidth := 3    // "> "
	marginWidth := 2    // Reduced safety margin
	
	availableForTitle := m.terminalWidth - dateTimeWidth - prefixWidth - marginWidth
	
	// Set minimum and maximum title character limits
	minTitleChars := 20
	maxTitleChars := 200
	
	// Add a boost for wider terminals to show more characters
	if m.terminalWidth > 80 {
		boost := (m.terminalWidth - 80) / 4 // Add extra chars for wide terminals
		availableForTitle += boost
	}
	
	if availableForTitle < minTitleChars {
		m.maxTitleChars = minTitleChars
	} else if availableForTitle > maxTitleChars {
		m.maxTitleChars = maxTitleChars
	} else {
		m.maxTitleChars = availableForTitle
	}
	
	// Keep maxDisplayFiles at default value - no dynamic adjustment
}

// formatResponsiveLine creates a responsive content line that adapts to terminal width
func (m Model) formatResponsiveLine(cursor, title string, availableWidth int) string {
	if availableWidth <= 0 {
		return cursor + " " + title
	}
	
	// Use dynamic title character limit instead of fixed truncation
	formattedTitle := types.TruncateTitleWithWidth(title, m.maxTitleChars)
	
	// Create the display line
	line := cursor + " " + formattedTitle
	
	// Final safety check: ensure line doesn't exceed terminal width
	finalRunes := []rune(line)
	if len(finalRunes) > m.terminalWidth && m.terminalWidth > 0 {
		line = string(finalRunes[:m.terminalWidth])
	}
	
	return line
}


// updatePreviewSize adjusts the preview size based on terminal dimensions
func (m *Model) updatePreviewSize() {
	if m.preview == nil {
		return
	}
	
	previewWidth := m.terminalWidth - 4 // Account for borders
	previewHeight := (m.terminalHeight / 2) - 3 // Split screen, account for borders
	
	if previewWidth < 0 {
		previewWidth = 0
	}
	if previewHeight < 0 {
		previewHeight = 0
	}
	
	m.preview.SetSize(previewWidth, previewHeight)
}

// updatePreviewContent updates the preview content based on current selection
func (m *Model) updatePreviewContent() tea.Cmd {
	if m.preview == nil || !m.preview.IsVisible() || len(m.files) == 0 {
		return nil
	}
	
	selectedFile := m.files[m.cursor]
	if selectedFile.IsDir {
		// Clear preview for directories
		return m.preview.SetContent("")
	}
	
	// Generate preview for JSONL files
	if strings.HasSuffix(selectedFile.Path, ".jsonl") {
		content, err := GeneratePreview(selectedFile.Path, m.enableFiltering)
		if err != nil {
			return m.preview.SetContent("Error generating preview: " + err.Error())
		} else {
			return m.preview.SetContent(content)
		}
	} else {
		return m.preview.SetContent("Preview not available for this file type")
	}
}

