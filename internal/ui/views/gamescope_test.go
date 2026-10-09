package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestGamescopeView(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.GamescopeWidth = 1920
	cfg.GamescopeHeight = 1080
	cfg.GamescopeRefresh = 0
	cfg.GamescopeOutput = "HDMI-A-1"

	gv := NewGamescopeView(cfg, 100, 30)

	// 1. Initial view checks
	v := gv.View()
	if !strings.Contains(v, "GAMESCOPE SANDBOX & UPSCALING CONFIGURATION") {
		t.Errorf("Expected title in Gamescope view")
	}
	if !strings.Contains(v, "Native / Untouched (0)") {
		t.Errorf("Expected default refresh to show Native/Untouched, got: %s", v)
	}

	// 2. Test navigation
	gv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if gv.Cursor != 1 {
		t.Errorf("Expected cursor at 1 after down, got %d", gv.Cursor)
	}
	gv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if gv.Cursor != 0 {
		t.Errorf("Expected cursor at 0 after up, got %d", gv.Cursor)
	}

	// 3. Test cycling resolution
	gv.Cursor = int(FieldResolution)
	gv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cfg.GamescopeWidth == 1920 && cfg.GamescopeHeight == 1080 {
		t.Errorf("Expected resolution to cycle, but remained 1920x1080")
	}

	// 4. Test cycling refresh rate
	gv.Cursor = int(FieldRefresh)
	gv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cfg.GamescopeRefresh == 0 {
		t.Errorf("Expected refresh to cycle from 0 to positive Hz")
	}

	// 5. Test toggles
	gv.Cursor = int(FieldAdaptiveSync)
	origSync := cfg.GamescopeAdaptiveSync
	gv.Update(tea.KeyMsg{Type: tea.KeySpace})
	if cfg.GamescopeAdaptiveSync == origSync {
		t.Errorf("Expected AdaptiveSync to toggle")
	}

	gv.Cursor = int(FieldMangoApp)
	origMango := cfg.GamescopeMangoApp
	gv.Update(tea.KeyMsg{Type: tea.KeySpace})
	if cfg.GamescopeMangoApp == origMango {
		t.Errorf("Expected MangoApp to toggle")
	}

	// 6. Test sharpness +/-
	gv.Cursor = int(FieldSharpness)
	cfg.GamescopeSharpness = 5
	gv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	if cfg.GamescopeSharpness != 6 {
		t.Errorf("Expected sharpness 6 after +, got %d", cfg.GamescopeSharpness)
	}
	gv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	if cfg.GamescopeSharpness != 5 {
		t.Errorf("Expected sharpness 5 after -, got %d", cfg.GamescopeSharpness)
	}

	// 7. Test Detect Size shortcut
	_, detected := gv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !detected {
		t.Errorf("Expected detected=true on 'd'")
	}
	if cfg.GamescopeRefresh != 0 {
		t.Errorf("Expected Detect to reset GamescopeRefresh to 0 (Native), got %d", cfg.GamescopeRefresh)
	}

	// 8. Test 'r' shortcut to reset to safe defaults
	cfg.GamescopeWidth = 3840
	cfg.GamescopeHeight = 2160
	cfg.GamescopeHDR = true
	gv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cfg.GamescopeWidth != 0 || cfg.GamescopeHeight != 0 || cfg.GamescopeHDR {
		t.Errorf("Expected 'r' to reset Gamescope to Auto geometry (0x0) SDR, got %dx%d HDR=%v", cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GamescopeHDR)
	}
	resetView := gv.View()
	if !strings.Contains(resetView, "Auto / Native (Blank - Host Negotiated)") {
		t.Errorf("Expected 'Auto / Native (Blank - Host Negotiated)' in view after reset, got: %s", resetView)
	}
	if strings.Contains(resetView, "External Preferred") {
		t.Errorf("Did not expect 'External Preferred' in view after reset, got: %s", resetView)
	}

	// Test 'u' shortcut to clear geometry to blank
	cfg.GamescopeWidth = 1920
	cfg.GamescopeHeight = 1080
	cfg.GeometryAutoDetected = true
	gv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if cfg.GamescopeWidth != 0 || cfg.GamescopeHeight != 0 || cfg.GeometryAutoDetected {
		t.Errorf("Expected 'u' to clear geometry to 0x0 and unset GeometryAutoDetected, got %dx%d auto=%v", cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GeometryAutoDetected)
	}

	// 9. Test Esc returns done=true
	done, _ := gv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !done {
		t.Errorf("Expected done=true on Esc")
	}
}
