package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestResetConfirmView(t *testing.T) {
	rcv := NewResetConfirmView("Test Game", "Game.exe", "GE-Proton11-7", 80, 24)

	// Test rendering
	out := rcv.View()
	if !strings.Contains(out, "REVERT TO SAFE DEFAULTS CONFIRMATION") {
		t.Errorf("View missing title")
	}
	if !strings.Contains(out, "Preserved (Game.exe)") {
		t.Errorf("View missing preserved target exe")
	}
	if !strings.Contains(out, "Preserved (GE-Proton11-7)") {
		t.Errorf("View missing preserved proton name")
	}

	// Test confirmation with 'y'
	confirmed, done := rcv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !confirmed || !done {
		t.Errorf("Expected confirmed=true, done=true on 'y'")
	}

	// Test cancellation with 'n'
	confirmed, done = rcv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if confirmed || !done {
		t.Errorf("Expected confirmed=false, done=true on 'n'")
	}

	// Test cancellation with Esc
	confirmed, done = rcv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if confirmed || !done {
		t.Errorf("Expected confirmed=false, done=true on Esc")
	}
}
