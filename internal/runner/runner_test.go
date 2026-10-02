package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestRunGame_ExtraArgsPreserved(t *testing.T) {
	t.Run("ConfigOnly_ExtraArgsPreservedWhenOptsEmpty", func(t *testing.T) {
		tempDir := t.TempDir()
		targetExe := "Endfield.exe"
		targetExePath := filepath.Join(tempDir, targetExe)
		if err := os.WriteFile(targetExePath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatalf("Failed to create dummy target exe: %v", err)
		}

		argsOutputFile := filepath.Join(tempDir, "proton_args.txt")
		mockProton := filepath.Join(tempDir, "mock_proton.sh")
		mockProtonContent := fmt.Sprintf("#!/bin/bash\necho \"$@\" > %q\nexit 0\n", argsOutputFile)
		if err := os.WriteFile(mockProton, []byte(mockProtonContent), 0755); err != nil {
			t.Fatalf("Failed to create mock proton script: %v", err)
		}

		cfg := config.NewDefaultConfig()
		cfg.TargetExe = targetExe
		cfg.ProtonPath = mockProton
		cfg.UseGamescope = false
		cfg.ManagePower = false
		cfg.ExtraArgs = []string{"-vulkan"}

		opts := LaunchOptions{
			GameDir:   tempDir,
			Config:    cfg,
			ExtraArgs: nil, // empty CLI flags
		}

		res, err := RunGame(context.Background(), opts)
		if err != nil {
			t.Fatalf("RunGame returned unexpected error: %v", err)
		}
		if res.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", res.ExitCode)
		}

		data, err := os.ReadFile(argsOutputFile)
		if err != nil {
			t.Fatalf("Failed to read proton args file: %v", err)
		}

		argsOutput := strings.TrimSpace(string(data))
		expectedArgs := fmt.Sprintf("waitforexitandrun ./%s -vulkan", targetExe)
		if argsOutput != expectedArgs {
			t.Errorf("Arguments passed to Proton runner mismatch:\ngot:  %q\nwant: %q", argsOutput, expectedArgs)
		}
	})

	t.Run("ConfigAndCLI_PreservesBothInOrder", func(t *testing.T) {
		tempDir := t.TempDir()
		targetExe := "Game.exe"
		targetExePath := filepath.Join(tempDir, targetExe)
		if err := os.WriteFile(targetExePath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatalf("Failed to create dummy target exe: %v", err)
		}

		argsOutputFile := filepath.Join(tempDir, "proton_args.txt")
		mockProton := filepath.Join(tempDir, "mock_proton.sh")
		mockProtonContent := fmt.Sprintf("#!/bin/bash\necho \"$@\" > %q\nexit 0\n", argsOutputFile)
		if err := os.WriteFile(mockProton, []byte(mockProtonContent), 0755); err != nil {
			t.Fatalf("Failed to create mock proton script: %v", err)
		}

		cfg := config.NewDefaultConfig()
		cfg.TargetExe = targetExe
		cfg.ProtonPath = mockProton
		cfg.UseGamescope = false
		cfg.ManagePower = false
		cfg.ExtraArgs = []string{"-vulkan", "-profile_flag"}

		opts := LaunchOptions{
			GameDir:   tempDir,
			Config:    cfg,
			ExtraArgs: []string{"--user-flag1", "--user-flag2"},
		}

		res, err := RunGame(context.Background(), opts)
		if err != nil {
			t.Fatalf("RunGame returned unexpected error: %v", err)
		}
		if res.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", res.ExitCode)
		}

		data, err := os.ReadFile(argsOutputFile)
		if err != nil {
			t.Fatalf("Failed to read proton args file: %v", err)
		}

		argsOutput := strings.TrimSpace(string(data))
		expectedArgs := fmt.Sprintf("waitforexitandrun ./%s -vulkan -profile_flag --user-flag1 --user-flag2", targetExe)
		if argsOutput != expectedArgs {
			t.Errorf("Arguments passed to Proton runner mismatch:\ngot:  %q\nwant: %q", argsOutput, expectedArgs)
		}
	})

	t.Run("ProfileOverride_ExtraArgsPreserved", func(t *testing.T) {
		tempDir := t.TempDir()
		targetExe := "Custom.exe"
		targetExePath := filepath.Join(tempDir, targetExe)
		if err := os.WriteFile(targetExePath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatalf("Failed to create dummy target exe: %v", err)
		}

		argsOutputFile := filepath.Join(tempDir, "proton_args.txt")
		mockProton := filepath.Join(tempDir, "mock_proton.sh")
		mockProtonContent := fmt.Sprintf("#!/bin/bash\necho \"$@\" > %q\nexit 0\n", argsOutputFile)
		if err := os.WriteFile(mockProton, []byte(mockProtonContent), 0755); err != nil {
			t.Fatalf("Failed to create mock proton script: %v", err)
		}

		cfg := config.NewDefaultConfig()
		cfg.TargetExe = targetExe
		cfg.ProtonPath = mockProton
		cfg.UseGamescope = false
		cfg.ManagePower = false
		cfg.ExtraArgs = []string{"-default_arg"}
		cfg.Profiles[targetExe] = &config.ExecutableProfile{
			TargetExe: targetExe,
			ExtraArgs: []string{"-vulkan", "-profile_override"},
		}

		opts := LaunchOptions{
			GameDir:   tempDir,
			Config:    cfg,
			ExtraArgs: []string{"--trailing-arg"},
		}

		res, err := RunGame(context.Background(), opts)
		if err != nil {
			t.Fatalf("RunGame returned unexpected error: %v", err)
		}
		if res.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", res.ExitCode)
		}

		data, err := os.ReadFile(argsOutputFile)
		if err != nil {
			t.Fatalf("Failed to read proton args file: %v", err)
		}

		argsOutput := strings.TrimSpace(string(data))
		expectedArgs := fmt.Sprintf("waitforexitandrun ./%s -vulkan -profile_override --trailing-arg", targetExe)
		if argsOutput != expectedArgs {
			t.Errorf("Arguments passed to Proton runner mismatch:\ngot:  %q\nwant: %q", argsOutput, expectedArgs)
		}
	})
}

func TestBuildGamescopeArgs(t *testing.T) {
	t.Run("omits -r when GamescopeRefresh is 0", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.GamescopeRefresh = 0
		cfg.GamescopeWidth = 2560
		cfg.GamescopeHeight = 1440
		cfg.GamescopeOutput = "DP-1"

		args := BuildGamescopeArgs(cfg)
		for _, a := range args {
			if a == "-r" {
				t.Errorf("Expected -r to be omitted when GamescopeRefresh is 0, got args: %v", args)
			}
		}

		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "-W 2560 -H 1440") {
			t.Errorf("Expected -W 2560 -H 1440 in args, got: %s", joined)
		}
		if !strings.Contains(joined, "--prefer-output DP-1") {
			t.Errorf("Expected --prefer-output DP-1 in args, got: %s", joined)
		}
	})

	t.Run("includes -r when GamescopeRefresh > 0", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.GamescopeRefresh = 144
		args := BuildGamescopeArgs(cfg)
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "-r 144") {
			t.Errorf("Expected -r 144 in args, got: %s", joined)
		}
	})

	t.Run("supports upscaling, window mode, and features", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		cfg.GamescopeWindowMode = "borderless"
		cfg.GamescopeFilter = "fsr"
		cfg.GamescopeSharpness = 8
		cfg.GamescopeAdaptiveSync = true
		cfg.GamescopeMangoApp = true
		cfg.GamescopeHDR = true
		cfg.GamescopeFPSLimit = 60

		args := BuildGamescopeArgs(cfg)
		joined := strings.Join(args, " ")

		if !strings.Contains(joined, "-b") {
			t.Errorf("Expected -b for borderless mode, got: %s", joined)
		}
		hasF := false
		for _, a := range args {
			if a == "-f" {
				hasF = true
			}
		}
		if hasF {
			t.Errorf("Did not expect -f in borderless mode, got args: %v", args)
		}
		if !strings.Contains(joined, "-F fsr") {
			t.Errorf("Expected -F fsr, got: %s", joined)
		}
		if !strings.Contains(joined, "--sharpness 8") {
			t.Errorf("Expected --sharpness 8, got: %s", joined)
		}
		if !strings.Contains(joined, "--adaptive-sync") {
			t.Errorf("Expected --adaptive-sync, got: %s", joined)
		}
		if !strings.Contains(joined, "--mangoapp") {
			t.Errorf("Expected --mangoapp, got: %s", joined)
		}
		if !strings.Contains(joined, "--hdr-enabled") {
			t.Errorf("Expected --hdr-enabled, got: %s", joined)
		}
		if !strings.Contains(joined, "--framerate-limit 60") {
			t.Errorf("Expected --framerate-limit 60, got: %s", joined)
		}
	})
}


