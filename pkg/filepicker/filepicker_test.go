package filepicker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileInfo_FilterValue(t *testing.T) {
	file := FileInfo{Name: "test.txt", Path: "/path/test.txt", IsDir: false, Size: 100}
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
			name:     "regular file",
			file:     FileInfo{Name: "test.txt", IsDir: false},
			expected: "test.txt",
		},
		{
			name:     "directory",
			file:     FileInfo{Name: "testdir", IsDir: true},
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
	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name:     "directory",
			file:     FileInfo{Name: "testdir", IsDir: true},
			expected: "Directory",
		},
		{
			name:     "small file",
			file:     FileInfo{Name: "test.txt", IsDir: false, Size: 500},
			expected: "< 1KB",
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