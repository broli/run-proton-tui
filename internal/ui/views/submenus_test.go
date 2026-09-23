package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSubMenuView(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "Endfield.exe"
	cfg.PresetName = "Arknights: Endfield"

	data := SubMenuData{
		Config:         cfg,
		ProtonName:     "GE-Proton11-6",
		HasPrimeRun:    true,
		HasGamescope:   true,
		OpaqueBackdrop: false,
		Width:          100,
		Height:         30,
	}

	// 1. Test MenuProton view & mnemonic keys
	mvProton := NewSubMenuView(MenuProton, data)
	pView := mvProton.View()
	if !strings.Contains(pView, "PROTON & GAME TARGETS") {
		t.Errorf("Expected 'PROTON & GAME TARGETS' title in view")
	}
	if !strings.Contains(pView, "[r / 1]") || !strings.Contains(pView, "[d / 3]") {
		t.Errorf("Expected mnemonic keybindings in view")
	}

	if act := mvProton.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}); act != ActionOpenProtonPicker {
		t.Errorf("Expected ActionOpenProtonPicker on 'r', got %v", act)
	}
	if act := mvProton.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}); act != ActionOpenProtonPicker {
		t.Errorf("Expected ActionOpenProtonPicker on '1', got %v", act)
	}
	if act := mvProton.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}); act != ActionOpenQuirks {
		t.Errorf("Expected ActionOpenQuirks on 'd', got %v", act)
	}
	if act := mvProton.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}); act != ActionOpenProtonDB {
		t.Errorf("Expected ActionOpenProtonDB on 'a', got %v", act)
	}

	// 2. Test MenuPerformance view & toggles
	mvPerf := NewSubMenuView(MenuPerformance, data)
	perfView := mvPerf.View()
	if !strings.Contains(perfView, "PERFORMANCE & SANDBOX") {
		t.Errorf("Expected 'PERFORMANCE & SANDBOX' title in view")
	}

	if act := mvPerf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}); act != ActionToggleGamescope {
		t.Errorf("Expected ActionToggleGamescope on 'g', got %v", act)
	}
	if act := mvPerf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}); act != ActionToggleGamescope {
		t.Errorf("Expected ActionToggleGamescope on '1', got %v", act)
	}
	if act := mvPerf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}); act != ActionTogglePCores {
		t.Errorf("Expected ActionTogglePCores on 'p', got %v", act)
	}
	if act := mvPerf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}); act != ActionTogglePrimeRun {
		t.Errorf("Expected ActionTogglePrimeRun on 'v', got %v", act)
	}
	if act := mvPerf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}); act != ActionToggleXalia {
		t.Errorf("Expected ActionToggleXalia on 'x', got %v", act)
	}

	// 3. Test MenuPrefix
	mvPrefix := NewSubMenuView(MenuPrefix, data)
	if act := mvPrefix.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}); act != ActionOpenOverrides {
		t.Errorf("Expected ActionOpenOverrides on 'o', got %v", act)
	}
	if act := mvPrefix.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}); act != ActionCleanPrefix {
		t.Errorf("Expected ActionCleanPrefix on 'c', got %v", act)
	}
	if act := mvPrefix.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}); act != ActionOpenDiagnostics {
		t.Errorf("Expected ActionOpenDiagnostics on 'h', got %v", act)
	}

	// 4. Test MenuLogs
	mvLogs := NewSubMenuView(MenuLogs, data)
	if act := mvLogs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}); act != ActionToggleLogging {
		t.Errorf("Expected ActionToggleLogging on 'L', got %v", act)
	}
	if act := mvLogs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}); act != ActionOpenLogs {
		t.Errorf("Expected ActionOpenLogs on 'l', got %v", act)
	}

	// 5. Test MenuSettings
	mvSettings := NewSubMenuView(MenuSettings, data)
	if act := mvSettings.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}); act != ActionToggleBackdrop {
		t.Errorf("Expected ActionToggleBackdrop on 'b', got %v", act)
	}
	if act := mvSettings.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}); act != ActionOpenHelp {
		t.Errorf("Expected ActionOpenHelp on '?', got %v", act)
	}
	if act := mvSettings.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); act != ActionQuit {
		t.Errorf("Expected ActionQuit on 'q', got %v", act)
	}

	// 6. Test Escape returns ActionClose
	if act := mvPerf.Update(tea.KeyMsg{Type: tea.KeyEsc}); act != ActionClose {
		t.Errorf("Expected ActionClose on Esc, got %v", act)
	}
}

func TestMenuPrefixDynamicCleanStatus(t *testing.T) {
	cfg := config.NewDefaultConfig()

	// Case 1: Prefix is active
	d1 := SubMenuData{
		Config:       cfg,
		PrefixActive: true,
		Width:        100,
		Height:       30,
	}
	v1 := NewSubMenuView(MenuPrefix, d1).View()
	if !strings.Contains(v1, "Active (Press to Wipe)") {
		t.Errorf("Expected active prefix badge in view, got:\n%s", v1)
	}

	// Case 2: Prefix is not initialized
	d2 := SubMenuData{
		Config:       cfg,
		PrefixActive: false,
		Width:        100,
		Height:       30,
	}
	v2 := NewSubMenuView(MenuPrefix, d2).View()
	if !strings.Contains(v2, "Not Initialized (Clean)") {
		t.Errorf("Expected uninitialized badge in view, got:\n%s", v2)
	}

	// Case 3: Prefix was cleaned
	d3 := SubMenuData{
		Config:        cfg,
		PrefixActive:  false,
		PrefixCleaned: true,
		StatusMessage: "Prefix cleaned and reset successfully.",
		Width:         100,
		Height:        30,
	}
	v3 := NewSubMenuView(MenuPrefix, d3).View()
	if !strings.Contains(v3, "Reset Complete (Clean)") {
		t.Errorf("Expected reset complete badge in view, got:\n%s", v3)
	}
	if !strings.Contains(v3, "Prefix cleaned and reset successfully.") {
		t.Errorf("Expected status message banner in view, got:\n%s", v3)
	}
}
