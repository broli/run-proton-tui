package quirks

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/proton"
)

// Preset represents a detected set of game-specific quirks and optimizations.
type Preset struct {
	Name          string
	MatchedSource string
	SummaryNotes  string
	UmuID         string
	EnvVars       map[string]string
	ExtraArgs     []string
	WaitProcesses []string
	DisplayFile   string
	Profiles      map[string]*config.ExecutableProfile
}

// ApplyToConfig hydrates a GameConfig with the settings from this Preset.
// If overwrite is false, it only populates empty fields and merges missing env vars/profiles.
// If overwrite is true, it replaces existing preset fields.
func (p *Preset) ApplyToConfig(cfg *config.GameConfig, overwrite bool) {
	if p == nil || cfg == nil {
		return
	}

	cfg.PresetName = p.Name
	if overwrite || cfg.UmuID == "" {
		if p.UmuID != "" {
			cfg.UmuID = p.UmuID
		}
	}
	if overwrite || cfg.DisplayFile == "" {
		if p.DisplayFile != "" {
			cfg.DisplayFile = p.DisplayFile
		}
	}

	if overwrite {
		if len(p.ExtraArgs) > 0 {
			cfg.ExtraArgs = make([]string, len(p.ExtraArgs))
			copy(cfg.ExtraArgs, p.ExtraArgs)
		}
		if len(p.WaitProcesses) > 0 {
			cfg.WaitProcesses = make([]string, len(p.WaitProcesses))
			copy(cfg.WaitProcesses, p.WaitProcesses)
		}
	} else {
		if len(p.ExtraArgs) > 0 && len(cfg.ExtraArgs) == 0 {
			cfg.ExtraArgs = make([]string, len(p.ExtraArgs))
			copy(cfg.ExtraArgs, p.ExtraArgs)
		}
		if len(p.WaitProcesses) > 0 && len(cfg.WaitProcesses) == 0 {
			cfg.WaitProcesses = make([]string, len(p.WaitProcesses))
			copy(cfg.WaitProcesses, p.WaitProcesses)
		}
	}

	if len(p.EnvVars) > 0 {
		if cfg.EnvVars == nil {
			cfg.EnvVars = make(map[string]string)
		}
		for k, v := range p.EnvVars {
			if overwrite {
				cfg.EnvVars[k] = v
			} else if _, exists := cfg.EnvVars[k]; !exists {
				cfg.EnvVars[k] = v
			}
		}
	}

	if len(p.Profiles) > 0 {
		if cfg.Profiles == nil {
			cfg.Profiles = make(map[string]*config.ExecutableProfile)
		}
		for k, v := range p.Profiles {
			if overwrite {
				cfg.Profiles[k] = v
			} else if _, exists := cfg.Profiles[k]; !exists {
				cfg.Profiles[k] = v
			}
		}
	}
}

// ClearPreset removes active preset metadata from the config while preserving user hardware/runner preferences.
func ClearPreset(cfg *config.GameConfig) {
	if cfg == nil {
		return
	}
	cfg.PresetName = ""
	cfg.UmuID = ""
	cfg.DisplayFile = ""
	cfg.WaitProcesses = make([]string, 0)
}

// CuratedPresets contains tested configurations for major standalone/non-Steam titles.
var CuratedPresets = map[string]Preset{
	"endfield": {
		Name:          "Arknights: Endfield",
		MatchedSource: "Curated Game Quirks Engine",
		SummaryNotes:  "Redirects 64 IL2CPP JIT mmap files to /dev/shm (RAM) via UMU, skips volatile memory check for ACE anti-cheat, enables -vulkan, registers display socket, and configures dual Launcher/Game profiles.",
		UmuID:         "umu-endfield",
		EnvVars: map[string]string{
			"WINE_CANONICAL_HOLE": "skip_volatile_check",
			"PROTON_USE_XALIA":    "0",
			"VKD3D_CONFIG":        "no_upload_hvv",
		},
		ExtraArgs:     []string{"-vulkan"},
		WaitProcesses: []string{"Updater.exe", "7zg.exe", "Patch.exe", "Games.exe"},
		DisplayFile:   "/tmp/gamescope-endfield-display",
		Profiles: map[string]*config.ExecutableProfile{
			"Launcher.exe": {
				TargetExe:    "Launcher.exe",
				UseGamescope: boolPtr(false),
				UsePrimeRun:  boolPtr(false),
				UsePCores:    boolPtr(false),
				ExtraArgs:    []string{},
			},
			"Games.exe": {
				TargetExe:    "Games.exe",
				UseGamescope: boolPtr(false),
				UsePrimeRun:  boolPtr(false),
				UsePCores:    boolPtr(false),
				ExtraArgs:    []string{},
			},
		},
	},
	"genshin": {
		Name:          "Genshin Impact",
		MatchedSource: "Curated Game Quirks Engine",
		SummaryNotes:  "Activates UMU Genshin compatibility shims and suppresses volatile memory checks for anti-cheat stability.",
		UmuID:         "umu-genshin",
		EnvVars: map[string]string{
			"WINE_CANONICAL_HOLE": "skip_volatile_check",
			"PROTON_USE_XALIA":    "0",
		},
		WaitProcesses: []string{"launcher.exe"},
	},
	"nikke": {
		Name:          "Goddess of Victory: NIKKE",
		MatchedSource: "Curated Game Quirks Engine",
		SummaryNotes:  "Activates UMU Nikke compatibility patches and suppresses anti-cheat memory conflicts.",
		UmuID:         "umu-nikke",
		EnvVars: map[string]string{
			"WINE_CANONICAL_HOLE": "skip_volatile_check",
			"PROTON_USE_XALIA":    "0",
		},
	},
	"repack_installer": {
		Name:          "Repack Installer / Setup",
		MatchedSource: "Installer Detection Engine",
		SummaryNotes:  "Configures 2D installer/unpacker to run natively on host display without Gamescope sandboxing, sharing the prefix with the game.",
		Profiles: map[string]*config.ExecutableProfile{
			"Setup.exe": {
				TargetExe:    "Setup.exe",
				UseGamescope: boolPtr(false),
				UsePrimeRun:  boolPtr(false),
				UsePCores:    boolPtr(false),
			},
		},
	},
}

func boolPtr(b bool) *bool {
	return &b
}

// DetectQuirks scans the game directory, executable name, Steam AppID, and local UMU database
// to suggest optimal runtime quirks and profiles.
func DetectQuirks(gameDir, exePath, appID, protonPath string) *Preset {
	dirName := strings.ToLower(filepath.Base(gameDir))
	exeName := strings.ToLower(filepath.Base(exePath))

	// 1. Check Declarative Quirks Manifests (.rpt/quirks.toml, ~/.config/rpt/quirks/*.toml, /usr/share/rpt/quirks/*.toml)
	if manifestPreset := MatchManifestQuirk(gameDir, exePath, appID); manifestPreset != nil {
		return manifestPreset
	}

	// 2. Check Curated Presets
	if strings.Contains(dirName, "endfield") || strings.Contains(dirName, "gryphlink") || strings.Contains(exeName, "endfield") {
		p := CuratedPresets["endfield"]
		return &p
	}

	if strings.Contains(dirName, "genshin") || strings.Contains(exeName, "genshin") {
		p := CuratedPresets["genshin"]
		return &p
	}

	if strings.Contains(dirName, "nikke") || strings.Contains(exeName, "nikke") {
		p := CuratedPresets["nikke"]
		return &p
	}

	// 2. Repack / Setup Installer Detection
	if exeName == "setup.exe" || exeName == "installer.exe" {
		p := CuratedPresets["repack_installer"]
		return &p
	}

	// 3. Scan Local UMU Database
	if preset := scanLocalUMUDatabase(dirName, exeName, appID, protonPath); preset != nil {
		return preset
	}

	return nil
}

// scanLocalUMUDatabase locates umu-database.csv in installed Proton paths and searches for matches.
func scanLocalUMUDatabase(dirName, exeName, appID, protonPath string) *Preset {
	csvPaths := findUMUDatabasePaths(protonPath)
	for _, csvPath := range csvPaths {
		if preset := scanSingleUMUDatabase(csvPath, dirName, exeName, appID); preset != nil {
			return preset
		}
	}
	return nil
}

func scanSingleUMUDatabase(csvPath, dirName, exeName, appID string) *Preset {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	// Avoid false positive substring matching if directory is a generic container
	isGenericDir := dirName == "games" || dirName == "game" || dirName == "bin" || dirName == "rpt" || dirName == "prefix" || len(dirName) < 3

	// Header: TITLE,STORE,CODENAME,UMU_ID,COMMON ACRONYM (Optional),NOTE (Optional),EXE_STRINGS (Optional)
	isHeader := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 4 {
			continue
		}
		if isHeader {
			isHeader = false
			continue
		}

		title := strings.ToLower(strings.TrimSpace(record[0]))
		store := strings.ToLower(strings.TrimSpace(record[1]))
		codename := strings.ToLower(strings.TrimSpace(record[2]))
		umuID := strings.TrimSpace(record[3])
		acronym := ""
		if len(record) > 4 {
			acronym = strings.ToLower(strings.TrimSpace(record[4]))
		}
		exeStrings := ""
		if len(record) > 6 {
			exeStrings = strings.ToLower(strings.TrimSpace(record[6]))
		}

		// Matching logic
		matched := false
		if appID != "" && appID != "0" && (codename == appID || umuID == "umu-"+appID) {
			matched = true
		} else if title != "" && !isGenericDir && (strings.Contains(dirName, title) || (len(dirName) >= 4 && strings.Contains(title, dirName))) {
			matched = true
		} else if acronym != "" && dirName == acronym {
			matched = true
		} else if exeStrings != "" && exeName != "" && strings.Contains(exeStrings, exeName) {
			matched = true
		}

		if matched && umuID != "" {
			return &Preset{
				Name:          record[0],
				MatchedSource: "UMU Database (" + store + ")",
				SummaryNotes:  "Matched in Proton UMU gamefixes database. Activates upstream ProtonFixes recipe.",
				UmuID:         umuID,
				EnvVars:       make(map[string]string),
				WaitProcesses: make([]string, 0),
				Profiles:      make(map[string]*config.ExecutableProfile),
			}
		}
	}
	return nil
}

// findUMUDatabasePaths returns candidate paths for umu-database.csv.
func findUMUDatabasePaths(protonPath string) []string {
	var paths []string

	// Check protonPath parent directory
	if protonPath != "" {
		pDir := filepath.Dir(protonPath) // e.g. /.../dwproton-11.0-12 or /.../dwproton-11.0-12/bin
		paths = append(paths, filepath.Join(pDir, "protonfixes", "umu-database.csv"))
		paths = append(paths, filepath.Join(filepath.Dir(pDir), "protonfixes", "umu-database.csv"))
	}

	// Standard Steam compatibilitytools.d paths
	for _, steamTools := range proton.GetCompatibilityToolDirs() {
		if entries, err := os.ReadDir(steamTools); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					paths = append(paths, filepath.Join(steamTools, entry.Name(), "protonfixes", "umu-database.csv"))
				}
			}
		}
	}

	return paths
}

// GetActiveOrDetectedPreset returns the currently configured quirks preset if present,
// or performs auto-detection across curated rules and the UMU database.
func GetActiveOrDetectedPreset(gameDir, exePath, appID, protonPath string, cfg *config.GameConfig) *Preset {
	if cfg != nil && cfg.PresetName != "" {
		// If preset is Repack Installer / Setup, but the target exe is no longer setup.exe or installer.exe, don't keep returning it
		exeName := strings.ToLower(filepath.Base(exePath))
		if strings.EqualFold(cfg.PresetName, "Repack Installer / Setup") && exeName != "setup.exe" && exeName != "installer.exe" && exeName != "" {
			return DetectQuirks(gameDir, exePath, appID, protonPath)
		}

		for _, cp := range CuratedPresets {
			if strings.EqualFold(cp.Name, cfg.PresetName) || (cp.UmuID != "" && strings.EqualFold(cp.UmuID, cfg.UmuID)) {
				p := cp
				return &p
			}
		}
		// Return synthetic preset representing the saved active config
		return &Preset{
			Name:          cfg.PresetName,
			MatchedSource: "Saved Configuration (.proton-config.toml)",
			SummaryNotes:  "Active quirks preset saved in game configuration.",
			UmuID:         cfg.UmuID,
			EnvVars:       cfg.EnvVars,
			ExtraArgs:     cfg.ExtraArgs,
			WaitProcesses: cfg.WaitProcesses,
			DisplayFile:   cfg.DisplayFile,
			Profiles:      cfg.Profiles,
		}
	}

	return DetectQuirks(gameDir, exePath, appID, protonPath)
}
