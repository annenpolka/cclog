package filepicker

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/annenpolka/cclog/internal/parser"
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

// generateResumeCommandWithDirectoryChange generates the claude resume command string with directory change
func generateResumeCommandWithDirectoryChange(filePath string, dangerous bool) (string, error) {
	sessionId, err := extractSessionID(filePath)
	if err != nil {
		return "", err
	}

	// Get the directory containing the file
	dir := filepath.Dir(filePath)
	
	// Quote the directory path if it contains spaces
	quotedDir := dir
	if strings.Contains(dir, " ") {
		quotedDir = fmt.Sprintf(`"%s"`, dir)
	}

	// Generate the resume command
	var resumeCmd string
	if dangerous {
		resumeCmd = fmt.Sprintf("claude -r %s --dangerously-skip-permissions", sessionId)
	} else {
		resumeCmd = fmt.Sprintf("claude -r %s", sessionId)
	}

	// Combine cd command with resume command
	return fmt.Sprintf("cd %s && %s", quotedDir, resumeCmd), nil
}

// extractCWDFromJSONL extracts CWD from JSONL file
func extractCWDFromJSONL(filePath string) (string, error) {
	// Parse the JSONL file to get the first message's CWD
	conversationLog, err := parser.ParseJSONLFile(filePath)
	if err != nil {
		return "", err
	}

	for _, message := range conversationLog.Messages {
		if message.CWD != "" {
			return message.CWD, nil
		}
	}

	return "", fmt.Errorf("no CWD found in file %s", filePath)
}

// generateResumeCommandWithCWDChange generates the claude resume command string with CWD directory change
func generateResumeCommandWithCWDChange(filePath string, dangerous bool) (string, error) {
	sessionId, err := extractSessionID(filePath)
	if err != nil {
		return "", err
	}

	// Extract CWD from JSONL file
	cwd, err := extractCWDFromJSONL(filePath)
	if err != nil {
		return "", err
	}

	// Quote the directory path if it contains spaces
	quotedCWD := cwd
	if strings.Contains(cwd, " ") {
		quotedCWD = fmt.Sprintf(`"%s"`, cwd)
	}

	// Generate the resume command
	var resumeCmd string
	if dangerous {
		resumeCmd = fmt.Sprintf("claude -r %s --dangerously-skip-permissions", sessionId)
	} else {
		resumeCmd = fmt.Sprintf("claude -r %s", sessionId)
	}

	// Combine cd command with resume command
	return fmt.Sprintf("cd %s && %s", quotedCWD, resumeCmd), nil
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

// executeResumeCommandWithCWDChange executes the claude resume command with CWD change in foreground
func executeResumeCommandWithCWDChange(filePath string, dangerous bool) tea.Cmd {
	cmdStr, err := generateResumeCommandWithCWDChange(filePath, dangerous)
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

// executeResumeCommandWithDirectoryChange executes the claude resume command with directory change in foreground
func executeResumeCommandWithDirectoryChange(filePath string, dangerous bool) tea.Cmd {
	cmdStr, err := generateResumeCommandWithDirectoryChange(filePath, dangerous)
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