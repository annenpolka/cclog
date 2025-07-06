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
				InputPath:  "/path/to/file.jsonl",
				OutputPath: "",
				IsDirectory: false,
			},
			wantErr: false,
		},
		{
			name: "directory input",
			args: []string{"cclog", "-d", "/path/to/dir"},
			expected: Config{
				InputPath:  "/path/to/dir",
				OutputPath: "",
				IsDirectory: true,
			},
			wantErr: false,
		},
		{
			name: "file with output",
			args: []string{"cclog", "/path/to/file.jsonl", "-o", "output.md"},
			expected: Config{
				InputPath:  "/path/to/file.jsonl",
				OutputPath: "output.md",
				IsDirectory: false,
			},
			wantErr: false,
		},
		{
			name: "no arguments - should enable TUI mode by default",
			args: []string{"cclog"},
			expected: Config{
				TUIMode: true,
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
	
	// Should contain .claude/projects
	if !strings.Contains(defaultDir, ".claude/projects") {
		t.Errorf("Default directory should contain '.claude/projects', got: %s", defaultDir)
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

