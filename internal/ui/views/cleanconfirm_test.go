package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCleanConfirmView_Rendering(t *testing.T) {
	v := NewCleanConfirmView("Styx: Master of Shadows", "./proton-prefix/", "", 80, 25)

	out := v.View()
	if !strings.Contains(out, "RESET WINE PREFIX CONFIRMATION") {
		t.Errorf("Expected title in CleanConfirmView, got: %s", out)
	}
	if !strings.Contains(out, "Styx: Master of Shadows") {
		t.Errorf("Expected game title in CleanConfirmView")
	}
	if !strings.Contains(out, "SafeClean Protection") {
		t.Errorf("Expected SafeClean protection notice in CleanConfirmView")
	}
	if !strings.Contains(out, "Yes, Reset Prefix") || !strings.Contains(out, "Cancel & Keep Prefix") {
		t.Errorf("Expected action prompts in dock")
	}
}

func TestCleanConfirmView_Keys(t *testing.T) {
	v := NewCleanConfirmView("Styx", "./proton-prefix/", "", 80, 25)

	// 'y' confirms
	confirmed, done := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !confirmed || !done {
		t.Errorf("Expected confirmed=true, done=true on 'y'")
	}

	// 'Enter' confirms
	confirmed, done = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !confirmed || !done {
		t.Errorf("Expected confirmed=true, done=true on Enter")
	}

	// 'n' cancels
	confirmed, done = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if confirmed || !done {
		t.Errorf("Expected confirmed=false, done=true on 'n'")
	}

	// 'Esc' cancels
	confirmed, done = v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if confirmed || !done {
		t.Errorf("Expected confirmed=false, done=true on Esc")
	}

	// Unrelated key ignored
	confirmed, done = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if confirmed || done {
		t.Errorf("Expected confirmed=false, done=false on 'x'")
	}
}
