package filepicker

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	dir      string
	files    []FileInfo
	cursor   int
	selected string
}

func NewModel(dir string) Model {
	return Model{
		dir:    dir,
		files:  []FileInfo{},
		cursor: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return loadFiles(m.dir)
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
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.files) > 0 {
				selectedItem := m.files[m.cursor]
				if selectedItem.IsDir {
					// Navigate into directory
					m.dir = selectedItem.Path
					m.cursor = 0
					return m, loadFiles(m.dir)
				} else {
					// Open file in editor
					cmd := getEditorCommand(selectedItem.Path)
					if cmd == nil {
						// No editor found, fall back to selection
						m.selected = selectedItem.Path
						return m, tea.Quit
					}
					return m, openInEditor(selectedItem.Path)
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
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder
	
	// Show current directory
	s.WriteString("📁 " + m.dir + "\n\n")
	
	// Show files list
	for i, file := range m.files {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		s.WriteString(cursor + " " + file.Title() + "\n")
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

func loadFiles(dir string) tea.Cmd {
	return func() tea.Msg {
		files, err := GetFiles(dir)
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