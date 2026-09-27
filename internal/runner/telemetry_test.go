package runner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
)

func TestRecordSession_AssumeOffByDefault(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rpt-telemetry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.NewDefaultConfig()
	// By default, EnableLocalTelemetry must be false
	if cfg.EnableLocalTelemetry {
		t.Fatalf("Expected EnableLocalTelemetry to default to false, got true")
	}

	res := &SessionResult{
		Duration:      15 * time.Minute,
		ExitCode:      0,
		CrashDetected: false,
	}

	err = RecordSession(tempDir, cfg, res, "NVIDIA RTX 4060 / Intel i7")
	if err != nil {
		t.Fatalf("RecordSession error: %v", err)
	}

	historyFile := GetHistoryFilePath(tempDir)
	if _, err := os.Stat(historyFile); !os.IsNotExist(err) {
		t.Errorf("Expected history file to NOT exist when telemetry is disabled, but it was created: %s", historyFile)
	}
}

func TestRecordSession_OptInEnabled(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rpt-telemetry-optin-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.NewDefaultConfig()
	cfg.EnableLocalTelemetry = true
	cfg.AppID = "4732690"
	cfg.ProtonPath = "/opt/proton-ge/proton"
	cfg.UseGamescope = true

	res := &SessionResult{
		Duration:      45 * time.Minute,
		ExitCode:      0,
		CrashDetected: false,
	}

	err = RecordSession(tempDir, cfg, res, "NVIDIA RTX 4060")
	if err != nil {
		t.Fatalf("RecordSession failed: %v", err)
	}

	history, err := ReadSessionHistory(tempDir)
	if err != nil {
		t.Fatalf("ReadSessionHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 history record, got %d", len(history))
	}

	rec := history[0]
	if rec.AppID != "4732690" {
		t.Errorf("Expected AppID 4732690, got %s", rec.AppID)
	}
	if rec.RunnerName != "proton" {
		t.Errorf("Expected RunnerName 'proton', got %s", rec.RunnerName)
	}
	if rec.DurationSeconds <= 0 {
		t.Errorf("Expected positive duration, got %f", rec.DurationSeconds)
	}

	// Test ClearSessionHistory
	if err := ClearSessionHistory(tempDir); err != nil {
		t.Fatalf("ClearSessionHistory failed: %v", err)
	}

	historyAfterClear, err := ReadSessionHistory(tempDir)
	if err != nil {
		t.Fatalf("ReadSessionHistory after clear failed: %v", err)
	}
	if len(historyAfterClear) != 0 {
		t.Errorf("Expected 0 history records after wipe, got %d", len(historyAfterClear))
	}
}

func TestSanitizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("No user home dir found")
	}

	testPath := filepath.Join(home, "Games", "MyGame", "game.exe")
	sanitized := SanitizePath(testPath)
	expectedPrefix := "~/"
	if filepath.Separator == '\\' {
		expectedPrefix = "~\\"
	}

	if sanitized == testPath {
		t.Errorf("SanitizePath failed to strip home dir: %s", sanitized)
	}
	if sanitized[:2] != expectedPrefix[:2] {
		t.Errorf("Expected prefix %s, got %s", expectedPrefix, sanitized)
	}
}
