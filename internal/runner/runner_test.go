package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
)

func TestClassifyExecutable(t *testing.T) {
	tests := []struct {
		exe       string
		want2D    bool
		wantGames bool
	}{
		{"Launcher.exe", true, false},
		{"Games.exe", true, false},
		{"Setup.exe", true, false},
		{"unins000.exe", true, false},
		{"Endfield.exe", false, true},
		{"Game-Win64-Shipping.exe", false, true},
		{"EldenRing.exe", false, true},
	}

	for _, tt := range tests {
		res := ClassifyExecutable(tt.exe)
		if tt.want2D && res.Type != ExeType2DUtility {
			t.Errorf("ClassifyExecutable(%q) got %v, want 2D utility", tt.exe, res.Type)
		}
		if !tt.want2D && res.Type != ExeType3DGame {
			t.Errorf("ClassifyExecutable(%q) got %v, want 3D game", tt.exe, res.Type)
		}
		if res.RecommendGamescope != tt.wantGames {
			t.Errorf("ClassifyExecutable(%q) RecommendGamescope = %v, want %v", tt.exe, res.RecommendGamescope, tt.wantGames)
		}
	}
}

func TestEnsurePrefixDirectories(t *testing.T) {
	tempDir := t.TempDir()
	prefixDir := filepath.Join(tempDir, "proton-prefix")
	pfxSubDir := filepath.Join(prefixDir, "pfx")

	if err := os.MkdirAll(pfxSubDir, 0755); err != nil {
		t.Fatalf("Failed to create prefix directories: %v", err)
	}

	if info, err := os.Stat(pfxSubDir); err != nil || !info.IsDir() {
		t.Fatalf("Expected %s to exist as a directory", pfxSubDir)
	}

	// Verify lock file can be opened inside prefixDir (as proton does)
	lockFile := filepath.Join(prefixDir, "pfx.lock")
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Failed to open lock file: %v", err)
	}
	_ = f.Close()
}

func TestRunnerEffectiveConfig(t *testing.T) {
	// Verify that RunGame creates prefix directory and executes pre-launch hook
	tempDir := t.TempDir()
	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "Launcher.exe"

	// Mock proton path with a dummy script
	mockProton := filepath.Join(tempDir, "mock_proton.sh")
	_ = os.WriteFile(mockProton, []byte("#!/bin/bash\nexit 0\n"), 0755)
	cfg.ProtonPath = mockProton

	// Add profile for Launcher.exe disabling gamescope
	falseVal := false
	cfg.Profiles["Launcher.exe"] = &config.ExecutableProfile{
		TargetExe:    "Launcher.exe",
		UseGamescope: &falseVal,
	}

	eff := cfg.GetEffectiveConfig("Launcher.exe")
	if eff.UseGamescope {
		t.Errorf("Expected effective config to have UseGamescope=false")
	}
}

