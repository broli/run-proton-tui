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

	if val, ok := env["VKD3D_CONFIG"]; ok {
		t.Errorf("Expected VKD3D_CONFIG to be unset by default (Clean Zero), got %s", val)
	}

	// Test with UmuID and CustomEnv
	customOpts := EnvOptions{
		PrefixDir: "/tmp/test-prefix",
		GameDir:   "/tmp/test-game",
		UmuID:     "umu-endfield",
		CustomEnv: map[string]string{
			"WINE_CANONICAL_HOLE": "skip_volatile_check",
			"VKD3D_CONFIG":        "no_upload_hvv",
		},
	}
	customEnv := BuildEnvironment(customOpts)
	if customEnv["UMU_ID"] != "umu-endfield" {
		t.Errorf("Expected UMU_ID=umu-endfield, got %s", customEnv["UMU_ID"])
	}
	if customEnv["WINE_CANONICAL_HOLE"] != "skip_volatile_check" {
		t.Errorf("Expected WINE_CANONICAL_HOLE=skip_volatile_check, got %s", customEnv["WINE_CANONICAL_HOLE"])
	}
	if customEnv["VKD3D_CONFIG"] != "no_upload_hvv" {
		t.Errorf("Expected VKD3D_CONFIG=no_upload_hvv from CustomEnv, got %s", customEnv["VKD3D_CONFIG"])
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
		t.Logf("Discovered %d runners successfully:", len(runners))
		for i, r := range runners {
			t.Logf("  [%2d] %-14s %-40s -> %s", i+1, r.Type, r.Name, r.Path)
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

func TestClassifyRunner(t *testing.T) {
	tests := []struct {
		name         string
		expectedType RunnerType
		isGE         bool
		isCachy      bool
		isDW         bool
		isExp        bool
		isValve      bool
	}{
		{"GE-Proton11-7", RunnerTypeGE, true, false, false, false, false},
		{"Proton-GE Latest", RunnerTypeGE, true, false, false, false, false},
		{"Proton-CachyOS Latest-x86_64_v3", RunnerTypeCachyOS, false, true, false, false, false},
		{"dwproton-11.0-14", RunnerTypeDW, false, false, true, false, false},
		{"DW-Proton Latest", RunnerTypeDW, false, false, true, false, false},
		{"Proton - Experimental", RunnerTypeExperimental, false, false, false, true, true},
		{"Proton 10.0", RunnerTypeValve, false, false, false, false, true},
		{"Proton 9.0 (Beta)", RunnerTypeValve, false, false, false, false, true},
		{"wine-lutris-GE-8.26", RunnerTypeWine, false, false, false, false, false},
		{"custom-runner", RunnerTypeCustom, false, false, false, false, false},
	}

	for _, tc := range tests {
		r := classifyRunner(tc.name, "/dummy/"+tc.name, "/dummy/"+tc.name+"/proton")
		if r.Type != tc.expectedType {
			t.Errorf("[%s] Expected Type %s, got %s", tc.name, tc.expectedType, r.Type)
		}
		if r.IsGE != tc.isGE {
			t.Errorf("[%s] Expected IsGE=%v, got %v", tc.name, tc.isGE, r.IsGE)
		}
		if r.IsCachy != tc.isCachy {
			t.Errorf("[%s] Expected IsCachy=%v, got %v", tc.name, tc.isCachy, r.IsCachy)
		}
		if r.IsDW != tc.isDW {
			t.Errorf("[%s] Expected IsDW=%v, got %v", tc.name, tc.isDW, r.IsDW)
		}
		if r.IsExperimental != tc.isExp {
			t.Errorf("[%s] Expected IsExperimental=%v, got %v", tc.name, tc.isExp, r.IsExperimental)
		}
		if r.IsValve != tc.isValve {
			t.Errorf("[%s] Expected IsValve=%v, got %v", tc.name, tc.isValve, r.IsValve)
		}
		if r.Recommendation == "" {
			t.Errorf("[%s] Expected non-empty recommendation", tc.name)
		}
	}
}

func TestNaturalCompare(t *testing.T) {
	// Should compare version numbers numerically: 10 > 9, 14 > 12
	if naturalCompare("Proton 10.0", "Proton 9.0 (Beta)") <= 0 {
		t.Errorf("Expected Proton 10.0 > Proton 9.0 (Beta)")
	}
	if naturalCompare("dwproton-11.0-14", "dwproton-11.0-12") <= 0 {
		t.Errorf("Expected dwproton-11.0-14 > dwproton-11.0-12")
	}
	if naturalCompare("GE-Proton11-7", "GE-Proton11-6") <= 0 {
		t.Errorf("Expected GE-Proton11-7 > GE-Proton11-6")
	}
	if naturalCompare("SameName", "SameName") != 0 {
		t.Errorf("Expected SameName == SameName")
	}
}
