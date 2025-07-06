package filepicker

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"cclog/internal/formatter"
	"cclog/internal/parser"
)

type Model struct {
	dir             string
	files           []FileInfo
	cursor          int
	selected        string
	recursive       bool
	maxDisplayFiles int
	scrollOffset    int
}

func NewModel(dir string, recursive bool) Model {
	return Model{
		dir:             dir,
		files:           []FileInfo{},
		cursor:          0,
		recursive:       recursive,
		maxDisplayFiles: 20, // Default limit
		scrollOffset:    0,
	}
}

func (m Model) Init() tea.Cmd {
	return loadFiles(m.dir, m.recursive)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Scroll up if cursor goes above visible range
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				// Scroll down if cursor goes below visible range
				if m.cursor >= m.scrollOffset+m.maxDisplayFiles {
					m.scrollOffset = m.cursor - m.maxDisplayFiles + 1
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
					// Convert to markdown and open in editor
					return m, convertAndOpenInEditor(selectedItem.Path)
				}
			}
		case " ":
			// Space key: select file only (don't navigate directories)
			if len(m.files) > 0 {
				selectedItem := m.files[m.cursor]
				if !selectedItem.IsDir {
					// Select file and quit
					m.selected = selectedItem.Path
					return m, tea.Quit
				}
				// Do nothing for directories with space key
			}
		}
	case filesLoadedMsg:
		m.files = msg.files
		// Reset cursor and scroll when loading new files
		if m.cursor >= len(m.files) {
			m.cursor = 0
		}
		m.scrollOffset = 0
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder
	
	// Show current directory with mode indicator
	modeStr := ""
	if m.recursive {
		modeStr = " [RECURSIVE]"
	}
	s.WriteString("📁 " + m.dir + modeStr + "\n\n")
	
	// Calculate display range with scrolling
	totalFiles := len(m.files)
	displayStart := m.scrollOffset
	displayEnd := m.scrollOffset + m.maxDisplayFiles
	
	if displayEnd > totalFiles {
		displayEnd = totalFiles
	}
	
	// Show scroll indicators
	if totalFiles > m.maxDisplayFiles {
		if m.scrollOffset > 0 {
			s.WriteString("↑ " + strconv.Itoa(m.scrollOffset) + " more above\n")
		}
	}
	
	// Show files list with scrolling
	for i := displayStart; i < displayEnd; i++ {
		file := m.files[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		displayLine := cursor + " " + file.Title()
		if desc := file.Description(); desc != "" {
			displayLine += " - " + desc
		}
		s.WriteString(displayLine + "\n")
	}
	
	// Show bottom scroll indicator
	if totalFiles > m.maxDisplayFiles {
		remainingBelow := totalFiles - displayEnd
		if remainingBelow > 0 {
			s.WriteString("↓ " + strconv.Itoa(remainingBelow) + " more below\n")
		}
	}
	
	// Show help text
	s.WriteString("\n")
	s.WriteString("Controls:\n")
	s.WriteString("  ↑/↓, j/k: Navigate\n")
	s.WriteString("  Enter: Open folder / Open file in editor\n")
	s.WriteString("  Space: Select file only\n")
	s.WriteString("  q: Quit\n")
	
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
func convertAndOpenInEditor(jsonlPath string) tea.Cmd {
	return func() tea.Msg {
		// Convert JSONL to markdown
		markdownContent, err := convertJSONLToMarkdown(jsonlPath)
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
func convertJSONLToMarkdown(jsonlPath string) (string, error) {
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