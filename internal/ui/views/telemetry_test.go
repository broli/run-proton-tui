package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTelemetryView_Rendering(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.EnableLocalTelemetry = false

	v := NewTelemetryView(cfg, "/tmp/game", 80, 24)
	out := v.View()

	if !strings.Contains(out, "LOCAL SESSION TELEMETRY & PRIVACY AUDIT") {
		t.Errorf("View missing header title")
	}
	if !strings.Contains(out, "DISABLED (Default / Zero Persistent Tracking)") {
		t.Errorf("Expected disabled status badge by default, got: %s", out)
	}
	if !strings.Contains(out, "WHAT IS NEVER RECORDED") {
		t.Errorf("View missing privacy guarantees")
	}

	// Toggle action
	action, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	if action != TelemetryActionToggle {
		t.Errorf("Expected TelemetryActionToggle on 't', got %v", action)
	}

	// Inspect toggle
	action, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	if action != TelemetryActionNone || !v.InspectMode {
		t.Errorf("Expected InspectMode to be true on 'v'")
	}

	// Wipe action
	action, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if action != TelemetryActionWipe {
		t.Errorf("Expected TelemetryActionWipe on 'w', got %v", action)
	}

	// Close action
	action, _ = v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if action != TelemetryActionClose {
		t.Errorf("Expected TelemetryActionClose on 'esc', got %v", action)
	}
}
