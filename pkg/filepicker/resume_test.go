package filepicker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResumeCommand(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		dangerous      bool
		expectedCmd    string
		expectedErr    bool
		description    string
	}{
		{
			name:           "normal_resume_command",
			filePath:       "/path/to/session-123.jsonl",
			dangerous:      false,
			expectedCmd:    "claude -r session-123",
			expectedErr:    false,
			description:    "通常のresume実行コマンドが正しく生成される",
		},
		{
			name:           "dangerous_resume_command",
			filePath:       "/path/to/session-456.jsonl",
			dangerous:      true,
			expectedCmd:    "claude -r session-456 --dangerously-skip-permissions",
			expectedErr:    false,
			description:    "dangerous付きresume実行コマンドが正しく生成される",
		},
		{
			name:           "complex_sessionid_resume",
			filePath:       "/path/to/conv-2024-01-15-abc123.jsonl",
			dangerous:      false,
			expectedCmd:    "claude -r conv-2024-01-15-abc123",
			expectedErr:    false,
			description:    "複雑なsessionIdでもresume実行コマンドが正しく生成される",
		},
		{
			name:           "non_jsonl_file_error",
			filePath:       "/path/to/session-123.txt",
			dangerous:      false,
			expectedCmd:    "",
			expectedErr:    true,
			description:    "非JSONLファイルではエラーが発生する",
		},
		{
			name:           "sessionid_with_dots",
			filePath:       "/path/to/session.with.dots.jsonl",
			dangerous:      true,
			expectedCmd:    "claude -r session.with.dots --dangerously-skip-permissions",
			expectedErr:    false,
			description:    "ドットを含むsessionIdでも正常に動作する",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test resume command generation
			cmd, err := generateResumeCommand(tt.filePath, tt.dangerous)

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

			if cmd != tt.expectedCmd {
				t.Errorf("Expected command %q, got %q. %s", tt.expectedCmd, cmd, tt.description)
			}
		})
	}
}

func TestResumeWithProjectDirectoryChange(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		dangerous      bool
		expectedCmd    string
		expectedErr    bool
		description    string
	}{
		{
			name:           "resume_with_directory_change",
			filePath:       "/path/to/project/session-123.jsonl",
			dangerous:      false,
			expectedCmd:    "cd /path/to/project && claude -r session-123",
			expectedErr:    false,
			description:    "プロジェクトフォルダに移動してからresume実行",
		},
		{
			name:           "dangerous_resume_with_directory_change",
			filePath:       "/path/to/project/session-456.jsonl",
			dangerous:      true,
			expectedCmd:    "cd /path/to/project && claude -r session-456 --dangerously-skip-permissions",
			expectedErr:    false,
			description:    "プロジェクトフォルダに移動してからdangerous付きresume実行",
		},
		{
			name:           "resume_with_spaces_in_path",
			filePath:       "/path/to/my project/session-789.jsonl",
			dangerous:      false,
			expectedCmd:    "cd \"/path/to/my project\" && claude -r session-789",
			expectedErr:    false,
			description:    "スペースを含むパスでも正常に動作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test resume command generation with directory change
			cmd, err := generateResumeCommandWithDirectoryChange(tt.filePath, tt.dangerous)

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

			if cmd != tt.expectedCmd {
				t.Errorf("Expected command %q, got %q. %s", tt.expectedCmd, cmd, tt.description)
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
		name        string
		filename    string
		dangerous   bool
		expectedCmd string
		expectedErr bool
		description string
	}{
		{
			name:        "resume_with_cwd_change",
			filename:    "session-123.jsonl",
			dangerous:   false,
			expectedCmd: "cd /project/working/directory && claude -r session-123",
			expectedErr: false,
			description: "CWDフィールドからプロジェクトディレクトリに移動してresume実行",
		},
		{
			name:        "dangerous_resume_with_cwd_change",
			filename:    "session-456.jsonl",
			dangerous:   true,
			expectedCmd: "cd /project/working/directory && claude -r session-456 --dangerously-skip-permissions",
			expectedErr: false,
			description: "CWDフィールドからプロジェクトディレクトリに移動してdangerous付きresume実行",
		},
		{
			name:        "resume_with_cwd_spaces",
			filename:    "session-with-spaces.jsonl",
			dangerous:   false,
			expectedCmd: "cd \"/project/working directory with spaces\" && claude -r session-with-spaces",
			expectedErr: false,
			description: "スペースを含むCWDパスでも正常に動作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.filename)
			
			// Test resume command generation with CWD change
			cmd, err := generateResumeCommandWithCWDChange(filePath, tt.dangerous)

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

			if cmd != tt.expectedCmd {
				t.Errorf("Expected command %q, got %q. %s", tt.expectedCmd, cmd, tt.description)
			}
		})
	}
}

func TestResumeKeyHandler(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		keyPressed     string
		expectedCmd    string
		expectedErr    bool
		description    string
	}{
		{
			name:           "r_key_normal_resume",
			filePath:       "/path/to/session-123.jsonl",
			keyPressed:     "r",
			expectedCmd:    "claude -r session-123",
			expectedErr:    false,
			description:    "rキーで通常のresume実行",
		},
		{
			name:           "shift_r_dangerous_resume",
			filePath:       "/path/to/session-456.jsonl",
			keyPressed:     "R",
			expectedCmd:    "claude -r session-456 --dangerously-skip-permissions",
			expectedErr:    false,
			description:    "Shift+Rキーでdangerous付きresume実行",
		},
		{
			name:           "r_key_non_jsonl_error",
			filePath:       "/path/to/session-123.txt",
			keyPressed:     "r",
			expectedCmd:    "",
			expectedErr:    true,
			description:    "rキーで非JSONLファイル選択時はエラー",
		},
		{
			name:           "shift_r_non_jsonl_error",
			filePath:       "/path/to/session-123.txt",
			keyPressed:     "R",
			expectedCmd:    "",
			expectedErr:    true,
			description:    "Shift+Rキーで非JSONLファイル選択時はエラー",
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
			cmd, err := generateResumeCommand(tt.filePath, dangerous)

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

			if cmd != tt.expectedCmd {
				t.Errorf("Expected command %q, got %q. %s", tt.expectedCmd, cmd, tt.description)
			}
		})
	}
}