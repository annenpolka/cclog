package filepicker

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
	"golang.org/x/term"
	"github.com/annenpolka/cclog/internal/formatter"
	"github.com/annenpolka/cclog/internal/parser"
	"github.com/annenpolka/cclog/pkg/types"
)

// Define styles for help text and UI elements
var (
	helpKeyStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "241", Dark: "241"})
	helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "239", Dark: "239"})
	helpSeparatorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "237", Dark: "237"})
	
	// File selection and highlighting styles
	selectedFileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).  // Bright white text
		Background(lipgloss.Color("33")).  // Bright blue background
		Bold(true).
		Padding(0, 1) // Horizontal padding for better visibility
	
	// File type specific styles
	normalFileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "250"}) // Adaptive gray
	
	directoryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).  // Bright blue for directories
		Bold(true)
	
	jsonlFileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("148")) // Green for JSONL files
	
	// UI element styles
	cursorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).  // Bright red cursor
		Bold(true)
	
	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).  // Blue header text
		Bold(true)
	
	modeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).  // Yellow mode indicators
		Bold(true)
	
	scrollIndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))  // Subtle gray for scroll hints
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
	// Initialize clipboard
	err := clipboard.Init()
	if err != nil {
		// If clipboard initialization fails, continue without clipboard functionality
		// This prevents the application from crashing on systems without clipboard support
	}
	
	return Model{
		dir:              dir,
		files:            []FileInfo{},
		cursor:           0,
		recursive:        recursive,
		maxDisplayFiles:  10, // Default limit
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
		case "c":
			// Copy sessionId to clipboard
			if len(m.files) > 0 {
				selectedItem := m.files[m.cursor]
				if !selectedItem.IsDir {
					return m, copySessionID(selectedItem.Path)
				}
			}
			return m, tea.Batch(cmds...)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Ensure cursor visibility after movement
				m.ensureCursorVisible()
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
				// Ensure cursor visibility after movement
				m.ensureCursorVisible()
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
	case copySessionIDMsg:
		// Handle clipboard copy result
		// For now, we silently handle success/failure
		// In a more advanced implementation, we could show a status message
		_ = msg
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var s strings.Builder
	
	// Show current directory with mode indicator using colorful styles
	modeStr := ""
	if m.recursive {
		modeStr = " " + modeStyle.Render("[RECURSIVE]")
	}
	if m.enableFiltering {
		modeStr += " " + modeStyle.Render("[FILTERED]")
	} else {
		modeStr += " " + modeStyle.Render("[UNFILTERED]")
	}
	
	// Truncate directory path for narrow terminals
	dirPath := m.dir
	if m.terminalWidth > 0 && len(dirPath) > m.terminalWidth-20 { // Reserve space for emoji, modes, and spaces
		availableWidth := m.terminalWidth - 20 // "📁 " + modes + "..."
		if availableWidth > 0 {
			dirPath = types.TruncateTitleWithWidth(dirPath, availableWidth)
		}
	}
	
	s.WriteString("📁 " + headerStyle.Render(dirPath) + modeStr + "\n\n")
	
	// Calculate available space for file list using dynamic layout
	listHeight := m.getListHeight()
	
	// Adjust maxDisplayFiles based on available space
	originalMaxDisplay := m.maxDisplayFiles
	if listHeight > 0 {
		m.maxDisplayFiles = listHeight
	}
	
	// Ensure cursor is visible with updated display count
	m.ensureCursorVisible()
	
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
	
	// Show files list with scrolling and colorful styling
	for i := displayStart; i < displayEnd; i++ {
		file := m.files[i]
		cursor := " "
		if i == m.cursor {
			cursor = cursorStyle.Render(">")
		}
		
		// Get base title and apply responsive formatting
		title := file.Title()
		
		// Calculate available width for content
		prefixWidth := 3 // cursor + spaces
		availableWidth := m.terminalWidth - prefixWidth
		
		// Truncate title first, then apply colorful styling
		truncatedTitle := types.TruncateTitleWithWidth(title, m.maxTitleChars)
		styledTitle := m.getStyledTitle(truncatedTitle, file.IsDir, i == m.cursor)
		
		// Create responsive content line
		displayLine := m.formatResponsiveColorLine(cursor, styledTitle, availableWidth)
		s.WriteString(displayLine + "\n")
	}
	
	// Show bottom scroll indicator with styling
	if totalFiles > m.maxDisplayFiles {
		remainingBelow := totalFiles - displayEnd
		if remainingBelow > 0 {
			s.WriteString(scrollIndicatorStyle.Render("↓ " + strconv.Itoa(remainingBelow) + " more below") + "\n")
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
		if m.preview.IsVisible() {
			s.WriteString(renderHelp([]helpItem{
				{keys: "↑↓/jk", desc: "move"},
				{keys: "enter", desc: "open"},
				{keys: "p", desc: "preview"},
				{keys: "s", desc: "filter"},
				{keys: "c", desc: "copy sessionId"},
				{keys: "d/u", desc: "scroll"},
				{keys: "g/G", desc: "top/bot"},
				{keys: "q", desc: "quit"},
			}))
		} else {
			s.WriteString(renderHelp([]helpItem{
				{keys: "↑↓/jk", desc: "move"},
				{keys: "enter", desc: "open"},
				{keys: "p", desc: "preview"},
				{keys: "s", desc: "filter"},
				{keys: "c", desc: "copy sessionId"},
				{keys: "q", desc: "quit"},
			}))
		}
	} else if m.terminalWidth < 40 {
		// Very narrow: minimal help
		if m.preview.IsVisible() {
			s.WriteString("\n")
			s.WriteString(renderHelp([]helpItem{
				{keys: "jk", desc: "move"},
				{keys: "du", desc: "scroll"},
				{keys: "gG", desc: "top/bot"},
				{keys: "p", desc: "preview"},
				{keys: "s", desc: "filter"},
				{keys: "c", desc: "copy sessionId"},
				{keys: "q", desc: "quit"},
			}))
		} else {
			s.WriteString("\n")
			s.WriteString(renderHelp([]helpItem{
				{keys: "jk", desc: "move"},
				{keys: "enter", desc: "open"},
				{keys: "p", desc: "preview"},
				{keys: "s", desc: "filter"},
				{keys: "c", desc: "copy sessionId"},
				{keys: "q", desc: "quit"},
			}))
		}
	} else {
		// Compact: abbreviated help
		if m.preview.IsVisible() {
			s.WriteString("\n")
			s.WriteString(renderHelp([]helpItem{
				{keys: "↑↓/jk", desc: "move"},
				{keys: "enter", desc: "open"},
				{keys: "p", desc: "preview"},
				{keys: "s", desc: "filter"},
				{keys: "c", desc: "copy sessionId"},
				{keys: "d/u", desc: "scroll"},
				{keys: "g/G", desc: "top/bot"},
				{keys: "q", desc: "quit"},
			}))
		} else {
			s.WriteString("\n")
			s.WriteString(renderHelp([]helpItem{
				{keys: "↑↓/jk", desc: "move"},
				{keys: "enter", desc: "open"},
				{keys: "p", desc: "preview"},
				{keys: "s", desc: "filter"},
				{keys: "c", desc: "copy sessionId"},
				{keys: "q", desc: "quit"},
			}))
		}
	}
	
	return s.String()
}

// helpItem represents a help text item with keys and description
type helpItem struct {
	keys string
	desc string
}

// renderHelp renders the help text with styling
func renderHelp(items []helpItem) string {
	var parts []string
	for i, item := range items {
		if i > 0 {
			parts = append(parts, helpSeparatorStyle.Render(" "))
		}
		parts = append(parts, helpKeyStyle.Render(item.keys)+helpSeparatorStyle.Render(":")+helpDescStyle.Render(item.desc))
	}
	return strings.Join(parts, "")
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
	markdown := formatter.FormatConversationToMarkdownWithOptions(filteredLog, formatter.FormatOptions{
		ShowUUID:         false,
		ShowPlaceholders: !enableFiltering, // Show placeholders when filtering is disabled (--include-all equivalent)
	})
	
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

// getStyledTitle applies colorful styling to title based on file type and selection
func (m Model) getStyledTitle(title string, isDir bool, isSelected bool) string {
	switch {
	case isSelected:
		// Selected item gets highlight background with high visibility
		return selectedFileStyle.Render(title)
	case isDir:
		// Directory gets distinctive blue color and bold formatting
		return directoryStyle.Render(title)
	case strings.HasSuffix(title, ".jsonl"):
		// JSONL files get green color for easy identification
		return jsonlFileStyle.Render(title)
	default:
		// Regular files get subtle normal color
		return normalFileStyle.Render(title)
	}
}

// formatResponsiveColorLine creates a responsive content line with colorful styling
func (m Model) formatResponsiveColorLine(cursor, styledTitle string, availableWidth int) string {
	if availableWidth <= 0 {
		return cursor + " " + styledTitle
	}
	
	// Create the display line
	line := cursor + " " + styledTitle
	
	// Note: We don't apply additional truncation here as the styling is already applied
	// The truncation should happen before styling in the caller
	
	return line
}


// updatePreviewSize adjusts the preview size based on terminal dimensions
func (m *Model) updatePreviewSize() {
	if m.preview == nil {
		return
	}
	
	previewWidth := m.terminalWidth // Use full terminal width
	if previewWidth < 0 {
		previewWidth = 0
	}
	
	// Use dynamic height calculation
	m.preview.SetDynamicHeight(m.terminalHeight, m.preview.GetSplitRatio(), 10)
	m.preview.SetSize(previewWidth, m.preview.height)
}

// updateDynamicLayout updates the layout based on split ratio
func (m *Model) updateDynamicLayout(splitRatio float64) {
	if m.preview == nil {
		return
	}
	
	previewWidth := m.terminalWidth - 4
	if previewWidth < 0 {
		previewWidth = 0
	}
	
	m.preview.SetDynamicHeight(m.terminalHeight, splitRatio, 10)
	m.preview.SetSize(previewWidth, m.preview.height)
}

// getListHeight returns the height available for the file list
func (m *Model) getListHeight() int {
	if !m.preview.IsVisible() {
		listHeight := m.terminalHeight - 5 // Full height minus header and help
		if listHeight < 1 {
			listHeight = 1 // Ensure minimum height
		}
		return listHeight
	}
	
	_, listHeight := calculatePreviewHeight(m.terminalHeight, m.preview.GetSplitRatio(), 10)
	if listHeight < 1 {
		listHeight = 1 // Ensure minimum height
	}
	return listHeight
}

// ensureCursorVisible ensures the cursor is within the visible range
func (m *Model) ensureCursorVisible() {
	if len(m.files) == 0 {
		return
	}
	
	// Ensure cursor is within bounds
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.files) {
		m.cursor = len(m.files) - 1
	}
	
	// Ensure maxDisplayFiles is positive
	if m.maxDisplayFiles <= 0 {
		m.maxDisplayFiles = 1
	}
	
	// Adjust scroll offset to keep cursor visible
	if m.cursor < m.scrollOffset {
		// Cursor is above visible range, scroll up
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+m.maxDisplayFiles {
		// Cursor is below visible range, scroll down
		m.scrollOffset = m.cursor - m.maxDisplayFiles + 1
	}
	
	// Ensure scroll offset is within bounds
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	maxScrollOffset := len(m.files) - m.maxDisplayFiles
	if maxScrollOffset < 0 {
		maxScrollOffset = 0
	}
	if m.scrollOffset > maxScrollOffset {
		m.scrollOffset = maxScrollOffset
	}
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

// copySessionIDMsg represents the result of copying sessionId to clipboard
type copySessionIDMsg struct {
	success bool
	error   error
}

// copySessionID copies the sessionId from the selected file to clipboard
func copySessionID(filePath string) tea.Cmd {
	return func() tea.Msg {
		sessionId, err := extractSessionID(filePath)
		if err != nil {
			return copySessionIDMsg{
				success: false,
				error:   err,
			}
		}

		clipboard.Write(clipboard.FmtText, []byte(sessionId))

		return copySessionIDMsg{
			success: true,
			error:   nil,
		}
	}
}

