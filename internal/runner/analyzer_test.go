package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
)

func TestAnalyzeSessionLog_JustCauseHang(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "game_test.log")

	logContent := `[gamescope] [Info]  xdg_backend: Post-Initted Wayland backend
Proton: Upgrading prefix from None to GE-Proton11-6
ntsync: up and running.
[Gamescope WSI] Executable name: explorer.exe
Allocating 67108864 bytes at 0x19e20020 for CODA transfers
`
	if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("failed to write test log: %v", err)
	}

	cfg := &config.GameConfig{
		TargetExe:    "JustCause.exe",
		AppID:        "225540",
		UseGamescope: true,
	}

	insights := AnalyzeSessionLog(logFile, cfg)
	if len(insights) < 2 {
		t.Fatalf("expected at least 2 diagnostic insights, got %d", len(insights))
	}

	foundGamescope := false
	foundCODA := false
	foundAppID := false

	for _, ins := range insights {
		if strings.Contains(ins.Category, "Gamescope") {
			foundGamescope = true
		}
		if strings.Contains(ins.Category, "Bink Video") || strings.Contains(ins.Category, "CODA") {
			foundCODA = true
		}
		if strings.Contains(ins.Category, "AppID") {
			foundAppID = true
		}
	}

	if !foundGamescope {
		t.Errorf("expected Gamescope insight to be generated")
	}
	if !foundCODA {
		t.Errorf("expected CODA / Bink Video insight to be generated")
	}
	if !foundAppID {
		t.Errorf("expected Steam AppID mismatch insight to be generated")
	}
}

func TestFormatDiagnosticReport(t *testing.T) {
	cfg := &config.GameConfig{
		TargetExe:    "JustCause.exe",
		ProtonPath:   "/home/user/Proton-GE",
		AppID:        "6880",
		UseGamescope: false,
	}

	result := &SessionResult{
		ExitCode:      130,
		Duration:      35 * time.Second,
		LogFile:       "/tmp/test.log",
		AbortedByUser: true,
	}

	insights := []DiagnosticInsight{
		{
			Category:       "Splash Screen & Bink Video / CODA Stall",
			Observation:    "Engine allocated 64MB CODA buffer for intro video playback.",
			Recommendation: "Run JCSetup.exe first to lock resolution.",
		},
	}

	report := FormatDiagnosticReport(result, cfg, insights)
	if !strings.Contains(report, "We believe the game crashed, did not work") {
		t.Errorf("expected report to contain explanation quote, got:\n%s", report)
	}
	if !strings.Contains(report, "Splash Screen") && !strings.Contains(report, "Bink Video") {
		t.Errorf("expected report to contain insight category")
	}
}

func TestGenerateAgentPrompt(t *testing.T) {
	cfg := &config.GameConfig{
		TargetExe:    "JustCause.exe",
		ProtonPath:   "/home/user/Proton-GE",
		AppID:        "225540",
		UseGamescope: true,
	}

	result := &SessionResult{
		ExitCode:      130,
		Duration:      40 * time.Second,
		LogFile:       "/tmp/test.log",
		AbortedByUser: true,
	}

	logTail := []string{
		"[Gamescope WSI] Executable name: explorer.exe",
		"Allocating 67108864 bytes at 0x19e20020 for CODA transfers",
	}

	prompt := GenerateAgentPrompt(result, cfg, nil, logTail)
	if !strings.Contains(prompt, "JustCause.exe") {
		t.Errorf("expected prompt to contain TargetExe")
	}
	if !strings.Contains(prompt, "Allocating 67108864 bytes") {
		t.Errorf("expected prompt to contain logTail line")
	}
	if !strings.Contains(prompt, "Question for Agent") {
		t.Errorf("expected prompt to contain Question for Agent")
	}
}

func TestAnalyzeSessionLog_FailedToOpen(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "failed_open.log")

	logContent := `ntsync: up and running.
wine: failed to open "./JCSetup.exe"
`
	if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("failed to write test log: %v", err)
	}

	cfg := &config.GameConfig{
		TargetExe: "JCSetup.exe",
	}

	insights := AnalyzeSessionLog(logFile, cfg)
	if len(insights) == 0 {
		t.Fatalf("expected insight for failed to open executable")
	}

	found := false
	for _, ins := range insights {
		if strings.Contains(ins.Category, "Executable Not Found") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Executable Not Found insight")
	}
}
