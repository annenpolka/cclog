package cli

import (
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
	
	if config.InputPath != "." {
		t.Errorf("Expected default InputPath to be '.', got '%s'", config.InputPath)
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