package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs_TUIMode(t *testing.T) {
	args := []string{"cclog", "--tui"}
	config, err := ParseArgs(args)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !config.TUIMode {
		t.Error("Expected TUIMode to be true")
	}
}

func TestParseArgs_TUIModeWithDirectory(t *testing.T) {
	args := []string{"cclog", "--tui", "/path/to/logs"}
	config, err := ParseArgs(args)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !config.TUIMode {
		t.Error("Expected TUIMode to be true")
	}

	if config.InputPath != "/path/to/logs" {
		t.Errorf("Expected InputPath to be '/path/to/logs', got '%s'", config.InputPath)
	}
}

func TestParseArgs_TUIModeDefaultDirectory(t *testing.T) {
	args := []string{"cclog", "--tui"}
	config, err := ParseArgs(args)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that config.InputPath is set to either the default directory or fallback
	expectedDir := getDefaultTUIDirectory()
	if config.InputPath != expectedDir && config.InputPath != "." {
		t.Errorf("Expected default InputPath to be '%s' or '.', got '%s'", expectedDir, config.InputPath)
	}
}

func TestRunCommand_TUIMode(t *testing.T) {
	config := Config{
		TUIMode:   true,
		InputPath: ".",
	}

	// This should return empty string and no error for TUI mode
	// since TUI mode is handled differently
	output, err := RunCommand(config)

	if err != nil {
		t.Errorf("Expected no error for TUI mode, got %v", err)
	}

	if output != "" {
		t.Errorf("Expected empty output for TUI mode, got '%s'", output)
	}
}

func TestEnsureDefaultDirectoryExists(t *testing.T) {
	// Create a temporary directory to simulate user home
	tempHome := t.TempDir()
	testDir := filepath.Join(tempHome, ".claude", "projects")

	// Directory should not exist initially
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Errorf("Test directory should not exist initially")
	}

	// Call the function to check if directory exists
	err := ensureDefaultDirectoryExists(testDir)
	if err == nil {
		t.Errorf("Expected error for non-existent directory")
	}

	// Create directory first
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Now the function should return no error
	err = ensureDefaultDirectoryExists(testDir)
	if err != nil {
		t.Errorf("Expected no error for existing directory, got %v", err)
	}
}

func TestEnsureDefaultDirectoryExists_AlreadyExists(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "existing")

	// Create the directory first
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Call the function - should not error on existing directory
	err = ensureDefaultDirectoryExists(testDir)
	if err != nil {
		t.Errorf("Expected no error for existing directory, got %v", err)
	}

	// Directory should still exist
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("Directory should still exist")
	}
}
