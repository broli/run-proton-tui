package integrations

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanEmulators(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock steam_appid.txt
	_ = os.WriteFile(filepath.Join(tmpDir, "steam_appid.txt"), []byte("1245620\n"), 0644)

	// Create mock Goldberg configs.app.ini
	_ = os.WriteFile(filepath.Join(tmpDir, "configs.app.ini"), []byte("[App]\nunlock_all=1\n"), 0644)

	info, err := ScanEmulators(tmpDir)
	if err != nil {
		t.Fatalf("ScanEmulators failed: %v", err)
	}

	if !info.Detected {
		t.Errorf("Expected emulator to be detected, got false")
	}
	if info.Type != "Goldberg Emulator" {
		t.Errorf("Expected Type='Goldberg Emulator', got %q", info.Type)
	}
	if info.AppID != "1245620" {
		t.Errorf("Expected AppID='1245620', got %q", info.AppID)
	}
	if info.DLCStatus != "All Unlocked (unlock_all=1)" {
		t.Errorf("Expected DLCStatus='All Unlocked (unlock_all=1)', got %q", info.DLCStatus)
	}
}

func TestProtonDBLiveAPI(t *testing.T) {
	// Query Elden Ring (1245620)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	report, err := FetchProtonDBReport(ctx, "1245620")
	if err != nil {
		t.Logf("ProtonDB network query skipped or errored (might be offline): %v", err)
		return
	}

	if report.Tier == "" {
		t.Errorf("Expected non-empty Tier from ProtonDB")
	} else {
		t.Logf("ProtonDB Report for 1245620: Tier=%s, Badge=%s, Confidence=%s", report.Tier, report.GetTierBadge(), report.Confidence)
	}
}

func TestCleanGameTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Just Cause [DODI Repack]", "Just Cause"},
		{"Styx - Master of Shadows - [DODI Repack]", "Styx - Master of Shadows"},
		{"Elden_Ring_(v1.12)_[FitGirl]", "Elden Ring"},
		{"Cyberpunk 2077 (GOG)", "Cyberpunk 2077"},
	}

	for _, tt := range tests {
		got := CleanGameTitle(tt.input)
		if got != tt.want {
			t.Errorf("CleanGameTitle(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveSteamAppID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	aid, name, err := ResolveSteamAppID(ctx, "Arknights: Endfield", "games/Arknights Endfield/Endfield.exe", "GRYPHLINK")
	if err != nil {
		t.Logf("Network query skipped or failed: %v", err)
		return
	}

	if aid != "4732690" {
		t.Errorf("Expected AppID '4732690', got %q (%s)", aid, name)
	} else {
		t.Logf("Resolved Endfield to AppID %s (%s)", aid, name)
	}
}
