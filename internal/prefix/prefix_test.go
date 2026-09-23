package prefix

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckPrefixContainment(t *testing.T) {
	tmpDir := t.TempDir()
	prefixDir := filepath.Join(tmpDir, "proton-prefix")
	_ = os.MkdirAll(filepath.Join(prefixDir, "pfx", "drive_c", "Game"), 0755)

	insideExe := filepath.Join(prefixDir, "pfx", "drive_c", "Game", "game.exe")
	outsideExe := filepath.Join(tmpDir, "Games", "MyGame", "game.exe")

	if !CheckPrefixContainment(prefixDir, insideExe) {
		t.Errorf("Expected insideExe to be detected inside prefix, got false")
	}

	if CheckPrefixContainment(prefixDir, outsideExe) {
		t.Errorf("Expected outsideExe to be detected outside prefix, got true")
	}
}

func TestGetPrefixSocketName(t *testing.T) {
	tmpDir := t.TempDir()
	name, err := GetPrefixSocketName(tmpDir)
	if err != nil {
		t.Fatalf("GetPrefixSocketName failed: %v", err)
	}

	if name == "" {
		t.Error("Expected non-empty socket name")
	}

	if len(name) < 8 || name[:7] != "server-" {
		t.Errorf("Expected socket name starting with 'server-', got %q", name)
	}
}

func TestPreserveSavesWithPicturesAndCustomDir(t *testing.T) {
	tmpDir := t.TempDir()
	prefixDir := filepath.Join(tmpDir, "proton-prefix")
	usersDir := filepath.Join(prefixDir, "pfx", "drive_c", "users", "steamuser")
	picDir := filepath.Join(usersDir, "Pictures", "MyGame")
	_ = os.MkdirAll(picDir, 0755)

	testPhoto := filepath.Join(picDir, "screenshot.png")
	_ = os.WriteFile(testPhoto, []byte("fake_png_data"), 0644)

	customBackupDir := filepath.Join(tmpDir, "custom_preservation")
	dest, err := PreserveSaves(PreserveOptions{
		PrefixDir: prefixDir,
		GameName:  "TestGame",
		DestDir:   customBackupDir,
	})
	if err != nil {
		t.Fatalf("PreserveSaves failed: %v", err)
	}

	if dest == "" {
		t.Fatalf("Expected non-empty destination path")
	}

	preservedPhoto := filepath.Join(dest, "steamuser", "Pictures", "MyGame", "screenshot.png")
	if _, err := os.Stat(preservedPhoto); err != nil {
		t.Errorf("Expected screenshot.png to be preserved in Pictures/, but stat failed: %v", err)
	}
}
