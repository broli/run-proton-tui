package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/quirks"
	tea "github.com/charmbracelet/bubbletea"
)

func TestPresetPickerView(t *testing.T) {
	preset := &quirks.Preset{
		Name:          "Arknights: Endfield",
		MatchedSource: "Curated Game Quirks Engine",
		SummaryNotes:  "Redirects IL2CPP JIT files to /dev/shm",
		UmuID:         "umu-endfield",
		EnvVars: map[string]string{
			"WINE_CANONICAL_HOLE": "skip_volatile_check",
		},
		ExtraArgs:     []string{"-vulkan"},
		WaitProcesses: []string{"Updater.exe", "Games.exe"},
	}

	cfg := config.NewDefaultConfig()
	view := NewPresetPickerView(preset, "GRYPHLINK", cfg, 100, 30)

	rendered := view.View()
	if !strings.Contains(rendered, "Arknights: Endfield") {
		t.Errorf("Expected rendered view to contain game title")
	}
	if !strings.Contains(rendered, "umu-endfield") {
		t.Errorf("Expected rendered view to contain UMU ID")
	}
	if !strings.Contains(rendered, "WINE_CANONICAL_HOLE") {
		t.Errorf("Expected rendered view to contain custom env var")
	}

	// Test Enter key
	applied, cancel := view.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !applied || cancel {
		t.Errorf("Expected applied=true, cancel=false on Enter")
	}

	// Test Esc key
	applied, cancel = view.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if applied || !cancel {
		t.Errorf("Expected applied=false, cancel=true on Esc")
	}
}

func TestExePickerFilterInstallers(t *testing.T) {
	items := []ExeItem{
		{RelativePath: "Game.exe"},
		{RelativePath: "Setup.exe"},
		{RelativePath: "unins000.exe"},
	}

	picker := NewExePickerView(items, "Game.exe")
	if len(picker.FilteredItems) != 3 {
		t.Errorf("Expected 3 items initially, got %d", len(picker.FilteredItems))
	}

	// Toggle filter
	picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if len(picker.FilteredItems) != 1 || picker.FilteredItems[0].RelativePath != "Game.exe" {
		t.Errorf("Expected only Game.exe when filtered, got %v", picker.FilteredItems)
	}

	// Toggle back
	picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if len(picker.FilteredItems) != 3 {
		t.Errorf("Expected 3 items after unfiltering, got %d", len(picker.FilteredItems))
	}
}
