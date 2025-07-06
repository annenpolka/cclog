package filepicker

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestModel_Init(t *testing.T) {
	model := NewModel(".")
	cmd := model.Init()
	
	if cmd == nil {
		t.Error("Init() should return a command to load files")
	}
}

func TestModel_QuitWithQ(t *testing.T) {
	model := NewModel(".")
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
	model := NewModel(".")
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
	model := NewModel(".")
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
	model := NewModel(".")
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
	model := NewModel(".")
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
	model := NewModel(".")
	model.files = []FileInfo{
		{Name: "file1.txt", Path: "/path/file1.txt", IsDir: false},
		{Name: "file2.txt", Path: "/path/file2.txt", IsDir: false},
	}
	model.cursor = 1
	
	// Test enter key selection
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	m := updatedModel.(Model)
	
	if m.selected != "/path/file2.txt" {
		t.Errorf("Expected selected file to be '/path/file2.txt', got '%s'", m.selected)
	}
	
	if cmd == nil {
		t.Error("Expected tea.Quit command after selection")
	}
}

func TestModel_EmptyFileList(t *testing.T) {
	model := NewModel(".")
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
	model := NewModel(".")
	model.files = []FileInfo{
		{Name: "file1.txt", Path: "/tmp/file1.txt", IsDir: false},
		{Name: "file2.txt", Path: "/tmp/file2.txt", IsDir: false},
		{Name: "dir1", Path: "/tmp/dir1", IsDir: true},
	}
	
	tm := teatest.NewTestModel(t, model)
	
	// Navigate down twice and select
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*2))
	
	finalModel := tm.FinalModel(t).(Model)
	if finalModel.selected != "/tmp/dir1" {
		t.Errorf("Expected selection '/tmp/dir1', got '%s'", finalModel.selected)
	}
}

func TestModel_FilesLoadedMessage(t *testing.T) {
	model := NewModel(".")
	
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
	model := NewModel(".")
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
	model := NewModel(".")
	selectedPath := model.GetSelectedFile()
	if selectedPath != "" {
		t.Errorf("Expected empty string for no selection, got '%s'", selectedPath)
	}
}