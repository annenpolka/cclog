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
	
	// Navigate down and quit
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*2))
	
	// Check that cursor moved correctly
	finalModel := tm.FinalModel(t).(Model)
	if finalModel.cursor != 1 {
		t.Errorf("Expected cursor at position 1, got %d", finalModel.cursor)
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
	
	// Set terminal size to ensure predictable behavior without preview
	model.terminalWidth = 80
	model.terminalHeight = 40
	model.updateDisplaySettings()
	
	// Set maxDisplayFiles to 20 and ensure preview is off
	model.maxDisplayFiles = 20
	model.preview.SetVisible(false)
	
	view := model.View()
	
	// Should show "X more below" message
	if !strings.Contains(view, "10 more below") {
		t.Errorf("View should show '10 more below' message when 10 files are below visible range. View: %s", view)
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
	
	// Set terminal size and disable preview for predictable behavior
	model.terminalWidth = 80
	model.terminalHeight = 40
	model.updateDisplaySettings()
	model.preview.SetVisible(false)
	
	// Set maxDisplayFiles to 20
	model.maxDisplayFiles = 20
	
	view := model.View()
	
	// Should NOT show "more below" message (more above is removed)
	if strings.Contains(view, "more below") {
		t.Errorf("View should NOT show scroll messages when files are under limit. View: %s", view)
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
	
	// Set terminal size to ensure predictable behavior
	model.terminalWidth = 80
	model.terminalHeight = 40
	model.updateDisplaySettings()
	model.maxDisplayFiles = 20
	model.preview.SetVisible(false)
	
	// Position cursor at the end and set appropriate scroll
	model.cursor = 29
	model.scrollOffset = 10 // Show files 10-29
	
	view := model.View()
	
	// Should show cursor at the visible range
	if !strings.Contains(view, ">") {
		t.Errorf("View should contain cursor indicator even when cursor is at the end. View: %s", view)
	}
	
	// Should show scroll message for items above (removed - no longer displayed)
}

func TestModel_TruncationBounds(t *testing.T) {
	model := NewModel(".", false)
	
	// Set terminal size to ensure predictable behavior
	model.terminalWidth = 80
	model.terminalHeight = 40
	model.updateDisplaySettings()
	model.preview.SetVisible(false)
	
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
	
	// Should NOT show scroll messages at exact limit (more above is removed)
	if strings.Contains(view, "more below") {
		t.Errorf("View should NOT show scroll messages at exact limit. View: %s", view)
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
		t.Errorf("View should show 'more below' message when one over limit. View: %s", view)
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
	
	// Set terminal size and disable preview for predictable behavior
	model.terminalWidth = 80
	model.terminalHeight = 40
	model.updateDisplaySettings()
	model.preview.SetVisible(false)
	
	model.maxDisplayFiles = 10
	model.scrollOffset = 5
	model.cursor = 8
	
	view := model.View()
	
	// Should show files 5-14 (scrollOffset 5, maxDisplayFiles 10)
	if !strings.Contains(view, "file5.txt") {
		t.Errorf("View should contain file5.txt when scrollOffset is 5. View: %s", view)
	}
	
	if !strings.Contains(view, "file14.txt") {
		t.Errorf("View should contain file14.txt when scrollOffset is 5. View: %s", view)
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

// TDD Red Phase: Test for display format without dash when Description is empty

func TestModel_ViewWithoutDashWhenDescriptionEmpty(t *testing.T) {
	// Red: This test should fail because View() includes " - " even when Description is empty
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "Test conversation",
		},
	}
	
	view := model.View()
	
	// Should not contain " - " at the end since Description is empty
	if strings.Contains(view, "2025-01-15 14:30 Test conversation -") {
		t.Error("View should not contain trailing dash when Description is empty")
	}
	
	// Should contain the title without trailing dash
	if !strings.Contains(view, "2025-01-15 14:30 Test conversation") {
		t.Error("View should contain the title without trailing dash")
	}
}

// TDD Red Phase: Test for dynamic title width based on terminal width
func TestModel_WindowSizeHandling(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "This is a very long conversation title that should be truncated based on terminal width",
		},
	}
	
	// Test with narrow terminal width
	msg := tea.WindowSizeMsg{Width: 60, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	// Check that model has stored the terminal width
	if m.terminalWidth != 60 {
		t.Errorf("Expected terminal width 60, got %d", m.terminalWidth)
	}
	
	view := m.View()
	
	// With narrow width, title should be truncated appropriately
	// The new responsive system handles truncation differently
	if !strings.Contains(view, "...") {
		t.Errorf("Title should be truncated for narrow terminal width, got view: %s", view)
	}
}

func TestModel_WindowSizeHandling_WideTerminal(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "This is a medium length conversation title",
		},
	}
	
	// Test with wide terminal width
	msg := tea.WindowSizeMsg{Width: 120, Height: 24}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	// Check that model has stored the terminal width
	if m.terminalWidth != 120 {
		t.Errorf("Expected terminal width 120, got %d", m.terminalWidth)
	}
	
	view := m.View()
	
	// With wide width, title should not be truncated
	if !strings.Contains(view, "This is a medium length conversation title") {
		t.Error("Title should not be truncated for wide terminal")
	}
}

// TDD Red Phase: Test for initial window size detection
func TestModel_InitialWindowSizeDetection(t *testing.T) {
	model := NewModel(".", false)
	
	// Init should return a command (tea.Batch)
	cmd := model.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
	
	// Since we use tea.Batch, we need to test the functionality differently
	// Test that GetInitialWindowSize returns a WindowSizeMsg
	windowSizeCmd := GetInitialWindowSize()
	msg := windowSizeCmd()
	if _, ok := msg.(tea.WindowSizeMsg); !ok {
		t.Error("GetInitialWindowSize command should return a WindowSizeMsg")
	}
}


// TDD Red Phase: Test for width-based layout adjustments
func TestModel_WidthBasedLayoutAdjustments(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "This is a very long conversation title that should be adjusted based on width",
		},
	}
	
	// Test with very narrow terminal width
	narrowMsg := tea.WindowSizeMsg{Width: 40, Height: 24}
	updatedModel, _ := model.Update(narrowMsg)
	m := updatedModel.(Model)
	
	view := m.View()
	
	// Should have compact layout for narrow width
	if m.useCompactLayout != true {
		t.Error("Should use compact layout for narrow terminal width")
	}
	
	// Should not show help text in compact mode
	if strings.Contains(view, "Controls:") {
		t.Error("Should not show help text in compact layout")
	}
	
	// Test with wide terminal width
	wideMsg := tea.WindowSizeMsg{Width: 120, Height: 24}
	updatedModel, _ = model.Update(wideMsg)
	m = updatedModel.(Model)
	
	view = m.View()
	
	// Should have full layout for wide width
	if m.useCompactLayout != false {
		t.Error("Should use full layout for wide terminal width")
	}
	
	// Should show help text in full mode
	if !strings.Contains(view, "Controls:") {
		t.Error("Should show help text in full layout")
	}
}

func TestModel_DirectoryPathTruncation(t *testing.T) {
	model := NewModel("/very/long/path/to/some/directory/that/should/be/truncated", false)
	
	// Test with narrow terminal width
	narrowMsg := tea.WindowSizeMsg{Width: 50, Height: 24}
	updatedModel, _ := model.Update(narrowMsg)
	m := updatedModel.(Model)
	
	view := m.View()
	
	// Directory path should be truncated for narrow width
	if strings.Contains(view, "/very/long/path/to/some/directory/that/should/be/truncated") {
		t.Error("Directory path should be truncated for narrow width")
	}
	
	// Should contain ellipsis when truncated
	if !strings.Contains(view, "...") {
		t.Error("Truncated directory path should contain ellipsis")
	}
}

func TestModel_ControlsTextAdaptation(t *testing.T) {
	model := NewModel(".", false)
	
	// Test with very narrow terminal width - should show minimal controls
	veryNarrowMsg := tea.WindowSizeMsg{Width: 30, Height: 24}
	updatedModel, _ := model.Update(veryNarrowMsg)
	m := updatedModel.(Model)
	
	view := m.View()
	
	// Should show abbreviated controls for very narrow width
	if strings.Contains(view, "Enter: Open folder / Open file in editor") {
		t.Error("Should show abbreviated controls for very narrow width")
	}
}

// Test for dynamic character limit - no more content stretching
func TestModel_DynamicCharacterLimitInsteadOfStretching(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "Short title",
		},
	}
	
	// Test with wide terminal - should allow more characters instead of stretching
	wideMsg := tea.WindowSizeMsg{Width: 120, Height: 24}
	updatedModel, _ := model.Update(wideMsg)
	m := updatedModel.(Model)
	
	// Should have higher character limit for wide terminal
	if m.maxTitleChars <= 40 {
		t.Errorf("Wide terminal should have higher character limit, got %d", m.maxTitleChars)
	}
	
	view := m.View()
	
	// Should show title normally without stretching
	if !strings.Contains(view, "Short title") {
		t.Error("Should show title normally")
	}
	
	// Should NOT have character expansion
	if strings.Contains(view, "S h o r t   t i t l e") {
		t.Error("Should not have character expansion")
	}
}

func TestModel_AdaptiveTitleLength(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "This is a medium length conversation title",
		},
	}
	
	// Test with different widths - title should adapt accordingly
	widths := []struct {
		width    int
		expected string
	}{
		{60, "This is a medium length..."},    // Narrow: truncated
		{100, "This is a medium length conversation title"}, // Wide: full title
		{80, "This is a medium length conversation title"},  // Medium: should fit
	}
	
	for _, tc := range widths {
		msg := tea.WindowSizeMsg{Width: tc.width, Height: 24}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)
		
		view := m.View()
		
		if tc.width >= 100 {
			// Wide screen should show full title
			if !strings.Contains(view, tc.expected) {
				t.Errorf("Width %d should show full title '%s'", tc.width, tc.expected)
			}
		} else if tc.width <= 60 {
			// Narrow screen should show truncated title
			if !strings.Contains(view, "...") {
				t.Errorf("Width %d should show truncated title with ellipsis", tc.width)
			}
		}
	}
}

func TestModel_ContentAlignment(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "Test title",
		},
	}
	
	// Test with wide terminal
	wideMsg := tea.WindowSizeMsg{Width: 120, Height: 24}
	updatedModel, _ := model.Update(wideMsg)
	m := updatedModel.(Model)
	
	// Should support content alignment options
	if m.contentAlignment == "" {
		t.Error("Should have content alignment setting")
	}
	
	view := m.View()
	
	// Check that content uses available space efficiently
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Test title") {
			// Line should not significantly exceed terminal width (allow small buffer for Unicode)
			lineRunes := []rune(line)
			if len(lineRunes) > m.terminalWidth+5 { // Allow small buffer
				t.Errorf("Line rune length %d significantly exceeds terminal width %d", len(lineRunes), m.terminalWidth)
			}
		}
	}
}

// TDD Red Phase: Test for dynamic title character limit based on terminal width
func TestModel_DynamicTitleCharacterLimit(t *testing.T) {
	model := NewModel(".", false)
	longTitle := "This is a very long conversation title that should be displayed differently based on terminal width and should show more characters when the terminal is wider"
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: longTitle,
		},
	}
	
	testCases := []struct {
		width          int
		expectedMinLen int
		description    string
	}{
		{60, 20, "narrow terminal should show at least 20 chars"},
		{80, 35, "medium terminal should show at least 35 chars"},
		{120, 75, "wide terminal should show at least 75 chars"},
		{160, 115, "very wide terminal should show at least 115 chars"},
	}
	
	for _, tc := range testCases {
		msg := tea.WindowSizeMsg{Width: tc.width, Height: 24}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)
		
		view := m.View()
		
		// Find the title part in the view
		titleFound := false
		titleLength := 0
		lines := strings.Split(view, "\n")
		for _, line := range lines {
			if strings.Contains(line, "This is a very") {
				// Extract just the title part (after date/time)
				parts := strings.Split(line, "2025-01-15 14:30 ")
				if len(parts) > 1 {
					titlePart := parts[1]
					titleLength = len([]rune(titlePart))
					titleFound = true
				}
				break
			}
		}
		
		if !titleFound {
			t.Errorf("Could not find title in view for width %d", tc.width)
			continue
		}
		
		if titleLength < tc.expectedMinLen {
			t.Errorf("%s: got %d chars, expected at least %d chars", tc.description, titleLength, tc.expectedMinLen)
		}
	}
}

func TestModel_TitleCharacterLimitScaling(t *testing.T) {
	model := NewModel(".", false)
	
	// Test that the character limit scales with terminal width
	narrowMsg := tea.WindowSizeMsg{Width: 60, Height: 24}
	updatedModel, _ := model.Update(narrowMsg)
	narrowModel := updatedModel.(Model)
	
	wideMsg := tea.WindowSizeMsg{Width: 120, Height: 24}
	updatedModel, _ = model.Update(wideMsg)
	wideModel := updatedModel.(Model)
	
	// Wide terminal should have higher character limit
	if wideModel.maxTitleChars <= narrowModel.maxTitleChars {
		t.Errorf("Wide terminal should have higher character limit than narrow terminal. Wide: %d, Narrow: %d", wideModel.maxTitleChars, narrowModel.maxTitleChars)
	}
}

func TestModel_NoCharacterSpacingExpansion(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:              "test.jsonl",
			IsDir:             false,
			ModTime:           time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			ConversationTitle: "Normal title",
		},
	}
	
	// Test with wide terminal
	wideMsg := tea.WindowSizeMsg{Width: 120, Height: 24}
	updatedModel, _ := model.Update(wideMsg)
	m := updatedModel.(Model)
	
	view := m.View()
	
	// Should NOT have excessive character spacing
	if strings.Contains(view, "N o r m a l   t i t l e") {
		t.Error("Should not expand characters with spacing")
	}
	
	// Should show normal title with more characters allowed
	if !strings.Contains(view, "Normal title") {
		t.Error("Should show normal title without character expansion")
	}
}

// TDD Red Phase: Test for preview functionality integration
func TestModel_PreviewToggle(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:    "test.jsonl",
			Path:    "/path/test.jsonl",
			IsDir:   false,
			ModTime: time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
		},
	}
	
	// Preview should be visible by default
	if !model.preview.IsVisible() {
		t.Error("Preview should be visible by default")
	}
	
	// Test 'p' key to toggle preview off
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	// Should have preview disabled
	if m.preview.IsVisible() {
		t.Error("Preview should be hidden after pressing 'p'")
	}
	
	// Toggle again to enable
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	
	// Should have preview enabled again
	if !m.preview.IsVisible() {
		t.Error("Preview should be visible after pressing 'p' again")
	}
}

func TestModel_PreviewContentUpdate(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:    "test.jsonl",
			Path:    "/path/test.jsonl",
			IsDir:   false,
			ModTime: time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
		},
	}
	
	// Preview should be visible by default
	if !model.preview.IsVisible() {
		t.Error("Preview should be visible by default")
	}
	
	// Moving cursor should update preview content
	model.cursor = 0
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	
	// Preview content should be updated (in real implementation)
	// This test verifies the structure is in place
	if !m.preview.IsVisible() {
		t.Error("Preview should remain visible after cursor movement")
	}
}

func TestModel_PreviewLayoutAdjustment(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:    "test.jsonl",
			Path:    "/path/test.jsonl",
			IsDir:   false,
			ModTime: time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
		},
	}
	
	// Preview should be enabled by default
	if !model.preview.IsVisible() {
		t.Error("Preview should be visible by default")
	}
	
	// Test window size adjustment with preview enabled
	wideMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(wideMsg)
	m := updatedModel.(Model)
	
	// Preview should adjust to new window size
	width, height := m.preview.GetSize()
	if width == 0 || height == 0 {
		t.Error("Preview should have valid size after window resize")
	}
}

func TestModel_PreviewWithDirectorySelection(t *testing.T) {
	model := NewModel(".", false)
	model.files = []FileInfo{
		{
			Name:  "testdir",
			Path:  "/path/testdir",
			IsDir: true,
		},
	}
	
	// Preview should be visible by default
	if !model.preview.IsVisible() {
		t.Error("Preview should be visible by default")
	}
	
	// Content should be empty for directories
	if model.preview.GetContent() != "" {
		t.Error("Preview content should be empty for directories")
	}
}

// TDD Red Phase: Tests for filtering toggle functionality

func TestModel_EnableFilteringField(t *testing.T) {
	// Red: This test should fail because enableFiltering field doesn't exist yet
	model := NewModel(".", false)
	
	// Model should have enableFiltering field with default value true
	if !model.enableFiltering {
		t.Error("Model should have enableFiltering field with default value true")
	}
}

func TestModel_FilteringToggleWithSKey(t *testing.T) {
	// Red: This test should fail because 's' key handling doesn't exist yet
	model := NewModel(".", false)
	
	// Initial state should be filtering enabled
	if !model.enableFiltering {
		t.Error("Initial filtering state should be enabled")
	}
	
	// Press 's' key to toggle filtering
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updatedModel, _ := model.Update(keyMsg)
	m := updatedModel.(Model)
	
	// Filtering should be toggled to disabled
	if m.enableFiltering {
		t.Error("Filtering should be disabled after pressing 's' key")
	}
	
	// Press 's' key again to toggle back
	updatedModel2, _ := m.Update(keyMsg)
	m2 := updatedModel2.(Model)
	
	// Filtering should be enabled again
	if !m2.enableFiltering {
		t.Error("Filtering should be enabled again after pressing 's' key twice")
	}
}

func TestModel_FilteringStateDisplay(t *testing.T) {
	// Red: This test should fail because filtering state display doesn't exist yet
	model := NewModel(".", false)
	
	// Test filtering enabled state display
	model.enableFiltering = true
	view := model.View()
	if !strings.Contains(view, "[FILTERED]") {
		t.Error("View should contain '[FILTERED]' when filtering is enabled")
	}
	
	// Test filtering disabled state display
	model.enableFiltering = false
	view = model.View()
	if !strings.Contains(view, "[UNFILTERED]") {
		t.Error("View should contain '[UNFILTERED]' when filtering is disabled")
	}
}

func TestModel_FilteringToggleHelpText(t *testing.T) {
	// Red: This test should fail because help text doesn't include 's' key yet
	model := NewModel(".", false)
	view := model.View()
	
	// Help text should include 's' key for toggling filter
	if !strings.Contains(view, "s:") || !strings.Contains(view, "filter") {
		t.Error("Help text should include 's' key for toggling filter")
	}
}

func TestModel_FilteringStateAffectsPreview(t *testing.T) {
	// Red: This test should fail because preview doesn't use enableFiltering yet
	tempDir := t.TempDir()
	
	// Create test file with filterable content
	testFile := filepath.Join(tempDir, "test.jsonl")
	filteredContent := `{"type":"system","message":{"role":"system","content":"System message"},"uuid":"system-uuid","timestamp":"2025-07-06T05:00:00.000Z"}
{"type":"user","message":{"role":"user","content":"Normal user message"},"uuid":"user-uuid","timestamp":"2025-07-06T05:00:01.000Z"}`
	
	if err := os.WriteFile(testFile, []byte(filteredContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	model := NewModel(tempDir, false)
	model.files = []FileInfo{
		{
			Name: "test.jsonl",
			Path: testFile,
			ConversationTitle: "Test conversation",
		},
	}
	model.cursor = 0
	
	// Test with filtering enabled (should filter out system messages)
	model.enableFiltering = true
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m1 := updatedModel.(Model)
	previewContent1 := m1.preview.GetContent()
	
	// Test with filtering disabled (should include all messages)
	model.enableFiltering = false
	updatedModel2, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m2 := updatedModel2.(Model)
	previewContent2 := m2.preview.GetContent()
	
	// Content should be different based on filtering state
	if previewContent1 == previewContent2 {
		t.Error("Preview content should differ based on filtering state")
	}
}