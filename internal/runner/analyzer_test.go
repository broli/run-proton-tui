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

	report := FormatDiagnosticReport(result, cfg, insights, "/tmp/ask-ai-help.txt")
	if strings.Contains(report, "We believe") {
		t.Errorf("expected report not to contain speculative 'We believe', got:\n%s", report)
	}
	if strings.Contains(report, "Automated Root-Cause Insights:") {
		t.Errorf("expected report not to contain automated RCA guessing")
	}
	if !strings.Contains(report, "SESSION TERMINATED BY USER (Ctrl+C)") {
		t.Errorf("expected report to contain session termination title")
	}
	if !strings.Contains(report, "Proposed Solution:") {
		t.Errorf("expected report to contain 'Proposed Solution:' section, got:\n%s", report)
	}
	if !strings.Contains(report, "Troubleshooting Suggestions:") {
		t.Errorf("expected report to contain 'Troubleshooting Suggestions:' section")
	}
	if !strings.Contains(report, "Search Google:") {
		t.Errorf("expected report to contain Google search suggestion")
	}
	if !strings.Contains(report, "Search Reddit:") {
		t.Errorf("expected report to contain Reddit search suggestion")
	}
	if !strings.Contains(report, "Clean Zero Baseline:") {
		t.Errorf("expected report to contain Clean Zero Baseline suggestion")
	}
	if !strings.Contains(report, "Ask AI to Analyze Logs:") {
		t.Errorf("expected report to contain AI helper section")
	}
	if !strings.Contains(report, "Copy the contents of the file above.") {
		t.Errorf("expected report to instruct copying the file above")
	}
	if strings.Contains(report, "block below") {
		t.Errorf("report should not reference block below")
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

func TestExplainExitCode(t *testing.T) {
	tests := []struct {
		code          int
		expectedTitle string
	}{
		{0, "Clean / Normal Exit (Code 0)"},
		{1, "General Application Failure (Exit Code 1)"},
		{53, "Executable File Not Found (Exit Code 53)"},
		{126, "Permission Denied (Exit Code 126)"},
		{127, "Binary / Dynamic Library Not Found (Exit Code 127)"},
		{130, "Session Terminated by User (SIGINT / Exit Code 130)"},
		{134, "Process Aborted (SIGABRT / Exit Code 134)"},
		{137, "Terminated by Out-Of-Memory Killer (SIGKILL / Exit Code 137)"},
		{139, "Segmentation Fault (SIGSEGV / Exit Code 139)"},
		{222, "Application Crash Handler Trapped Fatal Exception (Exit Code 222)"},
	}

	for _, tt := range tests {
		exp := ExplainExitCode(tt.code)
		if exp.Title != tt.expectedTitle {
			t.Errorf("ExplainExitCode(%d).Title = %q, expected %q", tt.code, exp.Title, tt.expectedTitle)
		}
		if exp.Description == "" {
			t.Errorf("ExplainExitCode(%d).Description should not be empty", tt.code)
		}
	}
}

func TestAnalyzeSessionLog_ExitCode222(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	_ = os.WriteFile(logFile, []byte("some log line\n"), 0644)

	cfg := &config.GameConfig{
		TargetExe: "Endfield.exe",
	}

	insights := AnalyzeSessionLog(logFile, cfg, 222)
	found := false
	for _, ins := range insights {
		if strings.Contains(ins.Category, "Crash Handler") || strings.Contains(ins.Category, "222") {
			found = true
			if !strings.Contains(ins.Recommendation, "analyze the log") {
				t.Errorf("expected recommendation to mention analyzing the log")
			}
		}
	}
	if !found {
		t.Errorf("expected Crash Handler insight when exit code 222 is provided")
	}
}

func TestAnalyzeSessionLog_GamescopeWSINoFalsePositive(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "gamescope_wsi_test.log")

	logContent := `[gamescope] [Info]  xdg_backend: Post-Initted Wayland backend
[Gamescope WSI] Application info:
  pApplicationName: StyxGame.exe
  pEngineName: DXVK
[Gamescope WSI] Made gamescope surface for xid: 0x1200086
[Gamescope WSI] Created swapchain for xid: 0x1200086 swapchain: 0x7fd93c0b1170 - imageCount: 4
[Gamescope WSI] Swapchain received new refresh cycle: 13.33ms
`
	if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("failed to write test log: %v", err)
	}

	cfg := &config.GameConfig{
		TargetExe:    "StyxGame.exe",
		AppID:        "242640",
		UseGamescope: true,
	}

	insights := AnalyzeSessionLog(logFile, cfg)
	for _, ins := range insights {
		if ins.Category == "Display Initialization" {
			t.Errorf("expected no false-positive 'Display Initialization' insight when Gamescope WSI swapchain was created, got: %+v", ins)
		}
	}
}

func TestGenerateAIHelperPackage_PrefixFlushInstruction(t *testing.T) {
	result := &SessionResult{
		ExitCode:      1,
		Duration:      10 * time.Second,
		CrashDetected: true,
		LogFile:       "/tmp/test.log",
	}
	cfg := &config.GameConfig{
		TargetExe:  "Game.exe",
		ProtonPath: "/path/to/proton",
		AppID:      "12345",
	}

	pkg := GenerateAIHelperPackage(result, cfg, nil, []string{"log line 1"}, "/tmp/test.log")
	if !strings.Contains(pkg, "DO NOT add Wine prefix flushing or wineserver kill/lock cleanup to hooks") {
		t.Errorf("Expected instructions to caution against adding wine prefix flushing to hooks, got:\n%s", pkg)
	}
}


