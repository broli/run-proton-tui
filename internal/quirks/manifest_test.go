package quirks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeclarativeQuirksManifest(t *testing.T) {
	tmpDir := t.TempDir()
	rptDir := filepath.Join(tmpDir, ".rpt")
	if err := os.MkdirAll(rptDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `
[[quirks]]
name = "Custom Modded Game"
match_exe = ["customgame.exe", "custom.exe"]
match_appid = "999999"
summary_notes = "Custom declarative quirk for test"
umu_id = "umu-custom"
extra_args = ["-customflag"]
wait_processes = ["CustomLauncher.exe"]
display_file = "/tmp/custom-display"
recommended_hook = "hooks/pre_launch.sh"

[quirks.env_vars]
CUSTOM_ENV = "1"
WINE_TEST = "true"

[quirks.profiles."custom.exe"]
target_exe = "custom.exe"
use_gamescope = false
`
	manifestPath := filepath.Join(rptDir, "quirks.toml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Test LoadManifestQuirks
	quirks := LoadManifestQuirks(tmpDir)
	if len(quirks) != 1 {
		t.Fatalf("expected 1 quirk loaded, got %d", len(quirks))
	}
	if quirks[0].Name != "Custom Modded Game" {
		t.Errorf("expected quirk name 'Custom Modded Game', got '%s'", quirks[0].Name)
	}

	// 2. Test Match by Executable name
	matched := MatchManifestQuirk(tmpDir, "customgame.exe", "")
	if matched == nil {
		t.Fatalf("expected MatchManifestQuirk to match customgame.exe")
	}
	if matched.Name != "Custom Modded Game" {
		t.Errorf("expected matched name 'Custom Modded Game', got '%s'", matched.Name)
	}
	if matched.EnvVars["CUSTOM_ENV"] != "1" {
		t.Errorf("expected CUSTOM_ENV=1, got '%s'", matched.EnvVars["CUSTOM_ENV"])
	}
	if !strings.Contains(matched.SummaryNotes, "Recommended Hook: hooks/pre_launch.sh") {
		t.Errorf("expected recommended hook note in summary, got '%s'", matched.SummaryNotes)
	}

	// 3. Test DetectQuirks with declarative manifest override
	preset := DetectQuirks(tmpDir, "customgame.exe", "", "")
	if preset == nil {
		t.Fatalf("expected DetectQuirks to find manifest preset")
	}
	if preset.MatchedSource != "Declarative Quirks Manifest" {
		t.Errorf("expected MatchedSource 'Declarative Quirks Manifest', got '%s'", preset.MatchedSource)
	}

	// 4. Test Match by AppID
	matchedAppID := MatchManifestQuirk(tmpDir, "unknown.exe", "999999")
	if matchedAppID == nil {
		t.Fatalf("expected MatchManifestQuirk to match appid 999999")
	}
	if matchedAppID.Name != "Custom Modded Game" {
		t.Errorf("expected matched name 'Custom Modded Game', got '%s'", matchedAppID.Name)
	}
}
