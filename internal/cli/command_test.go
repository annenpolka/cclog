package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "single file input",
			args: []string{"cclog", "/path/to/file.jsonl"},
			expected: Config{
				InputPath:   "/path/to/file.jsonl",
				OutputPath:  "",
				IsDirectory: false,
			},
			wantErr: false,
		},
		{
			name: "directory input",
			args: []string{"cclog", "-d", "/path/to/dir"},
			expected: Config{
				InputPath:   "/path/to/dir",
				OutputPath:  "",
				IsDirectory: true,
			},
			wantErr: false,
		},
		{
			name: "file with output",
			args: []string{"cclog", "/path/to/file.jsonl", "-o", "output.md"},
			expected: Config{
				InputPath:   "/path/to/file.jsonl",
				OutputPath:  "output.md",
				IsDirectory: false,
			},
			wantErr: false,
		},
		{
			name: "no arguments - should enable TUI mode and recursive mode by default",
			args: []string{"cclog"},
			expected: Config{
				TUIMode:   true,
				Recursive: true,
			},
			wantErr: false,
		},
		{
			name: "help flag",
			args: []string{"cclog", "-h"},
			expected: Config{
				ShowHelp: true,
			},
			wantErr: false,
		},
		{
			name: "include all flag",
			args: []string{"cclog", "/path/to/file.jsonl", "--include-all"},
			expected: Config{
				InputPath:   "/path/to/file.jsonl",
				OutputPath:  "",
				IsDirectory: false,
				IncludeAll:  true,
			},
			wantErr: false,
		},
		{
			name: "explicit TUI mode",
			args: []string{"cclog", "--tui"},
			expected: Config{
				TUIMode: true,
			},
			wantErr: false,
		},
		{
			name: "TUI mode with path",
			args: []string{"cclog", "--tui", "/path/to/logs"},
			expected: Config{
				InputPath: "/path/to/logs",
				TUIMode:   true,
			},
			wantErr: false,
		},
		{
			name: "recursive flag",
			args: []string{"cclog", "--recursive", "/path/to/logs"},
			expected: Config{
				InputPath: "/path/to/logs",
				Recursive: true,
				TUIMode:   true,
			},
			wantErr: false,
		},
		{
			name: "recursive and TUI mode combined",
			args: []string{"cclog", "--recursive", "--tui"},
			expected: Config{
				Recursive: true,
				TUIMode:   true,
			},
			wantErr: false,
		},
		{
			name: "recursive flag alone should enable TUI mode",
			args: []string{"cclog", "--recursive"},
			expected: Config{
				Recursive: true,
				TUIMode:   true,
			},
			wantErr: false,
		},
		{
			name: "short recursive flag alone should enable TUI mode",
			args: []string{"cclog", "-r"},
			expected: Config{
				Recursive: true,
				TUIMode:   true,
			},
			wantErr: false,
		},
		{
			name: "recursive with path should enable TUI mode",
			args: []string{"cclog", "--recursive", "/path/to/logs"},
			expected: Config{
				InputPath: "/path/to/logs",
				Recursive: true,
				TUIMode:   true,
			},
			wantErr: false,
		},
		{
			name: "path option should set input path",
			args: []string{"cclog", "--path", "/custom/path"},
			expected: Config{
				InputPath: "/custom/path",
				TUIMode:   true,
				Recursive: true,
			},
			wantErr: false,
		},
		{
			name:     "path option without value should return error",
			args:     []string{"cclog", "--path"},
			expected: Config{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantErr {
				// For TUI mode tests, we don't check InputPath if expected is empty
				// because the default directory is set automatically
				if tt.expected.InputPath != "" && config.InputPath != tt.expected.InputPath {
					t.Errorf("Expected InputPath %s, got %s", tt.expected.InputPath, config.InputPath)
				}
				if config.OutputPath != tt.expected.OutputPath {
					t.Errorf("Expected OutputPath %s, got %s", tt.expected.OutputPath, config.OutputPath)
				}
				if config.IsDirectory != tt.expected.IsDirectory {
					t.Errorf("Expected IsDirectory %v, got %v", tt.expected.IsDirectory, config.IsDirectory)
				}
				if config.ShowHelp != tt.expected.ShowHelp {
					t.Errorf("Expected ShowHelp %v, got %v", tt.expected.ShowHelp, config.ShowHelp)
				}
				if config.IncludeAll != tt.expected.IncludeAll {
					t.Errorf("Expected IncludeAll %v, got %v", tt.expected.IncludeAll, config.IncludeAll)
				}
				if config.TUIMode != tt.expected.TUIMode {
					t.Errorf("Expected TUIMode %v, got %v", tt.expected.TUIMode, config.TUIMode)
				}
				if config.Recursive != tt.expected.Recursive {
					t.Errorf("Expected Recursive %v, got %v", tt.expected.Recursive, config.Recursive)
				}
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jsonl")

	testContent := `{"type":"user","message":{"role":"user","content":"test"},"timestamp":"2025-07-06T05:01:29.618Z","uuid":"test-uuid"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"response"}]},"timestamp":"2025-07-06T05:01:30.618Z","uuid":"test-uuid-2"}`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		InputPath:   testFile,
		OutputPath:  "",
		IsDirectory: false,
	}

	output, err := RunCommand(config)
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}

	if !strings.Contains(output, "# Conversation Log") {
		t.Error("Output should contain conversation log header")
	}

	if !strings.Contains(output, "test") {
		t.Error("Output should contain test message content")
	}

	if !strings.Contains(output, "response") {
		t.Error("Output should contain response message content")
	}
}

func TestRunCommandWithDirectory(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "test1.jsonl")
	testFile2 := filepath.Join(tempDir, "test2.jsonl")

	testContent := `{"type":"user","message":{"role":"user","content":"test1"},"timestamp":"2025-07-06T05:01:29.618Z","uuid":"test-uuid"}`

	err := os.WriteFile(testFile1, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	testContent2 := `{"type":"user","message":{"role":"user","content":"test2"},"timestamp":"2025-07-06T05:01:29.618Z","uuid":"test-uuid"}`
	err = os.WriteFile(testFile2, []byte(testContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	config := Config{
		InputPath:   tempDir,
		OutputPath:  "",
		IsDirectory: true,
	}

	output, err := RunCommand(config)
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}

	if !strings.Contains(output, "# Claude Conversation Logs") {
		t.Error("Output should contain multiple conversations header")
	}

	if !strings.Contains(output, "test1") {
		t.Error("Output should contain content from test1")
	}

	if !strings.Contains(output, "test2") {
		t.Error("Output should contain content from test2")
	}
}

func TestGetDefaultTUIDirectory(t *testing.T) {
	defaultDir := getDefaultTUIDirectory()

	// Should contain either .claude/projects or .config/claude/projects
	hasClaudeProjects := strings.Contains(defaultDir, ".claude/projects")
	hasConfigClaudeProjects := strings.Contains(defaultDir, ".config/claude/projects")

	if !hasClaudeProjects && !hasConfigClaudeProjects {
		t.Errorf("Default directory should contain '.claude/projects' or '.config/claude/projects', got: %s", defaultDir)
	}

	// Should be an absolute path
	if !filepath.IsAbs(defaultDir) {
		t.Errorf("Default directory should be absolute path, got: %s", defaultDir)
	}
}

func TestGetDefaultTUIDirectory_ValidPath(t *testing.T) {
	defaultDir := getDefaultTUIDirectory()

	// Should be a valid path format
	if defaultDir == "" {
		t.Error("Default directory should not be empty")
	}

	// Should end with projects
	if !strings.HasSuffix(defaultDir, "projects") {
		t.Errorf("Default directory should end with 'projects', got: %s", defaultDir)
	}
}

func TestGetDefaultTUIDirectory_FallbackBehavior(t *testing.T) {
	// Create a temporary directory to simulate user home
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")

	defer func() {
		// Restore original HOME
		os.Setenv("HOME", originalHome)
	}()

	// Test case 1: When .claude directory exists, it should be preferred
	os.Setenv("HOME", tempHome)
	claudeDir := filepath.Join(tempHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	result := getDefaultTUIDirectory()
	expected := filepath.Join(tempHome, ".claude", "projects")
	if result != expected {
		t.Errorf("Expected %s when .claude exists, got %s", expected, result)
	}

	// Test case 2: When .claude directory doesn't exist, should fallback to .config/claude
	os.RemoveAll(claudeDir)
	result = getDefaultTUIDirectory()
	expected = filepath.Join(tempHome, ".config", "claude", "projects")
	if result != expected {
		t.Errorf("Expected %s when .claude doesn't exist, got %s", expected, result)
	}
}
