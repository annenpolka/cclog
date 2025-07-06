package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/annenpolka/cclog/pkg/filepicker"
)

// RunTUI starts the TUI file picker and returns the selected file
func RunTUI(config Config) (string, error) {
	// Create and run the TUI model
	model := filepicker.NewModel(config.InputPath, config.Recursive)
	program := tea.NewProgram(model)
	
	finalModel, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("TUI error: %w", err)
	}
	
	// Get the selected file
	if m, ok := finalModel.(filepicker.Model); ok {
		selectedFile := m.GetSelectedFile()
		if selectedFile == "" {
			return "", nil // User cancelled, not an error
		}
		return selectedFile, nil
	}
	
	return "", fmt.Errorf("unexpected model type")
}