package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Arknights: Endfield", "arknights-endfield"},
		{"Just Cause (2006)!", "just-cause-2006"},
		{"   Game   Name   ", "game-name"},
		{"", "game"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestCreateDesktopShortcut(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rpt-desktop-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gameDir := filepath.Join(tmpDir, "MyGame")
	_ = os.MkdirAll(gameDir, 0755)

	opts := ShortcutOptions{
		GameTitle: "Test Game",
		GameDir:   gameDir,
		TargetExe: "game.exe",
		Location:  LocationLocal,
	}

	path, err := CreateDesktopShortcut(opts)
	if err != nil {
		t.Fatalf("CreateDesktopShortcut failed: %v", err)
	}

	if !strings.HasSuffix(path, "rpt-test-game.desktop") {
		t.Errorf("unexpected shortcut path: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read desktop file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "Name=Test Game") {
		t.Errorf("missing Name in desktop file: %s", s)
	}
	if !strings.Contains(s, "Exec=") || !strings.Contains(s, "--now") {
		t.Errorf("missing Exec command with --now: %s", s)
	}
	if !strings.Contains(s, "Path="+gameDir) {
		t.Errorf("missing Path in desktop file: %s", s)
	}
	if !strings.Contains(s, "Categories=Game;") {
		t.Errorf("missing Categories in desktop file: %s", s)
	}
}

func TestCreateDesktopShortcutInjection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rpt-desktop-inject-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	opts := ShortcutOptions{
		GameTitle: "Malicious\nExec=malicious-command\nComment=pwned",
		GameDir:   tmpDir,
		TargetExe: "game.exe",
		Location:  LocationLocal,
	}

	path, err := CreateDesktopShortcut(opts)
	if err != nil {
		t.Fatalf("CreateDesktopShortcut failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read desktop file: %v", err)
	}

	s := string(content)
	// Verify that the title does not inject a raw newline creating an Exec=malicious-command key
	lines := strings.Split(s, "\n")
	execCount := 0
	for _, l := range lines {
		if strings.HasPrefix(l, "Exec=") {
			execCount++
			if strings.Contains(l, "malicious-command") {
				t.Fatalf("Vulnerability detected: malicious command injected into Exec line: %s", l)
			}
		}
	}
	if execCount != 1 {
		t.Errorf("Expected exactly 1 Exec line, got %d", execCount)
	}
}
