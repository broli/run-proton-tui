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

	// 1. Test MenuTarget view & mnemonic keys
	mvTarget := NewSubMenuView(MenuTarget, data)
	tView := mvTarget.View()
	if !strings.Contains(tView, "TARGET & RUNNER COMPATIBILITY") {
		t.Errorf("Expected 'TARGET & RUNNER COMPATIBILITY' title in view")
	}
	if !strings.Contains(tView, "[e / 1]") || !strings.Contains(tView, "[r / 2]") {
		t.Errorf("Expected mnemonic keybindings in view")
	}

	if act := mvTarget.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}); act != ActionOpenExePicker {
		t.Errorf("Expected ActionOpenExePicker on 'e', got %v", act)
	}
	if act := mvTarget.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}); act != ActionOpenExePicker {
		t.Errorf("Expected ActionOpenExePicker on '1', got %v", act)
	}
	if act := mvTarget.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}); act != ActionOpenProtonPicker {
		t.Errorf("Expected ActionOpenProtonPicker on 'r', got %v", act)
	}
	if act := mvTarget.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}); act != ActionOpenProtonPicker {
		t.Errorf("Expected ActionOpenProtonPicker on '2', got %v", act)
	}
	if act := mvTarget.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}); act != ActionOpenQuirks {
		t.Errorf("Expected ActionOpenQuirks on 'd', got %v", act)
	}
	if act := mvTarget.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}); act != ActionOpenProtonDB {
		t.Errorf("Expected ActionOpenProtonDB on 'a', got %v", act)
	}

	// 2. Test MenuDisplay view & actions
	mvDisplay := NewSubMenuView(MenuDisplay, data)
	dView := mvDisplay.View()
	if !strings.Contains(dView, "DISPLAY & GAMESCOPE SANDBOX") {
		t.Errorf("Expected 'DISPLAY & GAMESCOPE SANDBOX' title in view")
	}
	if act := mvDisplay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}); act != ActionToggleGamescope {
		t.Errorf("Expected ActionToggleGamescope on 'g', got %v", act)
	}
	if act := mvDisplay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}); act != ActionCycleDisplayOutput {
		t.Errorf("Expected ActionCycleDisplayOutput on 'm', got %v", act)
	}
	if act := mvDisplay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}); act != ActionDetectMonitor {
		t.Errorf("Expected ActionDetectMonitor on 'd', got %v", act)
	}
	if act := mvDisplay.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}); act != ActionOpenGamescopeSettings {
		t.Errorf("Expected ActionOpenGamescopeSettings on 's', got %v", act)
	}

	// 3. Test MenuHardware view & toggles
	mvHw := NewSubMenuView(MenuHardware, data)
	hwView := mvHw.View()
	if !strings.Contains(hwView, "HARDWARE & ENGINE PERFORMANCE") {
		t.Errorf("Expected 'HARDWARE & ENGINE PERFORMANCE' title in view")
	}
	if act := mvHw.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}); act != ActionTogglePrimeRun {
		t.Errorf("Expected ActionTogglePrimeRun on 'v', got %v", act)
	}
	if act := mvHw.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}); act != ActionTogglePCores {
		t.Errorf("Expected ActionTogglePCores on 'p', got %v", act)
	}
	if act := mvHw.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}); act != ActionToggleXalia {
		t.Errorf("Expected ActionToggleXalia on 'x', got %v", act)
	}

	// 4. Test MenuPrefix
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

	// 5. Test MenuDiagnostics
	mvDiag := NewSubMenuView(MenuDiagnostics, data)
	if act := mvDiag.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}}); act != ActionToggleLogging {
		t.Errorf("Expected ActionToggleLogging on 'L', got %v", act)
	}
	if act := mvDiag.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}); act != ActionOpenLogs {
		t.Errorf("Expected ActionOpenLogs on 'l', got %v", act)
	}
	if act := mvDiag.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}); act != ActionOpenDiagnostics {
		t.Errorf("Expected ActionOpenDiagnostics on 'k', got %v", act)
	}
	if act := mvDiag.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}); act != ActionCreateDesktopShortcut {
		t.Errorf("Expected ActionCreateDesktopShortcut on 's', got %v", act)
	}
	if act := mvDiag.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}); act != ActionToggleBackdrop {
		t.Errorf("Expected ActionToggleBackdrop on 'b', got %v", act)
	}
	if act := mvDiag.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}); act != ActionOpenHelp {
		t.Errorf("Expected ActionOpenHelp on '?', got %v", act)
	}

	// 6. Test Escape returns ActionClose
	if act := mvDisplay.Update(tea.KeyMsg{Type: tea.KeyEsc}); act != ActionClose {
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
	mv1 := NewSubMenuView(MenuPrefix, d1)
	v1 := mv1.View()
	if !strings.Contains(v1, "Active (Press to Wipe)") {
		t.Errorf("Expected 'Active (Press to Wipe)' status in view, got: %s", v1)
	}

	// Case 2: Prefix not initialized
	d2 := SubMenuData{
		Config:       cfg,
		PrefixActive: false,
		Width:        100,
		Height:       30,
	}
	mv2 := NewSubMenuView(MenuPrefix, d2)
	v2 := mv2.View()
	if !strings.Contains(v2, "Not Initialized (Clean)") {
		t.Errorf("Expected 'Not Initialized (Clean)' status in view, got: %s", v2)
	}

	// Case 3: Just cleaned
	d3 := SubMenuData{
		Config:        cfg,
		PrefixActive:  true,
		PrefixCleaned: true,
		Width:         100,
		Height:        30,
	}
	mv3 := NewSubMenuView(MenuPrefix, d3)
	v3 := mv3.View()
	if !strings.Contains(v3, "Reset Complete (Clean)") {
		t.Errorf("Expected 'Reset Complete (Clean)' status in view, got: %s", v3)
	}
}
