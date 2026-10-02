package pathutil

import (
	"path/filepath"
	"testing"
)

func TestUserHomeDir(t *testing.T) {
	home := UserHomeDir()
	if home == "" {
		t.Error("expected non-empty UserHomeDir")
	}
}

func TestSanitize(t *testing.T) {
	home := UserHomeDir()
	if home == "" || home == "/" {
		t.Skip("skipping test in environment without home directory")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "exact home directory",
			input:    home,
			expected: "~",
		},
		{
			name:     "file under home directory",
			input:    filepath.Join(home, "Games", "Cyberpunk", "game.exe"),
			expected: "~" + string(filepath.Separator) + filepath.Join("Games", "Cyberpunk", "game.exe"),
		},
		{
			name:     "log excerpt with home path embedded",
			input:    "wine: Unhandled exception at " + filepath.Join(home, ".wine", "drive_c") + " code 0xc0000005",
			expected: "wine: Unhandled exception at ~" + string(filepath.Separator) + filepath.Join(".wine", "drive_c") + " code 0xc0000005",
		},
		{
			name:     "system path outside home",
			input:    "/usr/share/steam/compatibilitytools.d",
			expected: "/usr/share/steam/compatibilitytools.d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sanitize(tt.input)
			if got != tt.expected {
				t.Errorf("Sanitize(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExpand(t *testing.T) {
	home := UserHomeDir()
	if home == "" {
		t.Skip("skipping test in environment without home directory")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "tilde only",
			input:    "~",
			expected: home,
		},
		{
			name:     "tilde slash path",
			input:    "~/Games/Backups",
			expected: filepath.Join(home, "Games", "Backups"),
		},
		{
			name:     "tilde backslash path",
			input:    "~\\Games\\Backups",
			expected: filepath.Join(home, "Games", "Backups"),
		},
		{
			name:     "absolute path unchanged",
			input:    "/opt/games/test",
			expected: "/opt/games/test",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path/to/file",
			expected: "relative/path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Expand(tt.input)
			if got != tt.expected {
				t.Errorf("Expand(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExpandWithBase(t *testing.T) {
	home := UserHomeDir()
	baseDir := "/var/games/mygame"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "tilde path",
			input:    "~/saves",
			expected: filepath.Join(home, "saves"),
		},
		{
			name:     "relative path",
			input:    "hooks/pre-launch.sh",
			expected: filepath.Join(baseDir, "hooks", "pre-launch.sh"),
		},
		{
			name:     "absolute path",
			input:    "/usr/bin/custom-hook",
			expected: "/usr/bin/custom-hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandWithBase(baseDir, tt.input)
			if got != tt.expected {
				t.Errorf("ExpandWithBase(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveTemplate(t *testing.T) {
	home := UserHomeDir()
	gameDir := "/games/Endfield"
	prefixDir := "/games/Endfield/proton-prefix"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "prefix macro",
			input:    "{PREFIX}/pfx/drive_c/users/steamuser/AppData",
			expected: "/games/Endfield/proton-prefix/pfx/drive_c/users/steamuser/AppData",
		},
		{
			name:     "game dir macro",
			input:    "{GAME_DIR}/saves",
			expected: "/games/Endfield/saves",
		},
		{
			name:     "host pictures macro",
			input:    "{HOST_PICTURES}/Endfield",
			expected: filepath.Join(home, "Pictures", "Endfield"),
		},
		{
			name:     "tilde path",
			input:    "~/Pictures/Endfield",
			expected: filepath.Join(home, "Pictures", "Endfield"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveTemplate(tt.input, gameDir, prefixDir)
			if got != tt.expected {
				t.Errorf("ResolveTemplate(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cyberpunk 2077", "cyberpunk-2077"},
		{"Arknights: Endfield", "arknights-endfield"},
		{"Game with Special Characters! @ # $ %", "game-with-special-characters"},
		{"", "game"},
		{"---multiple---dashes---", "multiple-dashes"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsIgnoredDir(t *testing.T) {
	if !IsIgnoredDir("proton-prefix") {
		t.Error("expected proton-prefix to be ignored")
	}
	if !IsIgnoredDir(".logs") {
		t.Error("expected .logs to be ignored")
	}
	if !IsIgnoredDir(".git") {
		t.Error("expected .git to be ignored")
	}
	if IsIgnoredDir("GameData") {
		t.Error("GameData should not be ignored")
	}
}
