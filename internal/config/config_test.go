package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadSaveMigration(t *testing.T) {
	tmpDir := t.TempDir()

	// Write mock legacy config
	legacyContent := `
target_exe=Game.exe
proton_path=/path/to/proton
use_gamescope=1
use_pcores=1
use_xalia=0
app_id=4567
`
	legacyFile := filepath.Join(tmpDir, LegacyConfigFileName)
	if err := os.WriteFile(legacyFile, []byte(legacyContent), 0644); err != nil {
		t.Fatalf("Failed to write legacy file: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.TargetExe != "Game.exe" {
		t.Errorf("Expected TargetExe=Game.exe, got %s", cfg.TargetExe)
	}
	if !cfg.UseGamescope {
		t.Errorf("Expected UseGamescope=true, got %v", cfg.UseGamescope)
	}
	if cfg.AppID != "4567" {
		t.Errorf("Expected AppID=4567, got %s", cfg.AppID)
	}

	// Verify migrated TOML file was created
	tomlPath := filepath.Join(tmpDir, ConfigFileName)
	if _, err := os.Stat(tomlPath); err != nil {
		t.Errorf("Expected %s to be created during migration", ConfigFileName)
	}
}
