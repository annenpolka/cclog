package cli

import (
	"testing"
	"time"

	"cclog/pkg/filepicker"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestRunTUI_WithTeatest(t *testing.T) {
	// Test the filepicker model directly using teatest
	model := filepicker.NewModel(".")
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	
	// Type 'q' to quit immediately for testing
	tm.Type("q")
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
	
	// Verify the model exists and test completed
	finalModel := tm.FinalModel(t)
	if finalModel == nil {
		t.Error("Expected final model to exist")
	}
}