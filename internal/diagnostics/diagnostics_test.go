package diagnostics

import (
	"os"
	"path/filepath"
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
