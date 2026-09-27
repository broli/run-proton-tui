package integrations

import (
	"strings"
	"testing"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/runner"
)

func TestDetermineRecommendedTier(t *testing.T) {
	// 1. Crash detected -> Borked
	crashedResult := &runner.SessionResult{
		CrashDetected: true,
		ExitCode:      139,
	}
	tier := DetermineRecommendedTier(config.NewDefaultConfig(), crashedResult, nil)
	if tier != "Borked" {
		t.Errorf("Expected Borked for crash, got %s", tier)
	}

	// 2. Tweaks applied (e.g. Gamescope + GE) -> Gold
	cleanResult := &runner.SessionResult{
		CrashDetected: false,
		ExitCode:      0,
		Duration:      30 * time.Minute,
	}
	cfgTweaked := config.NewDefaultConfig()
	cfgTweaked.UseGamescope = true
	cfgTweaked.ProtonPath = "/opt/GE-Proton9-25/proton"
	tier = DetermineRecommendedTier(cfgTweaked, cleanResult, nil)
	if tier != "Gold" {
		t.Errorf("Expected Gold for tweaked runner, got %s", tier)
	}

	// 3. No tweaks, clean exit -> Platinum
	cfgVanilla := config.NewDefaultConfig()
	cfgVanilla.UseGamescope = false
	cfgVanilla.UsePrimeRun = false
	cfgVanilla.UsePCores = false
	cfgVanilla.ProtonPath = "/opt/proton/proton"
	tier = DetermineRecommendedTier(cfgVanilla, cleanResult, nil)
	if tier != "Platinum" {
		t.Errorf("Expected Platinum for vanilla clean run, got %s", tier)
	}

	// 4. History with high crash rate -> Bronze
	badHistory := []runner.SessionRecord{
		{CrashDetected: true, ExitCode: 1},
		{CrashDetected: true, ExitCode: 139},
		{CrashDetected: false, ExitCode: 0},
	}
	tier = DetermineRecommendedTier(cfgTweaked, cleanResult, badHistory)
	if tier != "Bronze" {
		t.Errorf("Expected Bronze for high crash history, got %s", tier)
	}
}

func TestGenerateProtonDBReport(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.AppID = "4732690"
	cfg.UseGamescope = true
	cfg.GamescopeOutput = "HDMI-A-1"

	res := &runner.SessionResult{
		Duration:      15 * time.Minute,
		ExitCode:      0,
		CrashDetected: false,
	}

	report := GenerateProtonDBReport(ExportOptions{
		AppID:      "4732690",
		GameTitle:  "Arknights: Endfield",
		Config:     cfg,
		LastResult: res,
	})

	if !strings.Contains(report, "AppID: 4732690") {
		t.Errorf("Report missing AppID: %s", report)
	}
	if !strings.Contains(report, "HDMI-A-1") {
		t.Errorf("Report missing Gamescope info: %s", report)
	}
	if !strings.Contains(report, "Recommended Verdict: GOLD") {
		t.Errorf("Expected Recommended Verdict: GOLD, report had: %s", report)
	}
}
