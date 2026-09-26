package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHooksViewRenderingAndPager(t *testing.T) {
	tmpDir := t.TempDir()
	hooksDir := filepath.Join(tmpDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	testScript := `#!/usr/bin/env bash
set -euo pipefail
echo "Hook running for $RPT_TARGET_EXE in $RPT_GAME_DIR"
`
	scriptPath := filepath.Join(hooksDir, "pre_launch.sh")
	if err := os.WriteFile(scriptPath, []byte(testScript), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewDefaultConfig()
	cfg.TargetExe = "TestGame.exe"

	view := NewHooksView(tmpDir, cfg, 100, 30)

	// Verify overview renders
	rendered := view.View()
	if !strings.Contains(rendered, "LIFECYCLE HOOK INSPECTOR") {
		t.Errorf("expected view to contain header, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "ACTIVE:") {
		t.Errorf("expected view to show active pre-launch hook, got:\n%s", rendered)
	}

	// Press "1" to open pre_launch.sh pager
	view, closeRequested := view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if closeRequested {
		t.Errorf("did not expect close on pressing 1")
	}
	if !view.ViewingScript {
		t.Errorf("expected ViewingScript to be true after pressing 1")
	}

	// Verify pager view renders with script content
	pagerRendered := view.View()
	if !strings.Contains(pagerRendered, "Viewing Hook:") {
		t.Errorf("expected pager header, got:\n%s", pagerRendered)
	}
	if !strings.Contains(pagerRendered, "RPT_TARGET_EXE") {
		t.Errorf("expected script content in pager, got:\n%s", pagerRendered)
	}

	// Press Esc to exit pager
	view, closeRequested = view.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if closeRequested {
		t.Errorf("did not expect close on first Esc from pager")
	}
	if view.ViewingScript {
		t.Errorf("expected ViewingScript to be false after pressing Esc")
	}

	// Press Esc to exit hooks view
	_, closeRequested = view.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !closeRequested {
		t.Errorf("expected closeRequested true on Esc from overview")
	}
}
