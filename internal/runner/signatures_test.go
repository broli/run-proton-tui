package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSignatures_DefaultAndMatching(t *testing.T) {
	sigs := GetDefaultSignatures()
	if len(sigs) == 0 {
		t.Fatalf("expected default signatures to not be empty")
	}

	antiCheatSig := sigs[0]
	if !MatchSignature("ProtonFixes: ace-base.sys loaded into memory", antiCheatSig) {
		t.Errorf("expected MatchSignature to match ace-base.sys")
	}

	if MatchSignature("Clean game shutdown", antiCheatSig) {
		t.Errorf("expected clean line not to match anti-cheat signature")
	}
}

func TestSignatures_LoadExternal(t *testing.T) {
	tmpDir := t.TempDir()
	rptDir := filepath.Join(tmpDir, ".rpt")
	_ = os.MkdirAll(rptDir, 0755)

	customTOML := `[[signatures]]
category = "Custom Game Error"
keywords = ["custom_error_0x123", "crash_handler"]
observation = "Detected custom game crash pattern."
recommendation = "Apply custom launch parameter."
`
	_ = os.WriteFile(filepath.Join(rptDir, "signatures.toml"), []byte(customTOML), 0644)

	loaded := LoadSignatures(tmpDir)
	foundCustom := false
	for _, s := range loaded {
		if s.Category == "Custom Game Error" {
			foundCustom = true
			if !MatchSignature("Fatal: custom_error_0x123 encountered", s) {
				t.Errorf("expected custom keyword to match line")
			}
		}
	}

	if !foundCustom {
		t.Errorf("expected custom signature from signatures.toml to be loaded")
	}
}
