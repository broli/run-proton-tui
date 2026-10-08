package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestProfilesView_NavigationAndActivation(t *testing.T) {
	fileCfg := config.NewConfigFileWithDefault()
	fileCfg.GetActiveProfile().TargetExe = "Game.exe"
	fileCfg.AddProfile("launcher", &config.GameConfig{
		TargetExe: "Launcher.exe",
	})

	pv := NewProfilesView(fileCfg, 80, 24)
	if pv == nil {
		t.Fatal("expected non-nil ProfilesView")
	}

	// Active profile was "default"
	if fileCfg.ActiveProfile != "default" {
		t.Fatalf("expected active profile to be default, got %q", fileCfg.ActiveProfile)
	}

	// Move down
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if pv.Cursor != 1 {
		t.Fatalf("expected cursor to be 1, got %d", pv.Cursor)
	}

	// Press enter to activate profile
	changed, modified, exit := pv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !changed || !modified {
		t.Fatalf("expected changed and modified to be true, got changed=%v modified=%v", changed, modified)
	}
	if exit {
		t.Fatal("expected exit to be false on enter")
	}
	if fileCfg.ActiveProfile != "launcher" {
		t.Fatalf("expected active profile to be launcher, got %q", fileCfg.ActiveProfile)
	}

	// Test Esc to exit
	_, _, exit = pv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !exit {
		t.Fatal("expected exit to be true on Esc")
	}
}

func TestProfilesView_CreateAndClone(t *testing.T) {
	fileCfg := config.NewConfigFileWithDefault()
	fileCfg.GetActiveProfile().TargetExe = "Game.exe"
	fileCfg.GetActiveProfile().UseGamescope = true
	fileCfg.GetActiveProfile().GamescopeWidth = 1920

	pv := NewProfilesView(fileCfg, 80, 24)

	// Press 'n' to enter new mode
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if pv.Mode != profileModeNew {
		t.Fatalf("expected mode to be profileModeNew, got %v", pv.Mode)
	}

	// Set input value
	pv.NameInput.SetValue("benchmark")

	// Press Enter to confirm
	changed, modified, _ := pv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !changed || !modified {
		t.Fatalf("expected changed and modified to be true, got changed=%v modified=%v", changed, modified)
	}

	if fileCfg.ActiveProfile != "benchmark" {
		t.Fatalf("expected active profile to be benchmark, got %q", fileCfg.ActiveProfile)
	}

	cloned := fileCfg.GetActiveProfile()
	if cloned == nil || cloned.TargetExe != "Game.exe" || !cloned.UseGamescope {
		t.Fatalf("expected cloned profile to match base, got %+v", cloned)
	}
}

func TestProfilesView_Delete(t *testing.T) {
	fileCfg := config.NewConfigFileWithDefault()
	fileCfg.GetActiveProfile().TargetExe = "Game.exe"
	fileCfg.AddProfile("extra", &config.GameConfig{
		TargetExe: "Extra.exe",
	})

	pv := NewProfilesView(fileCfg, 80, 24)
	names := fileCfg.ProfileNames()
	for i, name := range names {
		if name == "extra" {
			pv.Cursor = i
			break
		}
	}

	// Press 'x' to initiate delete
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if pv.Mode != profileModeConfirmDelete {
		t.Fatalf("expected mode profileModeConfirmDelete, got %v", pv.Mode)
	}

	// Confirm with 'y'
	_, modified, _ := pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !modified {
		t.Fatal("expected modified=true after delete")
	}

	if _, exists := fileCfg.Profiles["extra"]; exists {
		t.Fatal("expected profile extra to be deleted")
	}

	// Attempt to delete remaining profile (only 1 profile left)
	pv.Cursor = 0
	pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if pv.Mode == profileModeConfirmDelete {
		t.Fatal("should not allow deleting the last remaining profile")
	}
	if !strings.Contains(pv.StatusMessage, "Cannot delete") {
		t.Fatalf("expected status message about cannot delete, got %q", pv.StatusMessage)
	}
}

func TestProfilesView_Render(t *testing.T) {
	fileCfg := config.NewConfigFileWithDefault()
	fileCfg.GetActiveProfile().TargetExe = "Game.exe"

	pv := NewProfilesView(fileCfg, 80, 24)
	out := pv.View()
	if !strings.Contains(out, "PROFILES MANAGER") {
		t.Fatalf("expected view to contain header, got: %s", out)
	}
	if !strings.Contains(out, "default") {
		t.Fatalf("expected view to contain 'default', got: %s", out)
	}
}
