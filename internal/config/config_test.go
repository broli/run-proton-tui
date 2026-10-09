package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigFileProfilesAndLoadSave(t *testing.T) {
	tmpDir := t.TempDir()

	fileCfg := NewConfigFileWithDefault()
	defProfile := fileCfg.GetActiveProfile()
	defProfile.TargetExe = "Game.exe"
	defProfile.UseGamescope = true
	defProfile.AppID = "4567"

	// Add a secondary launcher profile
	launcherCfg := NewDefaultConfig()
	launcherCfg.TargetExe = "Launcher.exe"
	launcherCfg.UseGamescope = false
	if err := fileCfg.AddProfile("launcher", launcherCfg); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	// Save to rpt.toml
	if err := SaveConfigFile(tmpDir, fileCfg); err != nil {
		t.Fatalf("SaveConfigFile failed: %v", err)
	}

	// Verify rpt.toml exists
	tomlPath := filepath.Join(tmpDir, ConfigFileName)
	if _, err := os.Stat(tomlPath); err != nil {
		t.Fatalf("Expected %s to exist", ConfigFileName)
	}

	// Load back
	loadedFile, err := LoadConfigFile(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	if loadedFile.ActiveProfile != "default" {
		t.Errorf("Expected ActiveProfile=default, got %s", loadedFile.ActiveProfile)
	}
	if len(loadedFile.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(loadedFile.Profiles))
	}
	loadedDef := loadedFile.GetActiveProfile()
	if loadedDef.TargetExe != "Game.exe" {
		t.Errorf("Expected TargetExe=Game.exe, got %s", loadedDef.TargetExe)
	}
	if !loadedDef.UseGamescope {
		t.Errorf("Expected UseGamescope=true, got %v", loadedDef.UseGamescope)
	}

	// Switch profile
	if err := loadedFile.SetActiveProfile("launcher"); err != nil {
		t.Fatalf("SetActiveProfile failed: %v", err)
	}
	active := loadedFile.GetActiveProfile()
	if active.TargetExe != "Launcher.exe" {
		t.Errorf("Expected active TargetExe=Launcher.exe, got %s", active.TargetExe)
	}

	// Clone profile
	if err := loadedFile.CloneProfile("launcher", "portable"); err != nil {
		t.Fatalf("CloneProfile failed: %v", err)
	}
	if len(loadedFile.Profiles) != 3 {
		t.Errorf("Expected 3 profiles after clone, got %d", len(loadedFile.Profiles))
	}

	// Delete profile
	if err := loadedFile.DeleteProfile("portable"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}
	if len(loadedFile.Profiles) != 2 {
		t.Errorf("Expected 2 profiles after delete, got %d", len(loadedFile.Profiles))
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

func TestSubfolderProfileMatching(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.TargetExe = "games/Arknights Endfield/Endfield.exe"
	cfg.UseGamescope = true
	cfg.UsePrimeRun = true

	falseVal := false
	// Key is simple basename
	cfg.Profiles["Games.exe"] = &ExecutableProfile{
		TargetExe:    "launcher/Games.exe",
		UseGamescope: &falseVal,
		UsePrimeRun:  &falseVal,
	}

	// 1. Should match when targetExe is "launcher/Games.exe"
	eff := cfg.GetEffectiveConfig("launcher/Games.exe")
	if eff.UseGamescope {
		t.Errorf("Expected UseGamescope=false when matching subfolder launcher/Games.exe against profile Games.exe")
	}
	if eff.UsePrimeRun {
		t.Errorf("Expected UsePrimeRun=false when matching subfolder launcher/Games.exe against profile Games.exe")
	}

	// 2. Should match case-insensitively
	effCase := cfg.GetEffectiveConfig("LAUNCHER/GAMES.EXE")
	if effCase.UseGamescope {
		t.Errorf("Expected case-insensitive match for subfolder profile")
	}
}

func TestCloudSyncConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := NewDefaultConfig()
	cfg.TargetExe = "Game.exe"
	cfg.CloudSync = &CloudSyncConfig{
		Backend:      "rclone",
		RemotePath:   "gdrive:GameSaves/TestGame",
		LocalPath:    "{PREFIX}/drive_c/users/steamuser/Saved Games",
		AutoSyncPre:  true,
		AutoSyncPost: true,
	}

	if err := SaveConfig(tmpDir, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.CloudSync == nil {
		t.Fatalf("expected loaded CloudSync to not be nil")
	}
	if loaded.CloudSync.Backend != "rclone" {
		t.Errorf("expected backend 'rclone', got '%s'", loaded.CloudSync.Backend)
	}
	if loaded.CloudSync.RemotePath != "gdrive:GameSaves/TestGame" {
		t.Errorf("expected remote path 'gdrive:GameSaves/TestGame', got '%s'", loaded.CloudSync.RemotePath)
	}
	if !loaded.CloudSync.AutoSyncPre || !loaded.CloudSync.AutoSyncPost {
		t.Errorf("expected auto sync flags to be true")
	}
}

func TestGamescopeConfiguration(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg.GamescopeRefresh != 0 {
		t.Errorf("Expected default GamescopeRefresh to be 0 (Native/Untouched), got %d", cfg.GamescopeRefresh)
	}
	if cfg.GamescopeScaling != "fit" {
		t.Errorf("Expected default scaling 'fit', got '%s'", cfg.GamescopeScaling)
	}
	if cfg.GamescopeFilter != "linear" {
		t.Errorf("Expected default filter 'linear', got '%s'", cfg.GamescopeFilter)
	}

	cfg.GamescopeAdaptiveSync = true
	cfg.GamescopeMangoApp = true
	cfg.GamescopeFilter = "fsr"
	cfg.GamescopeSharpness = 7
	cfg.GamescopeWindowMode = "borderless"
	cfg.GamescopeExtraArgs = []string{"--cursor-scale-height", "1080"}

	cpy := cfg.Clone()
	if !cpy.GamescopeAdaptiveSync || !cpy.GamescopeMangoApp || cpy.GamescopeFilter != "fsr" || cpy.GamescopeSharpness != 7 || cpy.GamescopeWindowMode != "borderless" {
		t.Errorf("Clone failed to preserve Gamescope fields")
	}
	if len(cpy.GamescopeExtraArgs) != 2 || cpy.GamescopeExtraArgs[0] != "--cursor-scale-height" {
		t.Errorf("Clone failed to copy GamescopeExtraArgs correctly")
	}
}

func TestResetToSafeDefaults(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.TargetExe = "Game.exe"
	cfg.ProtonPath = "/path/to/proton"
	cfg.AppID = "12345"

	// Modify settings to non-default complex values
	cfg.GamescopeWidth = 3840
	cfg.GamescopeHeight = 2160
	cfg.GamescopeRefresh = 144
	cfg.GamescopeFilter = "fsr"
	cfg.GamescopeSharpness = 9
	cfg.GamescopeHDR = true
	cfg.GamescopeMangoApp = true
	cfg.UseXalia = true
	cfg.EnableLogging = true
	cfg.DLLOverrides["d3d11"] = "native"
	cfg.ExtraArgs = []string{"-novid"}
	cfg.EnvVars["CUSTOM"] = "1"

	// Perform reset to safe defaults
	cfg.ResetToSafeDefaults(true, "0-7")

	// TargetExe, ProtonPath, and AppID must be preserved
	if cfg.TargetExe != "Game.exe" {
		t.Errorf("Expected TargetExe preserved as Game.exe, got %s", cfg.TargetExe)
	}
	if cfg.ProtonPath != "/path/to/proton" {
		t.Errorf("Expected ProtonPath preserved, got %s", cfg.ProtonPath)
	}
	if cfg.AppID != "12345" {
		t.Errorf("Expected AppID preserved as 12345, got %s", cfg.AppID)
	}

	// Gamescope must be reset to standard safe Auto (0x0), SDR, Linear, Untouched Hz
	if cfg.GamescopeWidth != 0 || cfg.GamescopeHeight != 0 {
		t.Errorf("Expected Gamescope Auto geometry (0x0), got %dx%d", cfg.GamescopeWidth, cfg.GamescopeHeight)
	}
	if cfg.GamescopeRefresh != 0 {
		t.Errorf("Expected GamescopeRefresh 0, got %d", cfg.GamescopeRefresh)
	}
	if cfg.GamescopeFilter != "linear" {
		t.Errorf("Expected GamescopeFilter linear, got %s", cfg.GamescopeFilter)
	}
	if cfg.GamescopeHDR || cfg.GamescopeMangoApp {
		t.Errorf("Expected HDR and MangoApp to be false")
	}

	// Tweaks must be cleared
	if cfg.UseXalia || cfg.EnableLogging {
		t.Errorf("Expected UseXalia and EnableLogging to be false")
	}
	if len(cfg.DLLOverrides) != 0 {
		t.Errorf("Expected DLLOverrides to be empty, got %v", cfg.DLLOverrides)
	}
	if len(cfg.ExtraArgs) != 0 || len(cfg.EnvVars) != 0 {
		t.Errorf("Expected ExtraArgs and EnvVars to be empty")
	}
}

func TestCleanZeroConfig(t *testing.T) {
	cfg := NewCleanZeroConfig()
	if cfg.UseGamescope {
		t.Errorf("Clean Zero must have UseGamescope=false")
	}
	if cfg.UsePrimeRun {
		t.Errorf("Clean Zero must have UsePrimeRun=false")
	}
	if cfg.UsePCores {
		t.Errorf("Clean Zero must have UsePCores=false")
	}
	if cfg.PCoresMask != "" {
		t.Errorf("Clean Zero must have empty PCoresMask, got %q", cfg.PCoresMask)
	}
	if cfg.ManagePower {
		t.Errorf("Clean Zero must have ManagePower=false")
	}
	if cfg.GamescopeOutput != "auto" {
		t.Errorf("Clean Zero must have GamescopeOutput='auto', got %q", cfg.GamescopeOutput)
	}
	if cfg.OpaqueBackdrop {
		t.Errorf("Clean Zero must have OpaqueBackdrop=false")
	}
}

func TestResetToCleanZero(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.TargetExe = "Game.exe"
	cfg.ProtonPath = "/opt/proton"
	cfg.AppID = "999"
	cfg.UseGamescope = true
	cfg.UsePrimeRun = true
	cfg.UsePCores = true
	cfg.PCoresMask = "0-7"
	cfg.ManagePower = true
	cfg.DLLOverrides["dxgi"] = "n"

	cfg.ResetToCleanZero()

	if cfg.TargetExe != "Game.exe" || cfg.ProtonPath != "/opt/proton" || cfg.AppID != "999" {
		t.Errorf("TargetExe, ProtonPath, and AppID must be preserved")
	}
	if cfg.UseGamescope || cfg.UsePrimeRun || cfg.UsePCores || cfg.ManagePower {
		t.Errorf("All performance wrappers must be disabled in Clean Zero")
	}
	if cfg.PCoresMask != "" {
		t.Errorf("PCoresMask must be empty in Clean Zero, got %q", cfg.PCoresMask)
	}
	if len(cfg.DLLOverrides) != 0 {
		t.Errorf("DLLOverrides must be empty in Clean Zero")
	}
}

func TestIsValidAppID(t *testing.T) {
	tests := []struct {
		appID    string
		expected bool
	}{
		{"", false},
		{"   ", false},
		{"0", false},
		{" 0 ", false},
		{"4567", true},
		{"225540", true},
	}

	for _, tt := range tests {
		if got := IsValidAppID(tt.appID); got != tt.expected {
			t.Errorf("IsValidAppID(%q) = %v, expected %v", tt.appID, got, tt.expected)
		}
	}

	var nilCfg *GameConfig
	if nilCfg.HasValidAppID() {
		t.Errorf("nil GameConfig.HasValidAppID() should be false")
	}

	cfg := &GameConfig{AppID: "0"}
	if cfg.HasValidAppID() {
		t.Errorf("GameConfig{AppID: 0}.HasValidAppID() should be false")
	}

	cfg.AppID = "12345"
	if !cfg.HasValidAppID() {
		t.Errorf("GameConfig{AppID: 12345}.HasValidAppID() should be true")
	}
}

func TestFormatDocumentedConfigFile(t *testing.T) {
	fileCfg := NewConfigFileWithDefault()
	active := fileCfg.GetActiveProfile()
	active.TargetExe = "Endfield.exe"
	active.ProtonPath = "/path/to/proton"
	active.AppID = "0"
	active.ExtraArgs = []string{"-vulkan"}
	active.PresetName = "Arknights: Endfield"
	active.UmuID = "umu-endfield"
	active.UseGamescope = true
	active.GamescopeWidth = 1920
	active.GamescopeHeight = 1080
	active.GamescopeRefresh = 75
	active.GeometryAutoDetected = true
	active.UsePrimeRun = true
	active.UsePCores = true
	active.PCoresMask = "0-11"
	active.EnvVars = map[string]string{
		"WINE_CANONICAL_HOLE": "skip_volatile_check",
		"VKD3D_CONFIG":        "no_upload_hvv",
	}

	formatted := FormatDocumentedConfigFile(fileCfg)

	// Verify required comments and documentation
	if !strings.Contains(formatted, "# auto detected from TUI") {
		t.Errorf("Expected formatted TOML to contain '# auto detected from TUI'")
	}
	if !strings.Contains(formatted, "gamescope_width = 1920") {
		t.Errorf("Expected formatted TOML to contain gamescope_width = 1920")
	}
	if !strings.Contains(formatted, "Prevents anti-cheat page fault crashes") {
		t.Errorf("Expected formatted TOML to contain WINE_CANONICAL_HOLE comment")
	}
	if !strings.Contains(formatted, "active_profile = \"default\"") {
		t.Errorf("Expected formatted TOML to declare active_profile")
	}

	// Verify round-trip unmarshaling
	tmpDir := t.TempDir()
	if err := SaveConfigFile(tmpDir, fileCfg); err != nil {
		t.Fatalf("SaveConfigFile failed: %v", err)
	}

	loaded, err := LoadConfigFile(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfigFile failed on documented TOML: %v", err)
	}

	loadedActive := loaded.GetActiveProfile()
	if loadedActive.TargetExe != "Endfield.exe" {
		t.Errorf("Expected TargetExe Endfield.exe, got %s", loadedActive.TargetExe)
	}
	if loadedActive.GamescopeWidth != 1920 || loadedActive.GamescopeHeight != 1080 {
		t.Errorf("Expected 1920x1080, got %dx%d", loadedActive.GamescopeWidth, loadedActive.GamescopeHeight)
	}
	if loadedActive.EnvVars["WINE_CANONICAL_HOLE"] != "skip_volatile_check" {
		t.Errorf("Expected WINE_CANONICAL_HOLE=skip_volatile_check, got %s", loadedActive.EnvVars["WINE_CANONICAL_HOLE"])
	}
}



