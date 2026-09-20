package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/runner"
)

func TestRenderDashboardWidths(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "Setup.exe"

	data := DashboardData{
		GameTitle:      "Just Cause [DODI Repack]",
		GameDir:        "/home/carlos/Games/Just Cause [DODI Repack]",
		Config:         cfg,
		Emulator:       &integrations.EmulatorInfo{Type: "Goldberg", AppID: "12345", DLCStatus: "All Unlocked", Detected: true},
		Classification: runner.ClassifyExecutable("Setup.exe"),
		HasNTSync:      true,
		HasPrimeRun:    true,
		HasGamescope:   true,
		ProtonName:     "Proton-GE Latest",
		PrefixSize:     "1.4 GB",
		Width:          120,
		Height:         35,
	}

	out := RenderDashboard(data)
	if out == "" {
		t.Fatal("RenderDashboard returned empty string")
	}
	t.Log("\n" + out)

	// Verify header title appears
	if !strings.Contains(out, "Just Cause [DODI Repack]") {
		t.Errorf("Expected GameTitle in output")
	}

	// Verify hotkeys are present
	if !strings.Contains(out, "Launch Game") || !strings.Contains(out, "Quit") {
		t.Errorf("Expected hotkeys in output")
	}

	// Verify 2D Utility badge
	if !strings.Contains(out, "2D Utility") {
		t.Errorf("Expected 2D Utility classification in output")
	}
}
