package filepicker

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// generateResumeCommand generates the claude resume command string
func generateResumeCommand(filePath string, dangerous bool) (string, error) {
	sessionId, err := extractSessionID(filePath)
	if err != nil {
		return "", err
	}

	if dangerous {
		return fmt.Sprintf("claude -r %s --dangerously-skip-permissions", sessionId), nil
	}
	return fmt.Sprintf("claude -r %s", sessionId), nil
}

// resumeMsg represents the result of executing a resume command
type resumeMsg struct {
	success bool
	error   error
}

// executeResumeCommand executes the claude resume command in foreground
func executeResumeCommand(filePath string, dangerous bool) tea.Cmd {
	cmdStr, err := generateResumeCommand(filePath, dangerous)
	if err != nil {
		return func() tea.Msg {
			return resumeMsg{
				success: false,
				error:   err,
			}
		}
	}

	// Execute the command in foreground using tea.ExecProcess
	cmd := exec.Command("sh", "-c", cmdStr)
	
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return resumeMsg{
				success: false,
				error:   fmt.Errorf("failed to execute command '%s': %w", cmdStr, err),
			}
		}
		
		return resumeMsg{
			success: true,
			error:   nil,
		}
	})
}