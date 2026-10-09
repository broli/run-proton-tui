package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDetectionMenuView_NavigationAndRoutines(t *testing.T) {
	cfg := config.NewDefaultConfig()
	menu := NewDetectionMenuView(cfg, "/tmp", "TestGame", 100, 30)

	// 1. Initial menu view
	view := menu.View()
	if !strings.Contains(view, "DETECTION ROUTINES") {
		t.Errorf("Expected header to contain 'DETECTION ROUTINES', got:\n%s", view)
	}
	if !strings.Contains(view, "Detect Displays") {
		t.Errorf("Expected routine list to contain display detection")
	}

	// 2. Cursor navigation
	menu.Update(tea.KeyMsg{Type: tea.KeyDown})
	if menu.Cursor != 1 {
		t.Errorf("Expected cursor=1 after Down, got %d", menu.Cursor)
	}
	menu.Update(tea.KeyMsg{Type: tea.KeyUp})
	if menu.Cursor != 0 {
		t.Errorf("Expected cursor=0 after Up, got %d", menu.Cursor)
	}

	// 3. Trigger routine by number key '1' (DetectDisplay)
	menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if menu.ActiveResult == nil {
		t.Fatalf("Expected ActiveResult to be populated after pressing '1'")
	}
	if !strings.Contains(menu.ActiveResult.Title, "Display") {
		t.Errorf("Expected Display title in result, got: %s", menu.ActiveResult.Title)
	}
	if menu.ActiveResult.Source == "" {
		t.Errorf("Expected Source attribution to be non-empty")
	}

	// 4. Result view contains findings and source
	resView := menu.View()
	if !strings.Contains(resView, "Source:") {
		t.Errorf("Expected result dialog to display 'Source:', got:\n%s", resView)
	}
	if !strings.Contains(resView, "Proposed Action:") {
		t.Errorf("Expected result dialog to display 'Proposed Action:', got:\n%s", resView)
	}

	// 5. Apply result with 'y'
	done, applied := menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if done {
		t.Errorf("Expected done=false when applying finding")
	}
	if !applied {
		t.Errorf("Expected applied=true when confirming with 'y'")
	}
	if menu.ActiveResult != nil {
		t.Errorf("Expected ActiveResult to be cleared after apply")
	}
	if !cfg.GeometryAutoDetected {
		t.Errorf("Expected GeometryAutoDetected=true after applying display detection")
	}

	// 6. Test Esc key on main menu returns done=true
	done, applied = menu.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !done {
		t.Errorf("Expected done=true on Esc from main menu")
	}
	if applied {
		t.Errorf("Expected applied=false on Esc from main menu")
	}
}

func TestDetectionMenuView_CancelResultDialog(t *testing.T) {
	cfg := config.NewDefaultConfig()
	menu := NewDetectionMenuView(cfg, "/tmp", "TestGame", 100, 30)

	// Trigger CPU detection
	menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if menu.ActiveResult == nil {
		t.Fatalf("Expected ActiveResult after '3'")
	}

	// Cancel with 'n'
	done, applied := menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if done || applied {
		t.Errorf("Expected done=false, applied=false on cancel")
	}
	if menu.ActiveResult != nil {
		t.Errorf("Expected ActiveResult to be nil after cancel")
	}
}

func TestDetectionMenuView_RunAllSweep(t *testing.T) {
	cfg := config.NewDefaultConfig()
	menu := NewDetectionMenuView(cfg, "/tmp", "TestGame", 100, 30)

	// Trigger 'A' (All)
	menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if menu.ActiveResult == nil {
		t.Fatalf("Expected ActiveResult after 'a'")
	}
	if !strings.Contains(menu.ActiveResult.Title, "Sweep") {
		t.Errorf("Expected Title to mention Sweep, got: %s", menu.ActiveResult.Title)
	}

	// Apply all
	done, applied := menu.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if done || !applied {
		t.Errorf("Expected done=false, applied=true")
	}
	if !cfg.GeometryAutoDetected {
		t.Errorf("Expected GeometryAutoDetected to be set after Sweep")
	}
}
