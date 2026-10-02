package diagnostics

import (
	"os"
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
)

func TestGenerateBugReport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rpt-bugreport-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "game.exe"
	cfg.AppID = "12345"

	opts := BugReportOptions{
		Version: "0.8.0",
		GameDir: tmpDir,
		Config:  cfg,
	}

	report, targetURL := GenerateBugReport(opts)

	if !strings.Contains(report, "rpt Version:      0.8.0") {
		t.Errorf("Expected rpt version in bug report")
	}
	if !strings.Contains(report, "Go Runtime:") {
		t.Errorf("Expected Go runtime in bug report")
	}
	if !strings.Contains(report, "Active Game Configuration") {
		t.Errorf("Expected configuration section in bug report")
	}
	if !strings.Contains(report, "app_id = \"12345\"") {
		t.Errorf("Expected AppID in bug report")
	}
	if !strings.Contains(report, "Health & Permissions Check") {
		t.Errorf("Expected health check section in bug report")
	}
	if !strings.Contains(report, "Privacy Confirmation") {
		t.Errorf("Expected privacy confirmation in bug report")
	}
	if !strings.Contains(targetURL, "https://github.com/broli/run-proton-tui/issues/new") {
		t.Errorf("Expected GitHub issue URL, got %s", targetURL)
	}

	// Verify no unsanitized home directory in report
	if home, err := os.UserHomeDir(); err == nil && home != "" && home != "/" {
		// Only check if home is an actual specific path like /home/username
		if strings.HasPrefix(tmpDir, home) && strings.Contains(report, home) {
			t.Errorf("Found unsanitized home path %q in bug report", home)
		}
	}
}
