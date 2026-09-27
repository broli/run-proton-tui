package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ShortcutLocation specifies where the .desktop shortcut should be saved.
type ShortcutLocation string

const (
	LocationApplications ShortcutLocation = "applications" // ~/.local/share/applications/
	LocationDesktop      ShortcutLocation = "desktop"      // ~/Desktop/
	LocationLocal        ShortcutLocation = "local"        // Game directory
)

// ShortcutOptions defines metadata and destinations for the desktop launcher.
type ShortcutOptions struct {
	GameTitle  string
	GameDir    string
	TargetExe  string
	Location   ShortcutLocation
	CustomIcon string
}

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify generates a filesystem-safe identifier for desktop files.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "game"
	}
	return s
}

// FindGameIcon looks for an icon in the game directory or returns a standard fallback.
func FindGameIcon(gameDir, customIcon string) string {
	if customIcon != "" {
		if _, err := os.Stat(customIcon); err == nil {
			return customIcon
		}
	}

	candidates := []string{
		filepath.Join(gameDir, "icon.png"),
		filepath.Join(gameDir, "icon.ico"),
		filepath.Join(gameDir, "logo.png"),
		filepath.Join(gameDir, ".rpt", "icon.png"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	// Standard Freedesktop fallback gaming icon
	return "applications-games"
}

// CreateDesktopShortcut generates a standard freedesktop.org .desktop entry.
func CreateDesktopShortcut(opts ShortcutOptions) (string, error) {
	if opts.GameDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get game directory: %w", err)
		}
		opts.GameDir = wd
	}

	absGameDir, err := filepath.Abs(opts.GameDir)
	if err == nil {
		opts.GameDir = absGameDir
	}

	if opts.GameTitle == "" {
		opts.GameTitle = filepath.Base(opts.GameDir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}

	slug := Slugify(opts.GameTitle)
	desktopFileName := fmt.Sprintf("rpt-%s.desktop", slug)

	var targetDir string
	switch opts.Location {
	case LocationDesktop:
		targetDir = filepath.Join(home, "Desktop")
	case LocationLocal:
		targetDir = opts.GameDir
	case LocationApplications:
		fallthrough
	default:
		targetDir = filepath.Join(home, ".local", "share", "applications")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory %s: %w", targetDir, err)
	}

	destPath := filepath.Join(targetDir, desktopFileName)

	// Determine rpt binary path
	rptBin := "rpt"
	if self, err := os.Executable(); err == nil {
		rptBin = self
	}

	icon := FindGameIcon(opts.GameDir, opts.CustomIcon)

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Comment=Launch %s via rpt (Run Proton TUI)
Exec="%s" --now
Path=%s
Icon=%s
Terminal=false
Categories=Game;
StartupNotify=true
X-rpt-target-exe=%s
`, opts.GameTitle, opts.GameTitle, rptBin, opts.GameDir, icon, opts.TargetExe)

	if err := os.WriteFile(destPath, []byte(content), 0755); err != nil {
		return "", fmt.Errorf("failed to write desktop file: %w", err)
	}

	return destPath, nil
}
