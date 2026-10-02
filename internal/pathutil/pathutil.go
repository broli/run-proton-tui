package pathutil

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// UserHomeDir returns the current user's home directory.
// It queries os.UserHomeDir() and falls back to os.Getenv("HOME").
func UserHomeDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home
	}
	return os.Getenv("HOME")
}

// Sanitize replaces all occurrences of the user's home directory with "~".
// This ensures that user paths and usernames in logs, error dumps, crash reports,
// telemetry, and issue submissions remain completely anonymous and private.
func Sanitize(p string) string {
	if p == "" {
		return ""
	}
	home := UserHomeDir()
	if home == "" || home == "/" {
		return p
	}
	return strings.ReplaceAll(p, home, "~")
}

// Expand substitutes leading "~" or "~/" with the user's home directory.
// If the path does not start with "~", it is returned unchanged.
func Expand(p string) string {
	if p == "" {
		return ""
	}
	if p == "~" {
		return UserHomeDir()
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		home := UserHomeDir()
		if home != "" {
			sub := strings.ReplaceAll(p[2:], "\\", string(filepath.Separator))
			return filepath.Join(home, sub)
		}
	}
	return p
}

// ExpandWithBase expands any leading "~" in path. If the resulting path is relative,
// it is joined with baseDir to return a fully resolved path.
func ExpandWithBase(baseDir, p string) string {
	expanded := Expand(p)
	if expanded == "" {
		return ""
	}
	if !filepath.IsAbs(expanded) && baseDir != "" {
		return filepath.Join(baseDir, expanded)
	}
	return expanded
}

// ResolveTemplate resolves common configuration macros ({PREFIX}, {GAME_DIR},
// {HOST_HOME}, {HOST_PICTURES}) and expands leading tildes (~).
func ResolveTemplate(template, gameDir, prefixDir string) string {
	if template == "" {
		return ""
	}
	home := UserHomeDir()
	p := template
	if prefixDir != "" {
		p = strings.ReplaceAll(p, "{PREFIX}", prefixDir)
	}
	if gameDir != "" {
		p = strings.ReplaceAll(p, "{GAME_DIR}", gameDir)
	}
	if home != "" {
		p = strings.ReplaceAll(p, "{HOST_HOME}", home)
		p = strings.ReplaceAll(p, "{HOST_PICTURES}", filepath.Join(home, "Pictures"))
	}
	return Expand(p)
}

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify generates a clean, filesystem-safe alphanumeric identifier (e.g. for filenames and desktop entries).
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "game"
	}
	return s
}

// IsIgnoredDir reports whether a directory name should be excluded from recursive game directory scans.
func IsIgnoredDir(name string) bool {
	return name == "proton-prefix" || name == ".logs" || name == ".git"
}
