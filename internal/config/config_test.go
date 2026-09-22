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

func TestEffectiveConfigProfiles(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.TargetExe = "Game.exe"
	cfg.UseGamescope = true
	cfg.UsePrimeRun = true

	// Add profile for Setup.exe (2D utility / repack unpacker)
	falseVal := false
	cfg.Profiles["Setup.exe"] = &ExecutableProfile{
		TargetExe:    "Setup.exe",
		UseGamescope: &falseVal,
		UsePrimeRun:  &falseVal,
		ExtraArgs:    []string{"/SILENT"},
		EnvVars: map[string]string{
			"WINEDLLOVERRIDES": "mscoree=d",
		},
	}

	// For default Game.exe
	gameEff := cfg.GetEffectiveConfig("Game.exe")
	if !gameEff.UseGamescope || !gameEff.UsePrimeRun {
		t.Errorf("Game.exe should preserve base settings")
	}

	// For Setup.exe
	setupEff := cfg.GetEffectiveConfig("Setup.exe")
	if setupEff.UseGamescope {
		t.Errorf("Setup.exe should have UseGamescope=false")
	}
	if setupEff.UsePrimeRun {
		t.Errorf("Setup.exe should have UsePrimeRun=false")
	}
	if len(setupEff.ExtraArgs) != 1 || setupEff.ExtraArgs[0] != "/SILENT" {
		t.Errorf("Setup.exe should have extra args")
	}
	if setupEff.EnvVars["WINEDLLOVERRIDES"] != "mscoree=d" {
		t.Errorf("Setup.exe should have custom env var")
	}
}
