package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectorySelectionHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test with directory selection
	if stat, err := os.Stat(tempDir); err == nil && stat.IsDir() {
		// This should be true for directory
		if !stat.IsDir() {
			t.Error("Expected directory to be identified as directory")
		}
	} else {
		t.Fatalf("Failed to stat test directory: %v", err)
	}
}

func TestFileSelectionHandling(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.jsonl")

	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with file selection
	if stat, err := os.Stat(tempFile); err == nil {
		// This should be false for file
		if stat.IsDir() {
			t.Error("Expected file to not be identified as directory")
		}
	} else {
		t.Fatalf("Failed to stat test file: %v", err)
	}
}

func TestShouldSetDirectoryFlag(t *testing.T) {
	// Create test directory
	tempDir := t.TempDir()

	// Test function
	isDir := shouldSetDirectoryFlag(tempDir)
	if !isDir {
		t.Errorf("Expected shouldSetDirectoryFlag to return true for directory")
	}

	// Create test file
	tempFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test function with file
	isDir = shouldSetDirectoryFlag(tempFile)
	if isDir {
		t.Errorf("Expected shouldSetDirectoryFlag to return false for file")
	}
}

func TestShouldSetDirectoryFlag_NonExistent(t *testing.T) {
	// Test with non-existent path
	isDir := shouldSetDirectoryFlag("/nonexistent/path")
	if isDir {
		t.Errorf("Expected shouldSetDirectoryFlag to return false for non-existent path")
	}
}
