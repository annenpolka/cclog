package filepicker

import (
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
				m.selected = m.files[m.cursor].Path
				return m, tea.Quit
			}
		}
	case filesLoadedMsg:
		m.files = msg.files
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder
	s.WriteString("Select a file:\n\n")
	
	for i, file := range m.files {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		s.WriteString(cursor + " " + file.Title() + "\n")
	}
	
	s.WriteString("\nPress q to quit, enter to select")
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