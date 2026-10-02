package proton

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// RunnerType identifies the compatibility tool flavor.
type RunnerType string

const (
	RunnerTypeGE           RunnerType = "GE-Proton"
	RunnerTypeCachyOS      RunnerType = "CachyOS"
	RunnerTypeDW           RunnerType = "DW-Proton"
	RunnerTypeExperimental RunnerType = "Experimental"
	RunnerTypeValve        RunnerType = "Valve Proton"
	RunnerTypeWine         RunnerType = "Wine"
	RunnerTypeCustom       RunnerType = "Custom"
)

// Runner represents an installed Proton or Wine compatibility tool.
type Runner struct {
	Name           string     // e.g. "GE-Proton11-7", "dwproton-11.0-14"
	Path           string     // Path to 'proton' binary
	Directory      string     // Base directory of the runner
	Type           RunnerType // Tool family / classification
	IsGE           bool       // GloriousEggroll custom build
	IsCachy        bool       // CachyOS optimized build
	IsDW           bool       // Dawn Winery gacha build
	IsExperimental bool       // Valve Proton Experimental
	IsValve        bool       // Valve official Proton release
	Recommendation string     // Short badge/hint for the TUI
}

func runnerTier(r Runner) int {
	switch r.Type {
	case RunnerTypeGE:
		return 1
	case RunnerTypeCachyOS:
		return 2
	case RunnerTypeDW:
		return 3
	case RunnerTypeExperimental:
		return 4
	case RunnerTypeValve:
		return 5
	case RunnerTypeWine:
		return 6
	default:
		return 7
	}
}

// naturalCompare compares two strings taking embedded integer numbers into account.
func naturalCompare(s1, s2 string) int {
	i, j := 0, 0
	r1, r2 := []rune(strings.ToLower(s1)), []rune(strings.ToLower(s2))
	for i < len(r1) && j < len(r2) {
		if unicode.IsDigit(r1[i]) && unicode.IsDigit(r2[j]) {
			startI, startJ := i, j
			for i < len(r1) && unicode.IsDigit(r1[i]) {
				i++
			}
			for j < len(r2) && unicode.IsDigit(r2[j]) {
				j++
			}
			num1, err1 := strconv.ParseUint(string(r1[startI:i]), 10, 64)
			num2, err2 := strconv.ParseUint(string(r2[startJ:j]), 10, 64)
			if err1 == nil && err2 == nil {
				if num1 != num2 {
					if num1 > num2 {
						return 1
					}
					return -1
				}
			} else {
				if string(r1[startI:i]) != string(r2[startJ:j]) {
					if string(r1[startI:i]) > string(r2[startJ:j]) {
						return 1
					}
					return -1
				}
			}
		} else {
			if r1[i] != r2[j] {
				if r1[i] > r2[j] {
					return 1
				}
				return -1
			}
			i++
			j++
		}
	}
	if len(r1) > len(r2) {
		return 1
	} else if len(r1) < len(r2) {
		return -1
	}
	return 0
}

func classifyRunner(name, dirPath, protonBin string) Runner {
	lower := strings.ToLower(name)

	var rType RunnerType
	var isGE, isCachy, isDW, isExp, isValve bool
	var rec string

	switch {
	case strings.Contains(lower, "ge-proton") || strings.Contains(lower, "proton-ge"):
		rType = RunnerTypeGE
		isGE = true
		rec = "GloriousEggroll: High game compatibility, media codecs & ProtonFixes"
	case strings.Contains(lower, "cachyos"):
		rType = RunnerTypeCachyOS
		isCachy = true
		rec = "CachyOS: x86-64-v3/v4 & NTSync tuned performance"
	case strings.Contains(lower, "dwproton") || strings.Contains(lower, "dw-proton"):
		rType = RunnerTypeDW
		isDW = true
		rec = "Dawn Winery: Optimized for Hoyoverse & anime gacha titles"
	case strings.Contains(lower, "experimental"):
		rType = RunnerTypeExperimental
		isExp = true
		isValve = true
		rec = "Valve Experimental: Bleeding edge Proton features & latest fixes"
	case strings.HasPrefix(lower, "proton") || strings.Contains(lower, "proton"):
		rType = RunnerTypeValve
		isValve = true
		rec = "Valve Official: Stable Steam Linux Runtime compatibility tool"
	case strings.Contains(lower, "wine") || strings.Contains(lower, "lutris") || strings.Contains(lower, "heroic"):
		rType = RunnerTypeWine
		rec = "Wine Runner: General Windows compatibility layer"
	default:
		rType = RunnerTypeCustom
		rec = "Custom compatibility runner"
	}

	return Runner{
		Name:           name,
		Path:           protonBin,
		Directory:      dirPath,
		Type:           rType,
		IsGE:           isGE,
		IsCachy:        isCachy,
		IsDW:           isDW,
		IsExperimental: isExp,
		IsValve:        isValve,
		Recommendation: rec,
	}
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

	seenDirs := make(map[string]bool)
	var uniqueSearchDirs []string
	for _, sDir := range searchDirs {
		realDir, err := filepath.EvalSymlinks(sDir)
		if err != nil {
			if _, statErr := os.Stat(sDir); statErr == nil {
				realDir = sDir
			} else {
				continue
			}
		}
		if seenDirs[realDir] {
			continue
		}
		seenDirs[realDir] = true
		uniqueSearchDirs = append(uniqueSearchDirs, sDir)
	}

	runnersMap := make(map[string]Runner)

	for _, sDir := range uniqueSearchDirs {
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
				realBin, err := filepath.EvalSymlinks(protonBin)
				if err != nil {
					realBin = protonBin
				}

				if _, exists := runnersMap[realBin]; exists {
					continue
				}

				runner := classifyRunner(entry.Name(), dirPath, protonBin)
				runnersMap[realBin] = runner
			}
		}
	}

	results := make([]Runner, 0, len(runnersMap))
	for _, r := range runnersMap {
		results = append(results, r)
	}

	// Sort runners: grouped by family tier, "Latest" first, then natural version descending
	sort.Slice(results, func(i, j int) bool {
		r1, r2 := results[i], results[j]
		t1, t2 := runnerTier(r1), runnerTier(r2)
		if t1 != t2 {
			return t1 < t2
		}

		l1 := strings.Contains(strings.ToLower(r1.Name), "latest")
		l2 := strings.Contains(strings.ToLower(r2.Name), "latest")
		if l1 != l2 {
			return l1
		}

		cmp := naturalCompare(r1.Name, r2.Name)
		if cmp != 0 {
			return cmp > 0
		}
		return r1.Name < r2.Name
	})

	return results, nil
}

