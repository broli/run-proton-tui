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
