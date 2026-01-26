package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/annenpolka/cclog/internal/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
		{
			name: "latest flag",
			args: []string{"cclog", "--latest"},
			expected: Config{
				Latest: true,
			},
			wantErr: false,
		},
		{
			name: "latest flag with path",
			args: []string{"cclog", "--latest", "--path", "/custom/path"},
			expected: Config{
				Latest:    true,
				InputPath: "/custom/path",
				Recursive: true, // --path enables recursive by default
			},
			wantErr: false,
		},
		{
			name: "latest flag should disable TUI mode",
			args: []string{"cclog", "--latest"},
			expected: Config{
				Latest:  true,
				TUIMode: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			gotErr := err != nil
			testutil.Diff(t, tt.wantErr, gotErr)

			if !tt.wantErr {
				opts := []cmp.Option{}
				// For TUI mode tests, we don't check InputPath if expected is empty
				if tt.expected.InputPath == "" {
					opts = append(opts, cmpopts.IgnoreFields(Config{}, "InputPath"))
				}
				testutil.Diff(t, tt.expected, config, opts...)
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

func TestRunCommand_ShowTitle_Single(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "titled.jsonl")

	// First user message content should become title
	testContent := `{"type":"user","message":{"role":"user","content":"hello world"},"timestamp":"2025-07-06T05:01:29.618Z","uuid":"u1"}`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := Config{InputPath: testFile, ShowTitle: true}
	out, err := RunCommand(cfg)
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}

	if !strings.HasPrefix(out, "# hello world\n\n") {
		t.Fatalf("expected markdown to start with title, got: %q", out[:min(40, len(out))])
	}
}

func TestRunCommand_ShowTitle_Directory(t *testing.T) {
	tempDir := t.TempDir()
	f1 := filepath.Join(tempDir, "a.jsonl")
	f2 := filepath.Join(tempDir, "b.jsonl")

	// Directory: title taken from first log after filtering; use f1
	c1 := `{"type":"user","message":{"role":"user","content":"dir-title"},"timestamp":"2025-07-06T05:01:29.618Z","uuid":"u1"}`
	c2 := `{"type":"user","message":{"role":"user","content":"other"},"timestamp":"2025-07-06T05:01:30.618Z","uuid":"u2"}`

	if err := os.WriteFile(f1, []byte(c1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte(c2), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Config{InputPath: tempDir, IsDirectory: true, ShowTitle: true}
	out, err := RunCommand(cfg)
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}

	if !strings.HasPrefix(out, "# dir-title\n\n") {
		t.Fatalf("expected directory markdown to start with title, got: %q", out[:min(40, len(out))])
	}
}

// helper: safe substring length for error messages
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestRunCommand_Latest(t *testing.T) {
	// Create a temporary directory with a test JSONL file
	// File name is the session UUID (matching real Claude Code conversation log naming)
	tempDir := t.TempDir()
	sessionUUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	testFile := filepath.Join(tempDir, sessionUUID+".jsonl")

	// Test content (sessionId in content is ignored; file name is the source of truth)
	testContent := `{"type":"user","sessionId":"ignored","message":{"role":"user","content":"hello world"},"timestamp":"2025-07-06T05:01:29.618Z","uuid":"msg-uuid"}`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		InputPath: tempDir,
		Latest:    true,
	}

	output, err := RunCommand(config)
	if err != nil {
		t.Fatalf("RunCommand with --latest failed: %v", err)
	}

	// Output should contain the session UUID (extracted from file name)
	if !strings.Contains(output, sessionUUID) {
		t.Errorf("Output should contain session UUID, got: %s", output)
	}

	// Output should contain the title (first user message content)
	if !strings.Contains(output, "hello world") {
		t.Errorf("Output should contain conversation title, got: %s", output)
	}

	// Output should be tab-separated
	if !strings.Contains(output, "\t") {
		t.Errorf("Output should be tab-separated, got: %s", output)
	}
}

func TestRunCommand_Latest_NoFiles(t *testing.T) {
	// Create an empty temporary directory
	tempDir := t.TempDir()

	config := Config{
		InputPath: tempDir,
		Latest:    true,
	}

	_, err := RunCommand(config)
	if err == nil {
		t.Fatal("RunCommand with --latest should fail when no files found")
	}

	if !strings.Contains(err.Error(), "no conversation files found") {
		t.Errorf("Error should mention no files found, got: %v", err)
	}
}

func TestGetUniqueFilename(t *testing.T) {
	tempDir := t.TempDir()

	// Create initial file
	testFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test: should return _1 suffix when file exists
	result := getUniqueFilename(testFile)
	expected := filepath.Join(tempDir, "test_1.md")
	testutil.Diff(t, expected, result)

	// Create _1 file
	if err := os.WriteFile(expected, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test: should return _2 suffix when _1 also exists
	result = getUniqueFilename(testFile)
	expected = filepath.Join(tempDir, "test_2.md")
	testutil.Diff(t, expected, result)

	// Test: non-existent file returns as-is
	nonExistent := filepath.Join(tempDir, "new.md")
	result = getUniqueFilename(nonExistent)
	testutil.Diff(t, nonExistent, result)
}

func TestGenerateFilenameFromTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "two English words",
			title:    "hello world foo bar",
			expected: "hello_world.md",
		},
		{
			name:     "single word",
			title:    "hello",
			expected: "hello.md",
		},
		{
			name:     "Japanese text truncated",
			title:    "これは長い日本語のタイトルですテスト",
			expected: "これは長い日本語のタイトルですテスト.md",
		},
		{
			name:     "very long Japanese text",
			title:    "これは非常に長い日本語のタイトルで二十文字を超えています",
			expected: "これは非常に長い日本語のタイトルで二十文.md", // 20 runes + .md
		},
		{
			name:     "empty title",
			title:    "",
			expected: "conversation.md",
		},
		{
			name:     "special characters removed",
			title:    "hello! @world# test",
			expected: "hello_world.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFilenameFromTitle(tt.title)
			testutil.Diff(t, tt.expected, result)
		})
	}
}

func TestRunCommand_Latest_AutoOutput(t *testing.T) {
	// Create a temporary directory with a test JSONL file
	tempDir := t.TempDir()
	sessionUUID := "test-session-auto"
	testFile := filepath.Join(tempDir, sessionUUID+".jsonl")

	testContent := `{"type":"user","sessionId":"ignored","message":{"role":"user","content":"hello world test message"},"timestamp":"2025-07-06T05:01:29.618Z","uuid":"msg-uuid"}`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp dir so output file is created there
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	config := Config{
		InputPath:  tempDir,
		Latest:     true,
		AutoOutput: true,
	}

	output, err := RunCommand(config)
	if err != nil {
		t.Fatalf("RunCommand with --latest -o failed: %v", err)
	}

	// Should report output file
	if !strings.Contains(output, "Output written to:") {
		t.Errorf("Expected output message, got: %s", output)
	}

	// Check that file was created with expected name
	expectedFile := "hello_world.md"
	if _, err := os.Stat(filepath.Join(tempDir, expectedFile)); os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", expectedFile)
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
	testutil.True(t, filepath.IsAbs(defaultDir))
}

func TestGetDefaultTUIDirectory_ValidPath(t *testing.T) {
	defaultDir := getDefaultTUIDirectory()

	// Should be a valid path format
	testutil.Diff(t, false, defaultDir == "")

	// Should end with projects
	testutil.True(t, strings.HasSuffix(defaultDir, "projects"))
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
	testutil.Diff(t, expected, result)

	// Test case 2: When .claude directory doesn't exist, should fallback to .config/claude
	os.RemoveAll(claudeDir)
	result = getDefaultTUIDirectory()
	expected = filepath.Join(tempHome, ".config", "claude", "projects")
	testutil.Diff(t, expected, result)
}
