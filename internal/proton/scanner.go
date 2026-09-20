package proton

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Runner represents an installed Proton or Wine compatibility tool.
type Runner struct {
	Name           string // e.g. "GE-Proton11-6"
	Path           string // Path to 'proton' binary
	Directory      string // Base directory of the runner
	IsGE           bool   // GloriousEggroll custom build
	IsCachy        bool   // CachyOS optimized build
	IsExperimental bool   // Valve Proton Experimental
	Recommendation string // Short badge/hint for the TUI
}

// DiscoverRunners scans all common Steam and custom compatibility tool locations.
func DiscoverRunners() ([]Runner, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	searchDirs := []string{
		filepath.Join(home, ".local/share/Steam/compatibilitytools.d"),
		filepath.Join(home, ".steam/root/compatibilitytools.d"),
		filepath.Join(home, ".steam/steam/compatibilitytools.d"),
		"/usr/share/steam/compatibilitytools.d",
		filepath.Join(home, ".local/share/Steam/steamapps/common"),
		filepath.Join(home, ".config/heroic/tools/wine"),
		filepath.Join(home, ".config/heroic/tools/proton"),
		filepath.Join(home, ".local/share/lutris/runners/wine"),
	}

	runnersMap := make(map[string]Runner)

	for _, sDir := range searchDirs {
		entries, err := os.ReadDir(sDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirPath := filepath.Join(sDir, entry.Name())
			protonBin := filepath.Join(dirPath, "proton")

			// Check if proton binary exists and is executable
			if info, err := os.Stat(protonBin); err == nil && !info.IsDir() {
				name := entry.Name()
				lower := strings.ToLower(name)

				isGE := strings.Contains(lower, "ge-proton") || strings.Contains(lower, "proton-ge")
				isCachy := strings.Contains(lower, "cachyos")
				isExp := strings.Contains(lower, "experimental")

				rec := ""
				if isGE {
					rec = "GE-Proton: High Game Compatibility & Codecs"
				} else if isCachy {
					rec = "CachyOS: x86-64-v3/v4 & NTSync Tuned"
				} else if isExp {
					rec = "Valve Experimental: Bleeding Edge"
				}

				runner := Runner{
					Name:           name,
					Path:           protonBin,
					Directory:      dirPath,
					IsGE:           isGE,
					IsCachy:        isCachy,
					IsExperimental: isExp,
					Recommendation: rec,
				}

				runnersMap[runner.Path] = runner
			}
		}
	}

	results := make([]Runner, 0, len(runnersMap))
	for _, r := range runnersMap {
		results = append(results, r)
	}

	// Sort runners: prefer GE-Proton and CachyOS at the top, then alphabetically
	sort.Slice(results, func(i, j int) bool {
		r1, r2 := results[i], results[j]
		if r1.IsGE != r2.IsGE {
			return r1.IsGE
		}
		if r1.IsCachy != r2.IsCachy {
			return r1.IsCachy
		}
		return r1.Name > r2.Name
	})

	return results, nil
}
