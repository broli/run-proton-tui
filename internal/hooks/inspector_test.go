package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHighlightBash(t *testing.T) {
	script := `#!/usr/bin/env bash
set -euo pipefail
echo "Starting test with $RPT_GAME_DIR"
if [ ! -d "$RPT_PREFIX_DIR" ]; then
    echo "Prefix not initialized"
fi
`
	highlighted, err := HighlightBash(script)
	if err != nil {
		t.Fatalf("HighlightBash failed: %v", err)
	}

	// Output must contain ANSI color escape sequences
	if !strings.Contains(highlighted, "\x1b[") {
		t.Errorf("expected ANSI escape codes in highlighted bash output, got:\n%s", highlighted)
	}
}

func TestInspectCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	hooksDir := filepath.Join(tmpDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	scriptPath := filepath.Join(hooksDir, "pre_launch.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho ok"), 0755); err != nil {
		t.Fatal(err)
	}

	details := InspectCandidates(tmpDir, PreLaunch, "")
	if details.Resolved != scriptPath {
		t.Errorf("expected resolved path %s, got %s", scriptPath, details.Resolved)
	}

	foundSelected := false
	for _, cand := range details.Candidates {
		if cand.Selected {
			foundSelected = true
			if cand.Path != scriptPath {
				t.Errorf("expected selected candidate to be %s, got %s", scriptPath, cand.Path)
			}
		}
	}
	if !foundSelected {
		t.Errorf("expected at least one candidate to be marked Selected")
	}
}

func TestInspectAllHooks(t *testing.T) {
	tmpDir := t.TempDir()
	report := InspectAllHooks(HookOptions{
		GameDir:          tmpDir,
		PrefixDir:        filepath.Join(tmpDir, ".prefix"),
		TargetExe:        "Game.exe",
		ProtonPath:       "/opt/proton",
		GamescopeDisplay: ":1",
	})

	if len(report.EnvVars) == 0 {
		t.Errorf("expected documented environment variables")
	}

	foundGameDir := false
	for _, env := range report.EnvVars {
		if env.Name == "RPT_GAME_DIR" && env.CurrentVal == tmpDir {
			foundGameDir = true
			break
		}
	}
	if !foundGameDir {
		t.Errorf("expected RPT_GAME_DIR env var with value %s", tmpDir)
	}
}
