package quirks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/pelletier/go-toml/v2"
)

// ManifestQuirk represents a single game quirk profile loaded from a declarative TOML manifest.
type ManifestQuirk struct {
	Name            string                               `toml:"name"`
	MatchExe        []string                             `toml:"match_exe"`
	MatchAppID      string                               `toml:"match_appid,omitempty"`
	SummaryNotes    string                               `toml:"summary_notes,omitempty"`
	UmuID           string                               `toml:"umu_id,omitempty"`
	EnvVars         map[string]string                    `toml:"env_vars,omitempty"`
	ExtraArgs       []string                             `toml:"extra_args,omitempty"`
	WaitProcesses   []string                             `toml:"wait_processes,omitempty"`
	DisplayFile     string                               `toml:"display_file,omitempty"`
	RecommendedHook string                               `toml:"recommended_hook,omitempty"`
	Profiles        map[string]*config.ExecutableProfile `toml:"profiles,omitempty"`
}

// QuirksManifest wraps an array of manifest quirks in a TOML file.
type QuirksManifest struct {
	Quirks []ManifestQuirk `toml:"quirks"`
}

// DiscoverManifestPaths returns all candidate quirk manifest paths in priority order.
func DiscoverManifestPaths(gameDir string) []string {
	var paths []string

	// 1. Local game directory manifests
	paths = append(paths,
		filepath.Join(gameDir, ".rpt", "quirks.toml"),
		filepath.Join(gameDir, "quirks.toml"),
	)
	if entries, err := os.ReadDir(filepath.Join(gameDir, ".rpt", "quirks")); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".toml") {
				paths = append(paths, filepath.Join(gameDir, ".rpt", "quirks", e.Name()))
			}
		}
	}

	// 2. User global config manifests (~/.config/rpt/quirks/*.toml)
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "rpt", "quirks.toml"))
		userQuirksDir := filepath.Join(home, ".config", "rpt", "quirks")
		if entries, err := os.ReadDir(userQuirksDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".toml") {
					paths = append(paths, filepath.Join(userQuirksDir, e.Name()))
				}
			}
		}
	}

	// 3. System manifests (/usr/share/rpt/quirks/*.toml)
	sysQuirksDir := "/usr/share/rpt/quirks"
	if entries, err := os.ReadDir(sysQuirksDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".toml") {
				paths = append(paths, filepath.Join(sysQuirksDir, e.Name()))
			}
		}
	}

	return paths
}

// LoadManifestQuirks loads and aggregates quirks from all discovered manifest files.
func LoadManifestQuirks(gameDir string) []ManifestQuirk {
	candidatePaths := DiscoverManifestPaths(gameDir)
	var allQuirks []ManifestQuirk

	for _, p := range candidatePaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}

		var manifest QuirksManifest
		if err := toml.Unmarshal(data, &manifest); err == nil && len(manifest.Quirks) > 0 {
			allQuirks = append(allQuirks, manifest.Quirks...)
		}
	}

	return allQuirks
}

// MatchManifestQuirk checks if any declarative manifest quirk matches the target game.
func MatchManifestQuirk(gameDir, exePath, appID string) *Preset {
	quirks := LoadManifestQuirks(gameDir)
	if len(quirks) == 0 {
		return nil
	}

	exeName := strings.ToLower(filepath.Base(exePath))
	dirName := strings.ToLower(filepath.Base(gameDir))

	for _, q := range quirks {
		// Match AppID if specified
		if q.MatchAppID != "" && appID != "" && strings.EqualFold(q.MatchAppID, appID) {
			return convertManifestToPreset(q)
		}

		// Match executable patterns
		for _, pattern := range q.MatchExe {
			patLower := strings.ToLower(pattern)
			if strings.EqualFold(exeName, patLower) || strings.Contains(exeName, patLower) || strings.Contains(dirName, patLower) {
				return convertManifestToPreset(q)
			}
		}
	}

	return nil
}

func convertManifestToPreset(q ManifestQuirk) *Preset {
	notes := q.SummaryNotes
	if q.RecommendedHook != "" {
		if notes != "" {
			notes += " "
		}
		notes += "(Recommended Hook: " + q.RecommendedHook + ")"
	}
	return &Preset{
		Name:          q.Name,
		MatchedSource: "Declarative Quirks Manifest",
		SummaryNotes:  notes,
		UmuID:         q.UmuID,
		EnvVars:       q.EnvVars,
		ExtraArgs:     q.ExtraArgs,
		WaitProcesses: q.WaitProcesses,
		DisplayFile:   q.DisplayFile,
		Profiles:      q.Profiles,
	}
}
