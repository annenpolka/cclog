package filepicker

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/annenpolka/cclog/internal/parser"
	tea "github.com/charmbracelet/bubbletea"
)

// execCommand is a variable that can be replaced in tests to mock os/exec.Command
var execCommand = exec.Command

// generateResumeCommand generates the claude resume command and its arguments
func generateResumeCommand(filePath string, dangerous bool) (string, []string, error) {
	sessionId, err := extractSessionID(filePath)
	if err != nil {
		return "", nil, err
	}

	args := []string{"-r", sessionId}
	if dangerous {
		args = append(args, "--dangerously-skip-permissions")
	}
	return "claude", args, nil
}

// generateResumeCommandWithDirectoryChange generates the claude resume command, its arguments, and the directory to execute in
func generateResumeCommandWithDirectoryChange(filePath string, dangerous bool) (string, []string, string, error) {
	sessionId, err := extractSessionID(filePath)
	if err != nil {
		return "", nil, "", err
	}

	dir := filepath.Dir(filePath)

	args := []string{"-r", sessionId}
	if dangerous {
		args = append(args, "--dangerously-skip-permissions")
	}
	return "claude", args, dir, nil
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

// generateResumeCommandWithCWDChange generates the claude resume command, its arguments, and the CWD to execute in
func generateResumeCommandWithCWDChange(filePath string, dangerous bool) (string, []string, string, error) {
	sessionId, err := extractSessionID(filePath)
	if err != nil {
		return "", nil, "", err
	}

	cwd, err := extractCWDFromJSONL(filePath)
	if err != nil {
		return "", nil, "", err
	}

	args := []string{"-r", sessionId}
	if dangerous {
		args = append(args, "--dangerously-skip-permissions")
	}
	return "claude", args, cwd, nil
}

// resumeMsg represents the result of executing a resume command
type resumeMsg struct {
	success bool
	error   error
}

// executeResumeCommand executes the claude resume command in foreground
func executeResumeCommand(filePath string, dangerous bool) tea.Cmd {
	cmdName, cmdArgs, err := generateResumeCommand(filePath, dangerous)
	if err != nil {
		return func() tea.Msg {
			return resumeMsg{
				success: false,
				error:   err,
			}
		}
	}

	cmd := execCommand(cmdName, cmdArgs...)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return resumeMsg{
				success: false,
				error:   fmt.Errorf("failed to execute command '%s %v': %w", cmdName, cmdArgs, err),
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
	cmdName, cmdArgs, cmdDir, err := generateResumeCommandWithCWDChange(filePath, dangerous)
	if err != nil {
		return func() tea.Msg {
			return resumeMsg{
				success: false,
				error:   err,
			}
		}
	}

	cmd := execCommand(cmdName, cmdArgs...)
	cmd.Dir = cmdDir

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return resumeMsg{
				success: false,
				error:   fmt.Errorf("failed to execute command '%s %v' in dir '%s': %w", cmdName, cmdArgs, cmdDir, err),
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
	cmdName, cmdArgs, cmdDir, err := generateResumeCommandWithDirectoryChange(filePath, dangerous)
	if err != nil {
		return func() tea.Msg {
			return resumeMsg{
				success: false,
				error:   err,
			}
		}
	}

	cmd := execCommand(cmdName, cmdArgs...)
	cmd.Dir = cmdDir

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return resumeMsg{
				success: false,
				error:   fmt.Errorf("failed to execute command '%s %v' in dir '%s': %w", cmdName, cmdArgs, cmdDir, err),
			}
		}

		return resumeMsg{
			success: true,
			error:   nil,
		}
	})
}
