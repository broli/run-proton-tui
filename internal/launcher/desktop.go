package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/pathutil"
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


// Slugify generates a filesystem-safe identifier for desktop files.
func Slugify(s string) string {
	return pathutil.Slugify(s)
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

	home := pathutil.UserHomeDir()

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

	cleanTitle := sanitizeDesktopValue(opts.GameTitle)
	cleanDir := sanitizeDesktopValue(opts.GameDir)
	cleanExe := sanitizeDesktopValue(opts.TargetExe)
	icon := FindGameIcon(opts.GameDir, opts.CustomIcon)
	cleanIcon := sanitizeDesktopValue(icon)
	execCmd := fmt.Sprintf("%s --now", quoteExecArg(rptBin))

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Comment=Launch %s via rpt (Run Proton TUI)
Exec=%s
Path=%s
Icon=%s
Terminal=false
Categories=Game;
StartupNotify=true
X-rpt-target-exe=%s
`, cleanTitle, cleanTitle, execCmd, cleanDir, cleanIcon, cleanExe)

	if err := os.WriteFile(destPath, []byte(content), 0755); err != nil {
		return "", fmt.Errorf("failed to write desktop file: %w", err)
	}

	return destPath, nil
}

func sanitizeDesktopValue(val string) string {
	val = strings.ReplaceAll(val, "\r", "")
	val = strings.ReplaceAll(val, "\n", " ")
	return strings.TrimSpace(val)
}

func quoteExecArg(arg string) string {
	arg = sanitizeDesktopValue(arg)
	if strings.ContainsAny(arg, " \t\"'") {
		arg = strings.ReplaceAll(arg, `\`, `\\`)
		arg = strings.ReplaceAll(arg, `"`, `\"`)
		return `"` + arg + `"`
	}
	return arg
}
