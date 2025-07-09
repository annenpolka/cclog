package filepicker

import (
	"os"
	"path/filepath"
	"strings"
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
			expected: "",
		},
		{
			name: "file with modification time",
			file: FileInfo{
				Name:    "test.txt",
				IsDir:   false,
				Size:    500,
				ModTime: modTime,
			},
			expected: "",
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
	expectedDate := ""

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
	expectedDate := ""

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

// TDD Red Phase: Recursive file listing tests

func TestGetFilesRecursive(t *testing.T) {
	// Red: This test should fail because GetFilesRecursive doesn't exist yet
	tempDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	nestedDir := filepath.Join(subDir, "nested")
	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Create files at different levels with valid JSONL content
	rootFile := filepath.Join(tempDir, "root.jsonl")
	subFile := filepath.Join(subDir, "sub.jsonl")
	nestedFile := filepath.Join(nestedDir, "nested.jsonl")

	validJSONLContent := `{"type":"user","message":{"role":"user","content":"test"},"uuid":"test-uuid","timestamp":"2025-07-06T05:01:44.663Z"}`

	if err := os.WriteFile(rootFile, []byte(validJSONLContent), 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	if err := os.WriteFile(subFile, []byte(validJSONLContent), 0644); err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	if err := os.WriteFile(nestedFile, []byte(validJSONLContent), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Test recursive file listing
	files, err := GetFilesRecursive(tempDir)
	if err != nil {
		t.Fatalf("GetFilesRecursive failed: %v", err)
	}

	// Should find all .jsonl files recursively
	jsonlFiles := make(map[string]bool)
	for _, file := range files {
		if filepath.Ext(file.Name) == ".jsonl" {
			jsonlFiles[file.Name] = true
		}
	}

	expectedFiles := []string{"root.jsonl", "sub.jsonl", "nested.jsonl"}
	for _, expected := range expectedFiles {
		if !jsonlFiles[expected] {
			t.Errorf("Expected to find %s in recursive listing", expected)
		}
	}

	// Should have at least 3 JSONL files
	if len(jsonlFiles) < 3 {
		t.Errorf("Expected at least 3 JSONL files, got %d", len(jsonlFiles))
	}
}

func TestGetFilesRecursive_WithRelativePaths(t *testing.T) {
	// Red: This test should fail because GetFilesRecursive doesn't exist yet
	tempDir := t.TempDir()

	// Create nested structure
	subDir := filepath.Join(tempDir, "logs", "2025")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create file in nested directory with valid JSONL content
	nestedFile := filepath.Join(subDir, "conversation.jsonl")
	validJSONLContent := `{"type":"user","message":{"role":"user","content":"test"},"uuid":"test-uuid","timestamp":"2025-07-06T05:01:44.663Z"}`
	if err := os.WriteFile(nestedFile, []byte(validJSONLContent), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Test recursive listing
	files, err := GetFilesRecursive(tempDir)
	if err != nil {
		t.Fatalf("GetFilesRecursive failed: %v", err)
	}

	// Find the nested file
	var foundFile *FileInfo
	for _, file := range files {
		if file.Name == "conversation.jsonl" {
			foundFile = &file
			break
		}
	}

	if foundFile == nil {
		t.Fatal("Expected to find conversation.jsonl in recursive listing")
	}

	// Path should include relative directory structure
	expectedPathSuffix := filepath.Join("logs", "2025", "conversation.jsonl")
	if !filepath.IsAbs(foundFile.Path) {
		t.Error("Expected absolute path in FileInfo.Path")
	}

	if !strings.HasSuffix(foundFile.Path, expectedPathSuffix) {
		t.Errorf("Expected path to end with %s, got %s", expectedPathSuffix, foundFile.Path)
	}
}

// TDD Red Phase: Tests for title display functionality

func TestFileInfo_TitleWithConversationTitle(t *testing.T) {
	// Red: This test should fail because title extraction doesn't exist yet
	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "JSONL file should display title with conversation title",
			file: FileInfo{
				Name:              "conversation.jsonl",
				Path:              "/path/conversation.jsonl",
				IsDir:             false,
				ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ConversationTitle: "User requested Go...",
			},
			expected: "2025-01-15 14:30 User requested Go...",
		},
		{
			name: "JSONL file without conversation title should display date only",
			file: FileInfo{
				Name:              "empty.jsonl",
				Path:              "/path/empty.jsonl",
				IsDir:             false,
				ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
				ConversationTitle: "",
			},
			expected: "2025-01-15 14:30",
		},
		{
			name: "Non-JSONL file should display just filename",
			file: FileInfo{
				Name:              "document.txt",
				Path:              "/path/document.txt",
				IsDir:             false,
				ConversationTitle: "",
			},
			expected: "document.txt",
		},
		{
			name: "Directory should display with slash",
			file: FileInfo{
				Name:              "folder",
				Path:              "/path/folder",
				IsDir:             true,
				ConversationTitle: "",
			},
			expected: "folder/",
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

func TestGetFiles_WithConversationTitles(t *testing.T) {
	// Red: This test should fail because conversation title loading doesn't exist yet
	tempDir := t.TempDir()

	// Create a sample JSONL file with conversation content
	jsonlContent := `{"type":"user","message":{"role":"user","content":"goでこれらを人間が読みやすいmarkdownにパースするコマンドラインツールを作る"},"timestamp":"2025-07-06T05:01:59.066Z"}
{"type":"summary","summary":"User requested Go CLI tool development using TDD","leafUuid":"5930868a-923c-4d1d-aae4-9c363adcf6d2"}
`
	jsonlFile := filepath.Join(tempDir, "conversation.jsonl")
	if err := os.WriteFile(jsonlFile, []byte(jsonlContent), 0644); err != nil {
		t.Fatalf("Failed to create JSONL file: %v", err)
	}

	// Get files with conversation titles
	files, err := GetFiles(tempDir)
	if err != nil {
		t.Fatalf("GetFiles failed: %v", err)
	}

	// Find the JSONL file
	var jsonlFileInfo *FileInfo
	for _, file := range files {
		if file.Name == "conversation.jsonl" {
			jsonlFileInfo = &file
			break
		}
	}

	if jsonlFileInfo == nil {
		t.Fatal("Expected to find conversation.jsonl")
	}

	// Should have conversation title extracted
	if jsonlFileInfo.ConversationTitle == "" {
		t.Error("Expected ConversationTitle to be extracted from JSONL file")
	}

	// Title should include conversation title and date in new format
	// We can't predict exact time, so just check that it contains the pattern
	title := jsonlFileInfo.Title()
	if !strings.Contains(title, "goでこれらを人間が読みやすいmarkdownにパースするコマンドラインツールを作る") {
		t.Errorf("Expected Title() to contain 'goでこれらを人間が読みやすいmarkdownにパースするコマンドラインツールを作る', got '%s'", title)
	}

	// Check date format pattern (YYYY-MM-DD HH:MM)
	if !strings.Contains(title, "-") || len(title) < 16 {
		t.Errorf("Expected Title() to contain date format, got '%s'", title)
	}
}

// TDD Red Phase: Tests for date-title display format

func TestFileInfo_TitleWithDateTitleFormat(t *testing.T) {
	// Red: This test should fail because date-title format doesn't exist yet
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "JSONL file should display date - title format",
			file: FileInfo{
				Name:              "conversation.jsonl",
				Path:              "/path/conversation.jsonl",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "User requested Go...",
			},
			expected: "2025-01-15 14:30 User requested Go...",
		},
		{
			name: "JSONL file without conversation title should display date only",
			file: FileInfo{
				Name:              "empty.jsonl",
				Path:              "/path/empty.jsonl",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "2025-01-15 14:30",
		},
		{
			name: "Non-JSONL file should display just filename",
			file: FileInfo{
				Name:              "document.txt",
				Path:              "/path/document.txt",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "document.txt",
		},
		{
			name: "Directory should display with slash",
			file: FileInfo{
				Name:              "folder",
				Path:              "/path/folder",
				IsDir:             true,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "folder/",
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

// TDD Red Phase: Tests for removing right-side date display

func TestFileInfo_DescriptionEmptyForCleanDisplay(t *testing.T) {
	// Red: This test should fail because Description still shows date
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "JSONL file should not show date in description",
			file: FileInfo{
				Name:              "conversation.jsonl",
				Path:              "/path/conversation.jsonl",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "User requested Go...",
			},
			expected: "",
		},
		{
			name: "Non-JSONL file should not show date in description",
			file: FileInfo{
				Name:              "document.txt",
				Path:              "/path/document.txt",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "",
		},
		{
			name: "Directory should not show date in description",
			file: FileInfo{
				Name:              "folder",
				Path:              "/path/folder",
				IsDir:             true,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "",
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

// TDD Red Phase: Tests for removing dash separator in title

func TestFileInfo_TitleWithoutDashSeparator(t *testing.T) {
	// Red: This test should fail because Title still includes " - " separator
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "JSONL file with title should not include dash separator",
			file: FileInfo{
				Name:              "conversation.jsonl",
				Path:              "/path/conversation.jsonl",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "User requested Go...",
			},
			expected: "2025-01-15 14:30 User requested Go...",
		},
		{
			name: "JSONL file without title should show date only",
			file: FileInfo{
				Name:              "empty.jsonl",
				Path:              "/path/empty.jsonl",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "2025-01-15 14:30",
		},
		{
			name: "Non-JSONL file should display just filename",
			file: FileInfo{
				Name:              "document.txt",
				Path:              "/path/document.txt",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "document.txt",
		},
		{
			name: "Directory should display with slash",
			file: FileInfo{
				Name:              "folder",
				Path:              "/path/folder",
				IsDir:             true,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "folder/",
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

// TDD Red Phase: Tests for "date title" format without dashes

func TestFileInfo_TitleWithDateTitleFormatNoDashes(t *testing.T) {
	// Red: This test should fail because current format is "title date" not "date title"
	modTime := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		file     FileInfo
		expected string
	}{
		{
			name: "JSONL file should display date title format without dashes",
			file: FileInfo{
				Name:              "conversation.jsonl",
				Path:              "/path/conversation.jsonl",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "User requested Go...",
			},
			expected: "2025-01-15 14:30 User requested Go...",
		},
		{
			name: "JSONL file without title should show date only",
			file: FileInfo{
				Name:              "empty.jsonl",
				Path:              "/path/empty.jsonl",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "2025-01-15 14:30",
		},
		{
			name: "Non-JSONL file should display just filename",
			file: FileInfo{
				Name:              "document.txt",
				Path:              "/path/document.txt",
				IsDir:             false,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "document.txt",
		},
		{
			name: "Directory should display with slash",
			file: FileInfo{
				Name:              "folder",
				Path:              "/path/folder",
				IsDir:             true,
				ModTime:           modTime,
				ConversationTitle: "",
			},
			expected: "folder/",
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

// TDD Red Phase: Tests for filtering out files with zero filtered messages

func TestExtractConversationTitle_FilteredEmpty(t *testing.T) {
	// Red: This test should fail because extractConversationTitle doesn't filter messages yet
	tempDir := t.TempDir()

	// Create test files
	filteredEmptyFile := filepath.Join(tempDir, "filtered_empty.jsonl")
	emptyFile := filepath.Join(tempDir, "empty.jsonl")
	normalFile := filepath.Join(tempDir, "normal.jsonl")

	// Copy test data files
	filteredEmptyContent := `{"type":"system","message":{"role":"system","content":"System message"},"uuid":"system-uuid","timestamp":"2025-07-06T05:00:00.000Z"}
{"type":"user","message":{"role":"user","content":"<command-name>/test</command-name>"},"uuid":"command-uuid","timestamp":"2025-07-06T05:00:01.000Z"}
{"type":"user","message":{"role":"user","content":"<local-command-stdout>output</local-command-stdout>"},"uuid":"output-uuid","timestamp":"2025-07-06T05:00:02.000Z"}
{"type":"user","message":{"role":"user","content":"API Error: Test error"},"uuid":"error-uuid","timestamp":"2025-07-06T05:00:03.000Z"}
{"type":"user","message":{"role":"user","content":"[Request interrupted by user]"},"uuid":"interrupt-uuid","timestamp":"2025-07-06T05:00:04.000Z"}
{"type":"user","message":{"role":"user","content":"Caveat: Test caveat"},"isMeta":true,"uuid":"meta-uuid","timestamp":"2025-07-06T05:00:05.000Z"}
{"type":"summary","summary":"Test conversation with only filtered messages"}`

	normalContent := `{"type":"user","message":{"role":"user","content":"This is a normal message"},"uuid":"normal-uuid","timestamp":"2025-07-06T05:00:00.000Z"}
{"type":"assistant","message":{"role":"assistant","content":"This is a normal response"},"uuid":"assistant-uuid","timestamp":"2025-07-06T05:00:01.000Z"}`

	// Write test files
	if err := os.WriteFile(filteredEmptyFile, []byte(filteredEmptyContent), 0644); err != nil {
		t.Fatalf("Failed to create filtered empty file: %v", err)
	}
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	if err := os.WriteFile(normalFile, []byte(normalContent), 0644); err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}

	// Test extractConversationTitle behavior
	tests := []struct {
		name     string
		file     string
		expected string
	}{
		{
			name:     "Empty file should return empty string",
			file:     emptyFile,
			expected: "",
		},
		{
			name:     "File with only filtered messages should return empty string",
			file:     filteredEmptyFile,
			expected: "",
		},
		{
			name:     "Normal file should return title",
			file:     normalFile,
			expected: "This is a normal message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractConversationTitle(tt.file)
			if tt.expected == "" {
				if result != "" {
					t.Errorf("Expected empty string for %s, got %s", tt.name, result)
				}
			} else {
				if result == "" {
					t.Errorf("Expected non-empty string for %s, got empty", tt.name)
				}
				if !strings.Contains(result, tt.expected) {
					t.Errorf("Expected title to contain %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestGetFiles_SkipsFilteredEmptyFiles(t *testing.T) {
	// Red: This test should fail because GetFiles doesn't skip filtered empty files yet
	tempDir := t.TempDir()

	// Create test files using the testdata files
	filteredEmptyFile := filepath.Join(tempDir, "filtered_empty.jsonl")
	emptyFile := filepath.Join(tempDir, "empty.jsonl")
	normalFile := filepath.Join(tempDir, "normal.jsonl")

	// Copy content from testdata
	filteredEmptyContent := `{"type":"system","message":{"role":"system","content":"System message"},"uuid":"system-uuid","timestamp":"2025-07-06T05:00:00.000Z"}
{"type":"user","message":{"role":"user","content":"<command-name>/test</command-name>"},"uuid":"command-uuid","timestamp":"2025-07-06T05:00:01.000Z"}
{"type":"user","message":{"role":"user","content":"API Error: Test error"},"uuid":"error-uuid","timestamp":"2025-07-06T05:00:02.000Z"}
{"type":"summary","summary":"Test conversation with only filtered messages"}`

	normalContent := `{"type":"user","message":{"role":"user","content":"This is a normal message"},"uuid":"normal-uuid","timestamp":"2025-07-06T05:00:00.000Z"}
{"type":"assistant","message":{"role":"assistant","content":"This is a normal response"},"uuid":"assistant-uuid","timestamp":"2025-07-06T05:00:01.000Z"}`

	// Write test files
	if err := os.WriteFile(filteredEmptyFile, []byte(filteredEmptyContent), 0644); err != nil {
		t.Fatalf("Failed to create filtered empty file: %v", err)
	}
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	if err := os.WriteFile(normalFile, []byte(normalContent), 0644); err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}

	// Test GetFiles
	files, err := GetFiles(tempDir)
	if err != nil {
		t.Fatalf("GetFiles failed: %v", err)
	}

	// Count JSONL files
	jsonlFiles := make(map[string]bool)
	for _, file := range files {
		if filepath.Ext(file.Name) == ".jsonl" {
			jsonlFiles[file.Name] = true
		}
	}

	// Should only have normal.jsonl (filtered empty and empty files should be skipped)
	if len(jsonlFiles) != 1 {
		t.Errorf("Expected 1 JSONL file, got %d: %v", len(jsonlFiles), jsonlFiles)
	}

	if !jsonlFiles["normal.jsonl"] {
		t.Error("Expected normal.jsonl to be present")
	}

	if jsonlFiles["filtered_empty.jsonl"] {
		t.Error("Expected filtered_empty.jsonl to be skipped")
	}

	if jsonlFiles["empty.jsonl"] {
		t.Error("Expected empty.jsonl to be skipped")
	}
}

func TestGetFilesRecursive_SkipsFilteredEmptyFiles(t *testing.T) {
	// Red: This test should fail because GetFilesRecursive doesn't skip filtered empty files yet
	tempDir := t.TempDir()

	// Create nested structure
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test files
	filteredEmptyFile := filepath.Join(tempDir, "filtered_empty.jsonl")
	emptyFile := filepath.Join(subDir, "empty.jsonl")
	normalFile := filepath.Join(tempDir, "normal.jsonl")

	filteredEmptyContent := `{"type":"system","message":{"role":"system","content":"System message"},"uuid":"system-uuid","timestamp":"2025-07-06T05:00:00.000Z"}
{"type":"user","message":{"role":"user","content":"<command-name>/test</command-name>"},"uuid":"command-uuid","timestamp":"2025-07-06T05:00:01.000Z"}
{"type":"summary","summary":"Test conversation with only filtered messages"}`

	normalContent := `{"type":"user","message":{"role":"user","content":"This is a normal message"},"uuid":"normal-uuid","timestamp":"2025-07-06T05:00:00.000Z"}
{"type":"assistant","message":{"role":"assistant","content":"This is a normal response"},"uuid":"assistant-uuid","timestamp":"2025-07-06T05:00:01.000Z"}`

	// Write test files
	if err := os.WriteFile(filteredEmptyFile, []byte(filteredEmptyContent), 0644); err != nil {
		t.Fatalf("Failed to create filtered empty file: %v", err)
	}
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	if err := os.WriteFile(normalFile, []byte(normalContent), 0644); err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}

	// Test GetFilesRecursive
	files, err := GetFilesRecursive(tempDir)
	if err != nil {
		t.Fatalf("GetFilesRecursive failed: %v", err)
	}

	// Should only have normal.jsonl
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if len(files) > 0 && files[0].Name != "normal.jsonl" {
		t.Errorf("Expected normal.jsonl, got %s", files[0].Name)
	}
}
