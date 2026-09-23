package diagnostics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPreflightCheck(t *testing.T) {
	tmpDir := t.TempDir()
	exePath := filepath.Join(tmpDir, "Game.exe")
	// Create executable without +x bit
	_ = os.WriteFile(exePath, []byte("fake-binary"), 0644)

	prefixDir := filepath.Join(tmpDir, "proton-prefix")
	report := RunPreflightCheck(tmpDir, "Game.exe", prefixDir)

	if !report.TargetExeExists {
		t.Errorf("Expected TargetExeExists=true")
	}

	if !report.TargetExeFixed {
		t.Errorf("Expected TargetExeFixed=true (auto-chmod +x)")
	}

	// Verify file is now executable
	info, err := os.Stat(exePath)
	if err != nil {
		t.Fatalf("Failed to stat exe: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("Expected executable bit to be set")
	}
}

func TestGenerateSpecDump(t *testing.T) {
	data, err := GenerateSpecDump("2.1.0")
	if err != nil {
		t.Fatalf("GenerateSpecDump failed: %v", err)
	}

	str := string(data)
	if len(str) == 0 {
		t.Fatalf("Expected non-empty output")
	}
	if !strings.Contains(str, "rpt-spec-v1") {
		t.Errorf("Expected rpt-spec-v1 in output")
	}
	if !strings.Contains(str, "connected_display_outputs") {
		t.Errorf("Expected connected_display_outputs in output")
	}
	if !strings.Contains(str, "lifecycle_hooks") {
		t.Errorf("Expected lifecycle_hooks in output")
	}
}
