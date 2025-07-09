package filepicker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResumeCommand(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		dangerous       bool
		expectedCmdName string
		expectedArgs    []string
		expectedErr     bool
		description     string
	}{
		{
			name:            "normal_resume_command",
			filePath:        "/path/to/session-123.jsonl",
			dangerous:       false,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-123"},
			expectedErr:     false,
			description:     "通常のresume実行コマンドが正しく生成される",
		},
		{
			name:            "dangerous_resume_command",
			filePath:        "/path/to/session-456.jsonl",
			dangerous:       true,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-456", "--dangerously-skip-permissions"},
			expectedErr:     false,
			description:     "dangerous付きresume実行コマンドが正しく生成される",
		},
		{
			name:            "complex_sessionid_resume",
			filePath:        "/path/to/conv-2024-01-15-abc123.jsonl",
			dangerous:       false,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "conv-2024-01-15-abc123"},
			expectedErr:     false,
			description:     "複雑なsessionIdでもresume実行コマンドが正しく生成される",
		},
		{
			name:            "non_jsonl_file_error",
			filePath:        "/path/to/session-123.txt",
			dangerous:       false,
			expectedCmdName: "",
			expectedArgs:    nil,
			expectedErr:     true,
			description:     "非JSONLファイルではエラーが発生する",
		},
		{
			name:            "sessionid_with_dots",
			filePath:        "/path/to/session.with.dots.jsonl",
			dangerous:       true,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session.with.dots", "--dangerously-skip-permissions"},
			expectedErr:     false,
			description:     "ドットを含むsessionIdでも正常に動作する",
		},
		{
			name:            "sessionId_injection_attempt",
			filePath:        "/path/to/session-id;evil_command.jsonl",
			dangerous:       false,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-id;evil_command"}, // Now treated as a single argument
			expectedErr:     false,
			description:     "sessionIdへのインジェクション試行が正しく処理される",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdName, args, err := generateResumeCommand(tt.filePath, tt.dangerous)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
				return
			}

			if cmdName != tt.expectedCmdName {
				t.Errorf("Expected command name %q, got %q. %s", tt.expectedCmdName, cmdName, tt.description)
			}
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d arguments, got %d. %s", len(tt.expectedArgs), len(args), tt.description)
			} else {
				for i := range args {
					if args[i] != tt.expectedArgs[i] {
						t.Errorf("Expected argument %d to be %q, got %q. %s", i, tt.expectedArgs[i], args[i], tt.description)
					}
				}
			}
		})
	}
}

func TestResumeWithProjectDirectoryChange(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		dangerous       bool
		expectedCmdName string
		expectedArgs    []string
		expectedDir     string
		expectedErr     bool
		description     string
	}{
		{
			name:            "resume_with_directory_change",
			filePath:        "/path/to/project/session-123.jsonl",
			dangerous:       false,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-123"},
			expectedDir:     "/path/to/project",
			expectedErr:     false,
			description:     "プロジェクトフォルダに移動してからresume実行",
		},
		{
			name:            "dangerous_resume_with_directory_change",
			filePath:        "/path/to/project/session-456.jsonl",
			dangerous:       true,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-456", "--dangerously-skip-permissions"},
			expectedDir:     "/path/to/project",
			expectedErr:     false,
			description:     "プロジェクトフォルダに移動してからdangerous付きresume実行",
		},
		{
			name:            "resume_with_spaces_in_path",
			filePath:        "/path/to/my project/session-789.jsonl", // Note: filePath should not be quoted here, filepath.Dir handles it
			dangerous:       false,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-789"},
			expectedDir:     "/path/to/my project",
			expectedErr:     false,
			description:     "スペースを含むパスでも正常に動作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdName, args, dir, err := generateResumeCommandWithDirectoryChange(tt.filePath, tt.dangerous)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
				return
			}

			if cmdName != tt.expectedCmdName {
				t.Errorf("Expected command name %q, got %q. %s", tt.expectedCmdName, cmdName, tt.description)
			}
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d arguments, got %d. %s", len(tt.expectedArgs), len(args), tt.description)
			} else {
				for i := range args {
					if args[i] != tt.expectedArgs[i] {
						t.Errorf("Expected argument %d to be %q, got %q. %s", i, tt.expectedArgs[i], args[i], tt.description)
					}
				}
			}
			if dir != tt.expectedDir {
				t.Errorf("Expected directory %q, got %q. %s", tt.expectedDir, dir, tt.description)
			}
		})
	}
}

func TestResumeWithCWDDirectoryChange(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create test JSONL files with different CWD values
	testFiles := []struct {
		filename string
		content  string
		cwd      string
	}{
		{
			filename: "session-123.jsonl",
			content:  `{"cwd":"/project/working/directory","sessionId":"session-123","type":"user","message":"test"}`,
			cwd:      "/project/working/directory",
		},
		{
			filename: "session-456.jsonl",
			content:  `{"cwd":"/project/working/directory","sessionId":"session-456","type":"user","message":"test"}`,
			cwd:      "/project/working/directory",
		},
		{
			filename: "session-with-spaces.jsonl",
			content:  `{"cwd":"/project/working directory with spaces","sessionId":"session-with-spaces","type":"user","message":"test"}`,
			cwd:      "/project/working directory with spaces",
		},
	}

	// Write test files
	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.filename)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.filename, err)
		}
	}

	tests := []struct {
		name            string
		filename        string
		dangerous       bool
		expectedCmdName string
		expectedArgs    []string
		expectedDir     string
		expectedErr     bool
		description     string
	}{
		{
			name:            "resume_with_cwd_change",
			filename:        "session-123.jsonl",
			dangerous:       false,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-123"},
			expectedDir:     "/project/working/directory",
			expectedErr:     false,
		},
		{
			name:            "dangerous_resume_with_cwd_change",
			filename:        "session-456.jsonl",
			dangerous:       true,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-456", "--dangerously-skip-permissions"},
			expectedDir:     "/project/working/directory",
			expectedErr:     false,
		},
		{
			name:            "resume_with_cwd_spaces",
			filename:        "session-with-spaces.jsonl",
			dangerous:       false,
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-with-spaces"},
			expectedDir:     "/project/working directory with spaces",
			expectedErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.filename)

			cmdName, args, dir, err := generateResumeCommandWithCWDChange(filePath, tt.dangerous)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
				return
			}

			if cmdName != tt.expectedCmdName {
				t.Errorf("Expected command name %q, got %q. %s", tt.expectedCmdName, cmdName, tt.description)
			}
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d arguments, got %d. %s", len(tt.expectedArgs), len(args), tt.description)
			} else {
				for i := range args {
					if args[i] != tt.expectedArgs[i] {
						t.Errorf("Expected argument %d to be %q, got %q. %s", i, tt.expectedArgs[i], args[i], tt.description)
					}
				}
			}
			if dir != tt.expectedDir {
				t.Errorf("Expected directory %q, got %q. %s", tt.expectedDir, dir, tt.description)
			}
		})
	}
}

func TestResumeKeyHandler(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		keyPressed      string
		expectedCmdName string
		expectedArgs    []string
		expectedErr     bool
		description     string
	}{
		{
			name:            "r_key_normal_resume",
			filePath:        "/path/to/session-123.jsonl",
			keyPressed:      "r",
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-123"},
			expectedErr:     false,
		},
		{
			name:            "shift_r_dangerous_resume",
			filePath:        "/path/to/session-456.jsonl",
			keyPressed:      "R",
			expectedCmdName: "claude",
			expectedArgs:    []string{"-r", "session-456", "--dangerously-skip-permissions"},
			expectedErr:     false,
		},
		{
			name:            "r_key_non_jsonl_error",
			filePath:        "/path/to/session-123.txt",
			keyPressed:      "r",
			expectedCmdName: "",
			expectedArgs:    nil,
			expectedErr:     true,
			description:     "rキーで非JSONLファイル選択時はエラー",
		},
		{
			name:            "shift_r_non_jsonl_error",
			filePath:        "/path/to/session-123.txt",
			keyPressed:      "R",
			expectedCmdName: "",
			expectedArgs:    nil,
			expectedErr:     true,
			description:     "Shift+Rキーで非JSONLファイル選択時はエラー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model with test file
			m := NewModel(".", false)
			m.files = []FileInfo{
				{Path: tt.filePath, IsDir: false},
			}
			m.cursor = 0

			// Test command generation based on key pressed
			dangerous := tt.keyPressed == "R"
			cmdName, args, err := generateResumeCommand(tt.filePath, dangerous)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
				return
			}

			if cmdName != tt.expectedCmdName {
				t.Errorf("Expected command name %q, got %q. %s", tt.expectedCmdName, cmdName, tt.description)
			}
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d arguments, got %d. %s", len(tt.expectedArgs), len(args), tt.description)
			} else {
				for i := range args {
					if args[i] != tt.expectedArgs[i] {
						t.Errorf("Expected argument %d to be %q, got %q. %s", i, tt.expectedArgs[i], args[i], tt.description)
					}
				}
			}
		})
	}
}
