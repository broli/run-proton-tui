package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOverridesViewPresetsAndCustom(t *testing.T) {
	tempDir := t.TempDir()
	activeReg := map[string]string{
		"dwmapi": "native,builtin",
		"custom1": "native",
	}
	savedCfg := map[string]string{
		"custom2": "builtin,native",
	}

	ov := NewOverridesView(tempDir, "", activeReg, savedCfg)

	// Verify presets and custom items exist
	if len(ov.Items) < 7 { // 5 presets + 2 custom
		t.Fatalf("Expected at least 7 items, got %d", len(ov.Items))
	}

	active := ov.GetActiveOverrides()
	if active["dwmapi"] != "native,builtin" {
		t.Errorf("Expected dwmapi to be active with native,builtin, got %s", active["dwmapi"])
	}
	if active["custom1"] != "native" {
		t.Errorf("Expected custom1 to be active with native, got %s", active["custom1"])
	}
	if active["custom2"] != "builtin,native" {
		t.Errorf("Expected custom2 to be active with builtin,native, got %s", active["custom2"])
	}

	// Test adding a new custom DLL
	// Trigger [a]
	ov, _ = ov.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !ov.InputMode {
		t.Fatalf("Expected InputMode to be true after pressing 'a'")
	}

	// Type "xinput1_3.dll"
	ov.TextInput.SetValue("xinput1_3.dll")
	ov, _ = ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if ov.InputMode {
		t.Fatalf("Expected InputMode to be false after pressing Enter")
	}

	activeAfterAdd := ov.GetActiveOverrides()
	if _, ok := activeAfterAdd["xinput1_3"]; !ok {
		t.Errorf("Expected xinput1_3 to be active after adding")
	}

	// Verify rendered view includes custom override
	viewStr := ov.View()
	if !strings.Contains(viewStr, "xinput1_3") {
		t.Errorf("Expected view to display xinput1_3, got:\n%s", viewStr)
	}
	if !strings.Contains(viewStr, "CUSTOM") {
		t.Errorf("Expected view to display CUSTOM tag, got:\n%s", viewStr)
	}

	// Test mode cycling on the newly added item (cursor is on it)
	prevMode := ov.Modes["xinput1_3"]
	ov, _ = ov.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	newMode := ov.Modes["xinput1_3"]
	if prevMode == newMode {
		t.Errorf("Expected mode to cycle after pressing 'm', but stayed %s", newMode)
	}

	// Test delete custom override with 'd'
	itemsBeforeDelete := len(ov.Items)
	ov, _ = ov.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if len(ov.Items) != itemsBeforeDelete-1 {
		t.Errorf("Expected items count to decrease by 1 after deletion, got %d (was %d)", len(ov.Items), itemsBeforeDelete)
	}
}
