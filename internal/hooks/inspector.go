package hooks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
)

// CascadeCandidate describes a candidate location checked during hook resolution.
type CascadeCandidate struct {
	Path     string `json:"path"`
	Source   string `json:"source"`
	Exists   bool   `json:"exists"`
	Selected bool   `json:"selected"`
}

// EnvVarDoc documents an environment variable injected into lifecycle hooks.
type EnvVarDoc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CurrentVal  string `json:"current_val"`
}

// HookInspectionDetails contains full resolution metadata for a lifecycle phase.
type HookInspectionDetails struct {
	HookType   HookType           `json:"hook_type"`
	Resolved   string             `json:"resolved"`
	Candidates []CascadeCandidate `json:"candidates"`
}

// FullHooksReport contains both pre-launch and post-exit inspection reports.
type FullHooksReport struct {
	PreLaunch HookInspectionDetails `json:"pre_launch"`
	PostExit  HookInspectionDetails  `json:"post_exit"`
	EnvVars   []EnvVarDoc            `json:"env_vars"`
}

// HighlightBash takes a bash script and formats it with ANSI colors using Chroma.
func HighlightBash(source string) (string, error) {
	var buf bytes.Buffer
	// Use terminal256 formatter and dracula theme for vibrant, terminal-compatible syntax colors
	err := quick.Highlight(&buf, source, "bash", "terminal256", "dracula")
	if err != nil {
		return source, err
	}
	return buf.String(), nil
}

// InspectCandidates evaluates all possible locations in priority cascade order.
func InspectCandidates(gameDir string, hookType HookType, configHookPath string, extraDirs ...string) HookInspectionDetails {
	scriptName := string(hookType) + ".sh"
	var candidates []CascadeCandidate

	// 1. Explicit config path
	if configHookPath != "" {
		target := configHookPath
		if strings.HasPrefix(target, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				target = filepath.Join(home, target[2:])
			}
		} else if !filepath.IsAbs(target) {
			target = filepath.Join(gameDir, target)
		}
		exists := fileExists(target)
		candidates = append(candidates, CascadeCandidate{
			Path:   target,
			Source: "Configuration (.proton-config.toml)",
			Exists: exists,
		})
	}

	// 2. Local game directory candidates
	localList := []struct {
		rel    string
		source string
	}{
		{filepath.Join("hooks", scriptName), "Local Game hooks/ Directory"},
		{filepath.Join(".rpt", "hooks", scriptName), "Local Game .rpt/hooks/ Directory"},
		{scriptName, "Game Root Directory"},
	}
	for _, l := range localList {
		fullPath := filepath.Join(gameDir, l.rel)
		candidates = append(candidates, CascadeCandidate{
			Path:   fullPath,
			Source: l.source,
			Exists: fileExists(fullPath),
		})
	}

	// 3. Extra custom hook directories
	for _, dir := range extraDirs {
		candidate := dir
		if strings.HasPrefix(candidate, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				candidate = filepath.Join(home, candidate[2:])
			}
		} else if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(gameDir, candidate)
		}
		fullPath := filepath.Join(candidate, scriptName)
		candidates = append(candidates, CascadeCandidate{
			Path:   fullPath,
			Source: fmt.Sprintf("Custom Hook Directory (%s)", dir),
			Exists: fileExists(fullPath),
		})
	}

	// 4. User global directories
	if home, err := os.UserHomeDir(); err == nil {
		globalConfig := filepath.Join(home, ".config", "rpt", "hooks", scriptName)
		candidates = append(candidates, CascadeCandidate{
			Path:   globalConfig,
			Source: "User Global Config (~/.config/rpt/hooks/)",
			Exists: fileExists(globalConfig),
		})

		globalLegacy := filepath.Join(home, ".rpt", "hooks", scriptName)
		candidates = append(candidates, CascadeCandidate{
			Path:   globalLegacy,
			Source: "User Legacy Global (~/.rpt/hooks/)",
			Exists: fileExists(globalLegacy),
		})
	}

	// Determine resolved path (first existing candidate)
	resolved := ""
	for i := range candidates {
		if candidates[i].Exists {
			candidates[i].Selected = true
			resolved = candidates[i].Path
			break
		}
	}

	return HookInspectionDetails{
		HookType:   hookType,
		Resolved:   resolved,
		Candidates: candidates,
	}
}

// InspectAllHooks compiles an inspection report for both lifecycle phases and environment variables.
func InspectAllHooks(opts HookOptions) FullHooksReport {
	pre := InspectCandidates(opts.GameDir, PreLaunch, opts.ConfigHookPath, opts.ExtraHookDirs...)
	post := InspectCandidates(opts.GameDir, PostExit, opts.ConfigHookPath, opts.ExtraHookDirs...)

	envDocs := []EnvVarDoc{
		{
			Name:        "RPT_GAME_DIR",
			Description: "Absolute path to the game installation root directory",
			CurrentVal:  opts.GameDir,
		},
		{
			Name:        "RPT_PREFIX_DIR",
			Description: "Absolute path to the isolated Wine prefix directory (.prefix)",
			CurrentVal:  opts.PrefixDir,
		},
		{
			Name:        "RPT_TARGET_EXE",
			Description: "Selected target Windows executable relative to game root",
			CurrentVal:  opts.TargetExe,
		},
		{
			Name:        "RPT_PROTON_PATH",
			Description: "Absolute path to the active Proton runner installation",
			CurrentVal:  opts.ProtonPath,
		},
		{
			Name:        "RPT_GAMESCOPE_DISPLAY",
			Description: "Gamescope nested display identifier (e.g. :1, or empty if disabled)",
			CurrentVal:  opts.GamescopeDisplay,
		},
		{
			Name:        "RPT_HOOK_TYPE",
			Description: "Lifecycle phase currently executing ('pre_launch' or 'post_exit')",
			CurrentVal:  "pre_launch | post_exit",
		},
		{
			Name:        "RPT_CHILD_PID",
			Description: "Process ID of the game process (available in post_exit)",
			CurrentVal:  "$PID (post_exit only)",
		},
	}

	return FullHooksReport{
		PreLaunch: pre,
		PostExit:  post,
		EnvVars:   envDocs,
	}
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}
