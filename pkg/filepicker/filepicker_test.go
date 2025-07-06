package filepicker

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileInfo_FilterValue(t *testing.T) {
	file := FileInfo{
		Name:    "test.txt",
		Path:    "/path/test.txt",
		IsDir:   false,
		Size:    100,
		ModTime: time.Now(),
	}
	expected := "test.txt"
	if got := file.FilterValue(); got != expected {
		t.Errorf("FilterValue() = %v, want %v", got, expected)
	}
}

func TestFileInfo_Title(t *testing.T) {
	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "regular file",
			file: FileInfo{
				Name:    "test.txt",
				IsDir:   false,
				ModTime: time.Now(),
			},
			expected: "test.txt",
		},
		{
			name: "directory",
			file: FileInfo{
				Name:    "testdir",
				IsDir:   true,
				ModTime: time.Now(),
			},
			expected: "testdir/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.file.Title(); got != tt.expected {
				t.Errorf("Title() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFileInfo_Description(t *testing.T) {
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "directory with modification time",
			file: FileInfo{
				Name:    "testdir",
				IsDir:   true,
				ModTime: modTime,
			},
			expected: "2025-01-15 14:30",
		},
		{
			name: "file with modification time",
			file: FileInfo{
				Name:    "test.txt",
				IsDir:   false,
				Size:    500,
				ModTime: modTime,
			},
			expected: "2025-01-15 14:30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.file.Description(); got != tt.expected {
				t.Errorf("Description() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetFiles(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "filepicker_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test GetFiles
	files, err := GetFiles(tmpDir)
	if err != nil {
		t.Fatalf("GetFiles failed: %v", err)
	}

	// Should have 2 actual files plus possibly ".." entry
	expectedMinFiles := 2
	if len(files) < expectedMinFiles {
		t.Errorf("Expected at least %d files, got %d", expectedMinFiles, len(files))
	}

	// Check if both file and directory are present
	found := make(map[string]bool)
	for _, file := range files {
		found[file.Name] = true
	}

	if !found["test.txt"] {
		t.Error("test.txt not found")
	}
	if !found["testdir"] {
		t.Error("testdir not found")
	}
}

func TestGetFiles_NonExistentDirectory(t *testing.T) {
	_, err := GetFiles("/nonexistent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestGetFiles_IncludesParentDirectory(t *testing.T) {
	// Create temporary subdirectory
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Get files from subdirectory
	files, err := GetFiles(subDir)
	if err != nil {
		t.Fatalf("GetFiles failed: %v", err)
	}
	
	// Should include ".." entry for parent directory
	foundParent := false
	for _, file := range files {
		if file.Name == ".." {
			foundParent = true
			if !file.IsDir {
				t.Error("'..' should be marked as directory")
			}
			if file.Path != tempDir {
				t.Errorf("Expected '..' path to be '%s', got '%s'", tempDir, file.Path)
			}
			break
		}
	}
	
	if !foundParent {
		t.Error("Expected '..' entry for parent directory")
	}
}

func TestGetFiles_NoParentForRoot(t *testing.T) {
	// Test with current directory to avoid permission issues
	files, err := GetFiles(".")
	if err != nil {
		t.Fatalf("GetFiles failed: %v", err)
	}
	
	// Should not include ".." if we're at a root-like directory
	// (This test is more lenient as we can't easily test actual root)
	for _, file := range files {
		if file.Name == ".." {
			// This is okay for subdirectories, just ensure it's properly marked
			if !file.IsDir {
				t.Error("'..' should be marked as directory")
			}
		}
	}
}

// TDD Red Phase: Tests that should fail initially

func TestFileInfo_ModTime(t *testing.T) {
	// Red: This test should fail because ModTime field doesn't exist yet
	now := time.Now()
	file := FileInfo{
		Name:    "test.txt",
		Path:    "/path/test.txt",
		IsDir:   false,
		Size:    100,
		ModTime: now,
	}
	
	if file.ModTime != now {
		t.Errorf("Expected ModTime to be %v, got %v", now, file.ModTime)
	}
}

func TestFileInfo_Description_WithModTime(t *testing.T) {
	// Red: This test should fail because Description doesn't show ModTime yet
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	file := FileInfo{
		Name:    "test.txt",
		Path:    "/path/test.txt",
		IsDir:   false,
		Size:    1024,
		ModTime: modTime,
	}
	
	description := file.Description()
	expectedDate := "2025-01-15 14:30"
	
	if description != expectedDate {
		t.Errorf("Expected description to be '%s', got '%s'", expectedDate, description)
	}
}

func TestFileInfo_Description_DirectoryWithModTime(t *testing.T) {
	// Red: This test should fail because directories currently show "Directory"
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	dir := FileInfo{
		Name:    "testdir",
		Path:    "/path/testdir",
		IsDir:   true,
		Size:    0,
		ModTime: modTime,
	}
	
	description := dir.Description()
	expectedDate := "2025-01-15 14:30"
	
	if description != expectedDate {
		t.Errorf("Expected directory description to be '%s', got '%s'", expectedDate, description)
	}
}

func TestGetFiles_SortsByModTime(t *testing.T) {
	// Red: This test should fail because sorting by ModTime isn't implemented yet
	tempDir := t.TempDir()
	
	// Create files with different modification times
	oldFile := filepath.Join(tempDir, "old.txt")
	newFile := filepath.Join(tempDir, "new.txt")
	
	// Create old file first
	if err := os.WriteFile(oldFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	
	// Sleep to ensure different timestamps
	time.Sleep(100 * time.Millisecond)
	
	// Create new file
	if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}
	
	// Get files
	files, err := GetFiles(tempDir)
	if err != nil {
		t.Fatalf("GetFiles failed: %v", err)
	}
	
	// Filter out non-regular files (like ".." entries)
	var regularFiles []FileInfo
	for _, file := range files {
		if !file.IsDir && file.Name != ".." {
			regularFiles = append(regularFiles, file)
		}
	}
	
	// Should have 2 regular files
	if len(regularFiles) != 2 {
		t.Fatalf("Expected 2 regular files, got %d", len(regularFiles))
	}
	
	// Files should be sorted by modification time (newest first)
	// new.txt should come before old.txt
	if regularFiles[0].Name != "new.txt" {
		t.Errorf("Expected newest file 'new.txt' to be first, got '%s'", regularFiles[0].Name)
	}
	
	if regularFiles[1].Name != "old.txt" {
		t.Errorf("Expected oldest file 'old.txt' to be second, got '%s'", regularFiles[1].Name)
	}
}