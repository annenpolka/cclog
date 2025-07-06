package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestModel_Init(t *testing.T) {
	model := NewModel(".", false)
	cmd := model.Init()
	
	if cmd == nil {
		t.Error("Init() should return a command to load files")
	}
}

func TestModel_QuitWithQ(t *testing.T) {
	model := NewModel(".", false)
	tm := teatest.NewTestModel(t, model)
	
	// Type 'q' to quit
	tm.Type("q")
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
	
	// Check that program exited cleanly
	if tm.FinalModel(t) == nil {
		t.Error("Program should have exited cleanly")
	}
}

func TestModel_CursorMovement(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file1.txt", IsDir: false},
		{Name: "file2.txt", IsDir: false},
		{Name: "file3.txt", IsDir: false},
	}
	
	// Test cursor down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	if m.cursor != 1 {
		t.Errorf("Expected cursor to be 1, got %d", m.cursor)
	}
	
	// Test cursor up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	
	if m.cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", m.cursor)
	}
}

func TestModel_View(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file1.txt", IsDir: false},
		{Name: "dir1", IsDir: true},
	}
	
	view := model.View()
	
	if view == "" {
		t.Error("View() should return non-empty string")
	}
	
	// Check if cursor indicator is present
	if !strings.Contains(view, ">") {
		t.Error("View should contain cursor indicator '>'")
	}
}

func TestModel_CursorBounds(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file1.txt", IsDir: false},
		{Name: "file2.txt", IsDir: false},
	}
	
	// Test cursor doesn't go below 0
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	if m.cursor != 0 {
		t.Errorf("Cursor should stay at 0, got %d", m.cursor)
	}
	
	// Move cursor to last item
	m.cursor = len(m.files) - 1
	
	// Test cursor doesn't go beyond last item
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	
	if m.cursor != len(m.files)-1 {
		t.Errorf("Cursor should stay at last position %d, got %d", len(m.files)-1, m.cursor)
	}
}

func TestModel_VimStyleKeys(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file1.txt", IsDir: false},
		{Name: "file2.txt", IsDir: false},
		{Name: "file3.txt", IsDir: false},
	}
	
	// Test 'j' key (down)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	if m.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after pressing 'j', got %d", m.cursor)
	}
	
	// Test 'k' key (up)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	
	if m.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after pressing 'k', got %d", m.cursor)
	}
}

func TestModel_EnterSelection(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file1.txt", Path: "/path/file1.txt", IsDir: false},
		{Name: "file2.txt", Path: "/path/file2.txt", IsDir: false},
	}
	model.cursor = 1
	
	// Test enter key selection - should now open editor
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should not set selected file when opening editor
	if m.selected != "" {
		t.Errorf("Expected no selection when opening editor, got '%s'", m.selected)
	}
	
	if cmd == nil {
		t.Error("Expected editor command after enter on file")
	}
}

func TestModel_EmptyFileList(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{} // empty list
	
	// Test enter key on empty list
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	if m.selected != "" {
		t.Errorf("Expected no selection on empty list, got '%s'", m.selected)
	}
	
	if cmd != nil {
		t.Error("Should not quit on empty list enter")
	}
}

func TestModel_Integration_WithTeatest(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file1.txt", Path: "/tmp/file1.txt", IsDir: false},
		{Name: "file2.txt", Path: "/tmp/file2.txt", IsDir: false},
	}
	
	tm := teatest.NewTestModel(t, model)
	
	// Navigate down and select file with space
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*2))
	
	finalModel := tm.FinalModel(t).(Model)
	if finalModel.selected != "/tmp/file2.txt" {
		t.Errorf("Expected selection '/tmp/file2.txt', got '%s'", finalModel.selected)
	}
}

func TestModel_FilesLoadedMessage(t *testing.T) {
	model := NewModel(".", false)
	
	testFiles := []FileInfo{
		{Name: "test1.txt", IsDir: false},
		{Name: "test2.txt", IsDir: false},
	}
	
	msg := filesLoadedMsg{files: testFiles}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	if len(m.files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(m.files))
	}
	
	if m.files[0].Name != "test1.txt" {
		t.Errorf("Expected first file 'test1.txt', got '%s'", m.files[0].Name)
	}
}

func TestModel_GetSelectedFile(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file1.txt", Path: "/path/file1.txt", IsDir: false},
		{Name: "file2.txt", Path: "/path/file2.txt", IsDir: false},
	}
	model.cursor = 1
	model.selected = "/path/file2.txt"
	
	selectedPath := model.GetSelectedFile()
	if selectedPath != "/path/file2.txt" {
		t.Errorf("Expected '/path/file2.txt', got '%s'", selectedPath)
	}
}

func TestModel_GetSelectedFile_NoSelection(t *testing.T) {
	model := NewModel(".", false)
	selectedPath := model.GetSelectedFile()
	if selectedPath != "" {
		t.Errorf("Expected empty string for no selection, got '%s'", selectedPath)
	}
}

func TestModel_DirectoryNavigation(t *testing.T) {
	// Create test directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	model := NewModel(tempDir, false)
	model.files = []FileInfo{
		{Name: "subdir", Path: subDir, IsDir: true},
		{Name: "file.txt", Path: filepath.Join(tempDir, "file.txt"), IsDir: false},
	}
	
	// Test navigating into directory
	model.cursor = 0 // Select the directory
	
	// Simulate enter key on directory
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should navigate into directory and return load command
	if cmd == nil {
		t.Error("Should return load command when entering directory")
	}
	
	// Directory should change to subdirectory
	if m.dir != subDir {
		t.Errorf("Expected directory to change to '%s', got '%s'", subDir, m.dir)
	}
	
	// Cursor should reset to 0
	if m.cursor != 0 {
		t.Errorf("Expected cursor to reset to 0, got %d", m.cursor)
	}
}

func TestModel_FileSelection(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file.txt", Path: "/path/file.txt", IsDir: false},
	}
	model.cursor = 0
	
	// Simulate enter key on file - should now open in editor
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should return editor command (not quit directly)
	if cmd == nil {
		t.Error("Should return editor command when selecting file")
	}
	
	// Selected file should not be set when opening in editor
	if m.selected != "" {
		t.Errorf("Expected no selection when opening in editor, got '%s'", m.selected)
	}
}

func TestModel_BackNavigation(t *testing.T) {
	// Create test directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Start from subdirectory
	model := NewModel(subDir, false)
	model.files = []FileInfo{
		{Name: "..", Path: tempDir, IsDir: true},
		{Name: "file.txt", Path: filepath.Join(subDir, "file.txt"), IsDir: false},
	}
	
	// Test navigating back with ..
	model.cursor = 0 // Select ".."
	
	// Simulate enter key on ".."
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should navigate back and return load command
	if cmd == nil {
		t.Error("Should return load command when navigating back")
	}
	
	// Directory should change back to parent
	if m.dir != tempDir {
		t.Errorf("Expected directory to change back to '%s', got '%s'", tempDir, m.dir)
	}
}

func TestModel_BackNavigationFromRoot(t *testing.T) {
	// Test that ".." doesn't appear at filesystem root
	// Simulate loading files (normally done by loadFiles command)
	files, _ := GetFiles("/")
	
	// Check that ".." is not in the list
	for _, file := range files {
		if file.Name == ".." {
			t.Error("'..' should not appear when at root directory")
		}
	}
}

func TestModel_SpaceKeyFileSelection(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "file.txt", Path: "/path/file.txt", IsDir: false},
	}
	model.cursor = 0
	
	// Simulate space key on file
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should select file and quit
	if cmd == nil {
		t.Error("Should return quit command when selecting file with space")
	}
	
	if m.selected != "/path/file.txt" {
		t.Errorf("Expected selected file '/path/file.txt', got '%s'", m.selected)
	}
}

func TestModel_SpaceKeyOnDirectory(t *testing.T) {
	// Create test directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	model := NewModel(tempDir, false)
	model.files = []FileInfo{
		{Name: "subdir", Path: subDir, IsDir: true},
	}
	model.cursor = 0
	
	// Simulate space key on directory - should NOT navigate (different from Enter)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should NOT navigate into directory with space
	if cmd != nil {
		t.Error("Should not return command when pressing space on directory")
	}
	
	// Directory should not change
	if m.dir != tempDir {
		t.Error("Directory should not change when pressing space on directory")
	}
}

func TestModel_EditorOpening(t *testing.T) {
	// Test that editor command is properly created
	model := NewModel(".", false)
	model.files = []FileInfo{
		{Name: "test.txt", Path: "/tmp/test.txt", IsDir: false},
	}
	model.cursor = 0
	
	// Simulate enter key on file
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should return editor command
	if cmd == nil {
		t.Error("Should return editor command when pressing enter on file")
	}
	
	// Should not set selected file
	if m.selected != "" {
		t.Errorf("Expected no selection when opening editor, got '%s'", m.selected)
	}
}

func TestModel_FileTruncation(t *testing.T) {
	model := NewModel(".", false)
	
	// Create large file list (30 files)
	files := make([]FileInfo, 30)
	for i := 0; i < 30; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	
	// Set maxDisplayFiles to 20
	model.maxDisplayFiles = 20
	
	view := model.View()
	
	// Should show "X more below" message
	if !strings.Contains(view, "10 more below") {
		t.Error("View should show '10 more below' message when 10 files are below visible range")
	}
}

func TestModel_NoTruncationWhenFilesUnderLimit(t *testing.T) {
	model := NewModel(".", false)
	
	// Create small file list (10 files)
	files := make([]FileInfo, 10)
	for i := 0; i < 10; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	
	// Set maxDisplayFiles to 20
	model.maxDisplayFiles = 20
	
	view := model.View()
	
	// Should NOT show "more above/below" message
	if strings.Contains(view, "more above") || strings.Contains(view, "more below") {
		t.Error("View should NOT show scroll messages when files are under limit")
	}
}

func TestModel_TruncationWithCursorPositioning(t *testing.T) {
	model := NewModel(".", false)
	
	// Create large file list (30 files)
	files := make([]FileInfo, 30)
	for i := 0; i < 30; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	model.maxDisplayFiles = 20
	
	// Position cursor at the end and set appropriate scroll
	model.cursor = 29
	model.scrollOffset = 10 // Show files 10-29
	
	view := model.View()
	
	// Should show cursor at the visible range
	if !strings.Contains(view, ">") {
		t.Error("View should contain cursor indicator even when cursor is at the end")
	}
	
	// Should show scroll message for items above
	if !strings.Contains(view, "more above") {
		t.Error("View should show 'more above' message when cursor is at the end")
	}
}

func TestModel_TruncationBounds(t *testing.T) {
	model := NewModel(".", false)
	
	// Test with exact limit
	files := make([]FileInfo, 20)
	for i := 0; i < 20; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	model.maxDisplayFiles = 20
	
	view := model.View()
	
	// Should NOT show scroll messages at exact limit
	if strings.Contains(view, "more above") || strings.Contains(view, "more below") {
		t.Error("View should NOT show scroll messages at exact limit")
	}
	
	// Test with one over limit
	files = append(files, FileInfo{
		Name:  "file20.txt",
		Path:  "/tmp/file20.txt",
		IsDir: false,
	})
	model.files = files
	
	view = model.View()
	
	// Should show scroll message when over limit
	if !strings.Contains(view, "more below") {
		t.Error("View should show 'more below' message when one over limit")
	}
}

func TestModel_ScrollDown(t *testing.T) {
	model := NewModel(".", false)
	
	// Create 30 files
	files := make([]FileInfo, 30)
	for i := 0; i < 30; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	model.maxDisplayFiles = 10
	
	// Move cursor to position 9 (last visible item)
	model.cursor = 9
	
	// Move down - should scroll the view
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	// Cursor should be at position 10
	if m.cursor != 10 {
		t.Errorf("Expected cursor at position 10, got %d", m.cursor)
	}
	
	// Scroll offset should be 1
	if m.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be 1, got %d", m.scrollOffset)
	}
}

func TestModel_ScrollUp(t *testing.T) {
	model := NewModel(".", false)
	
	// Create 30 files
	files := make([]FileInfo, 30)
	for i := 0; i < 30; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	model.maxDisplayFiles = 10
	model.scrollOffset = 5
	model.cursor = 5 // At top of visible range
	
	// Move up - should scroll the view
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	// Cursor should be at position 4
	if m.cursor != 4 {
		t.Errorf("Expected cursor at position 4, got %d", m.cursor)
	}
	
	// Scroll offset should be 4
	if m.scrollOffset != 4 {
		t.Errorf("Expected scrollOffset to be 4, got %d", m.scrollOffset)
	}
}

func TestModel_ScrollView(t *testing.T) {
	model := NewModel(".", false)
	
	// Create 30 files
	files := make([]FileInfo, 30)
	for i := 0; i < 30; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	model.maxDisplayFiles = 10
	model.scrollOffset = 5
	model.cursor = 8
	
	view := model.View()
	
	// Should show files 5-14 (scrollOffset 5, maxDisplayFiles 10)
	if !strings.Contains(view, "file5.txt") {
		t.Error("View should contain file5.txt when scrollOffset is 5")
	}
	
	if !strings.Contains(view, "file14.txt") {
		t.Error("View should contain file14.txt when scrollOffset is 5")
	}
	
	// Should NOT show file4.txt (before scroll range)
	if strings.Contains(view, "file4.txt") {
		t.Error("View should NOT contain file4.txt when scrollOffset is 5")
	}
	
	// Should show cursor at position 8
	if !strings.Contains(view, "> file8.txt") {
		t.Error("View should show cursor at file8.txt")
	}
}

func TestModel_ScrollBounds(t *testing.T) {
	model := NewModel(".", false)
	
	// Create 30 files
	files := make([]FileInfo, 30)
	for i := 0; i < 30; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	model.maxDisplayFiles = 10
	
	// Test scroll at top - should not scroll up
	model.cursor = 0
	model.scrollOffset = 0
	
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	if m.cursor != 0 {
		t.Errorf("Cursor should stay at 0, got %d", m.cursor)
	}
	
	if m.scrollOffset != 0 {
		t.Errorf("ScrollOffset should stay at 0, got %d", m.scrollOffset)
	}
	
	// Test scroll at bottom - should not scroll down
	model.cursor = 29
	model.scrollOffset = 20 // Max scroll position for 30 files, 10 display
	
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = model.Update(msg)
	m = updatedModel.(Model)
	
	if m.cursor != 29 {
		t.Errorf("Cursor should stay at 29, got %d", m.cursor)
	}
	
	if m.scrollOffset != 20 {
		t.Errorf("ScrollOffset should stay at 20, got %d", m.scrollOffset)
	}
}

func TestModel_ScrollReset(t *testing.T) {
	model := NewModel(".", false)
	
	// Create 30 files
	files := make([]FileInfo, 30)
	for i := 0; i < 30; i++ {
		files[i] = FileInfo{
			Name:  fmt.Sprintf("file%d.txt", i),
			Path:  fmt.Sprintf("/tmp/file%d.txt", i),
			IsDir: false,
		}
	}
	model.files = files
	model.maxDisplayFiles = 10
	model.scrollOffset = 5
	model.cursor = 15
	
	// Load new files - should reset scroll
	msg := filesLoadedMsg{files: files[:10]} // Only 10 files now
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should reset scroll and cursor
	if m.scrollOffset != 0 {
		t.Errorf("ScrollOffset should reset to 0, got %d", m.scrollOffset)
	}
	
	if m.cursor != 0 {
		t.Errorf("Cursor should reset to 0, got %d", m.cursor)
	}
}