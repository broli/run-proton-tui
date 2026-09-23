package proton

import (
	"os"
	"strings"
	"testing"
)

func TestBuildEnvironment(t *testing.T) {
	opts := EnvOptions{
		PrefixDir:     "/tmp/test-prefix",
		GameDir:       "/tmp/test-game",
		AppID:         "12345",
		UseXalia:      false,
		EnableLogging: true,
	}

	env := BuildEnvironment(opts)

	if env["STEAM_COMPAT_APP_ID"] != "12345" {
		t.Errorf("Expected STEAM_COMPAT_APP_ID=12345, got %s", env["STEAM_COMPAT_APP_ID"])
	}

	if env["SteamGameId"] != "12345" {
		t.Errorf("Expected SteamGameId=12345, got %s", env["SteamGameId"])
	}

	if env["UMU_USE_STEAM"] != "0" {
		t.Errorf("Expected UMU_USE_STEAM=0, got %s", env["UMU_USE_STEAM"])
	}

	if env["PROTON_USE_XALIA"] != "0" {
		t.Errorf("Expected PROTON_USE_XALIA=0, got %s", env["PROTON_USE_XALIA"])
	}

	if !strings.Contains(env["WINEDLLOVERRIDES"], "lsteamclient=d") {
		t.Errorf("Expected lsteamclient=d in WINEDLLOVERRIDES, got %s", env["WINEDLLOVERRIDES"])
	}

	// Test with UmuID and CustomEnv
	customOpts := EnvOptions{
		PrefixDir: "/tmp/test-prefix",
		GameDir:   "/tmp/test-game",
		UmuID:     "umu-endfield",
		CustomEnv: map[string]string{
			"WINE_CANONICAL_HOLE": "skip_volatile_check",
		},
	}
	customEnv := BuildEnvironment(customOpts)
	if customEnv["UMU_ID"] != "umu-endfield" {
		t.Errorf("Expected UMU_ID=umu-endfield, got %s", customEnv["UMU_ID"])
	}
	if customEnv["WINE_CANONICAL_HOLE"] != "skip_volatile_check" {
		t.Errorf("Expected WINE_CANONICAL_HOLE=skip_volatile_check, got %s", customEnv["WINE_CANONICAL_HOLE"])
	}
}

func TestDiscoverRunners(t *testing.T) {
	runners, err := DiscoverRunners()
	if err != nil {
		t.Fatalf("DiscoverRunners failed: %v", err)
	}

	// Carlos has multiple Proton versions installed
	if len(runners) == 0 {
		t.Log("Note: No runners discovered in standard paths on this test run.")
	} else {
		t.Logf("Discovered %d runners successfully", len(runners))
		for _, r := range runners {
			if _, err := os.Stat(r.Path); err != nil {
				t.Errorf("Runner path does not exist: %s", r.Path)
			}
		}
	}
}

func Test2DUtilityEnvStripping(t *testing.T) {
	// 1. When 2D Utility
	opts2D := EnvOptions{
		PrefixDir:   "/tmp/test-pfx",
		GameDir:     "/tmp/test-game",
		UsePrimeRun: true,
		Is2DUtility: true,
	}
	env2D := BuildEnvironment(opts2D)
	if _, ok := env2D["PROTON_ENABLE_NVAPI"]; ok {
		t.Errorf("Expected PROTON_ENABLE_NVAPI to be stripped for 2D utility, but it was present")
	}
	if _, ok := env2D["DXVK_ENABLE_NVAPI"]; ok {
		t.Errorf("Expected DXVK_ENABLE_NVAPI to be stripped for 2D utility, but it was present")
	}
	if _, ok := env2D["STEAMOS"]; ok {
		t.Errorf("Expected STEAMOS to be stripped for 2D utility, but it was present")
	}
	if _, ok := env2D["STEAMDECK"]; ok {
		t.Errorf("Expected STEAMDECK to be stripped for 2D utility, but it was present")
	}

	// 2. When 3D game with prime-run
	opts3D := EnvOptions{
		PrefixDir:   "/tmp/test-pfx",
		GameDir:     "/tmp/test-game",
		UsePrimeRun: true,
		Is2DUtility: false,
	}
	env3D := BuildEnvironment(opts3D)
	if env3D["PROTON_ENABLE_NVAPI"] != "1" {
		t.Errorf("Expected PROTON_ENABLE_NVAPI=1 for 3D game, got %s", env3D["PROTON_ENABLE_NVAPI"])
	}
	if env3D["STEAMOS"] != "1" {
		t.Errorf("Expected STEAMOS=1 for 3D game, got %s", env3D["STEAMOS"])
	}
}
