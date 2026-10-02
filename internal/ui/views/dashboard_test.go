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

func TestRenderDashboardOpaque(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "Setup.exe"
	cfg.OpaqueBackdrop = true

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
		OpaqueBackdrop: true,
	}

	out := RenderDashboard(data)
	if out == "" {
		t.Fatal("RenderDashboard returned empty string with opaque backdrop")
	}

	if !strings.Contains(out, "Backdrop: Solid") {
		t.Errorf("Expected 'Backdrop: Solid' toggle text in opaque mode")
	}
}

func TestRenderDashboardProtonDBTitle(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "StyxGame.exe"

	data := DashboardData{
		GameTitle: "Styx: Master of Shadows",
		GameDir:   "/home/carlos/Games/Styx",
		Config:    cfg,
		ProtonDB: &integrations.ProtonDBReport{
			AppID:      "242640",
			Title:      "Styx: Master of Shadows",
			Tier:       "platinum",
			Confidence: "strong",
			Total:      142,
		},
		Classification: runner.ClassifyExecutable("StyxGame.exe"),
		Width:          120,
		Height:         35,
	}

	out := RenderDashboard(data)
	if !strings.Contains(out, "⭐ PLATINUM") {
		t.Errorf("Expected tier badge in output")
	}
	if !strings.Contains(out, "Styx: Master of Shadows") {
		t.Errorf("Expected matched game title in output")
	}
	if !strings.Contains(out, "142 reports") {
		t.Errorf("Expected report count in output")
	}
}

func TestRenderDashboardGPUOffloadStatus(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "game.exe"
	cfg.UsePrimeRun = false

	data := DashboardData{
		Config:      cfg,
		HasPrimeRun: true,
		Width:       120,
		Height:      35,
	}

	// 1. Unset state
	out := RenderDashboard(data)
	if !strings.Contains(out, "Unset (System Default)") {
		t.Errorf("Expected 'Unset (System Default)' for unforced GPU offload, got output: %s", out)
	}
	if strings.Contains(out, "Host iGPU") {
		t.Errorf("Did not expect 'Host iGPU' when GPU is not forced, got output: %s", out)
	}

	// 2. Active prime-run state
	cfg.UsePrimeRun = true
	outActive := RenderDashboard(data)
	if !strings.Contains(outActive, "prime-run (Dedicated GPU)") {
		t.Errorf("Expected 'prime-run (Dedicated GPU)' when prime-run is active, got output: %s", outActive)
	}
}
