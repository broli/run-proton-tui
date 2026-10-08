package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/pelletier/go-toml/v2"
)

const (
	ConfigFileName    = "rpt.toml"
	DotConfigFileName = ".rpt.toml"
)

// LoadConfigFile loads the game configuration file from rpt.toml (or .rpt.toml).
// Returns os.ErrNotExist if neither file exists.
func LoadConfigFile(gameDir string) (*GameConfigFile, error) {
	candidates := []string{
		filepath.Join(gameDir, ConfigFileName),
		filepath.Join(gameDir, DotConfigFileName),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			var fileCfg GameConfigFile
			if err := toml.Unmarshal(data, &fileCfg); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", filepath.Base(path), err)
			}
			if fileCfg.Profiles == nil {
				fileCfg.Profiles = make(map[string]*GameConfig)
			}
			if len(fileCfg.Profiles) == 0 {
				fileCfg.Profiles["default"] = NewDefaultConfig()
				fileCfg.ActiveProfile = "default"
			}
			if fileCfg.ActiveProfile == "" || fileCfg.Profiles[fileCfg.ActiveProfile] == nil {
				for name := range fileCfg.Profiles {
					fileCfg.ActiveProfile = name
					break
				}
			}
			return &fileCfg, nil
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read %s: %w", filepath.Base(path), err)
		}
	}

	return nil, os.ErrNotExist
}

// SaveConfigFile serializes the GameConfigFile into rpt.toml in the target game directory.
func SaveConfigFile(gameDir string, fileCfg *GameConfigFile) error {
	if fileCfg == nil {
		return fmt.Errorf("cannot save nil configuration")
	}
	tomlPath := filepath.Join(gameDir, ConfigFileName)
	data, err := toml.Marshal(fileCfg)
	if err != nil {
		return err
	}
	return os.WriteFile(tomlPath, data, 0644)
}

// LoadConfig loads the active GameConfig from rpt.toml (or .rpt.toml).
// If no configuration file exists, it returns a new default configuration with hardware safe standards applied.
func LoadConfig(gameDir string) (*GameConfig, error) {
	fileCfg, err := LoadConfigFile(gameDir)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := NewDefaultConfig()
			ApplyHardwareSafeStandards(cfg)
			return cfg, nil
		}
		return nil, err
	}
	return fileCfg.GetActiveProfile(), nil
}

// SaveConfig saves or updates the active profile in rpt.toml.
func SaveConfig(gameDir string, cfg *GameConfig) error {
	fileCfg, err := LoadConfigFile(gameDir)
	if err != nil {
		fileCfg = NewConfigFileWithDefault()
	}
	if fileCfg.ActiveProfile == "" {
		fileCfg.ActiveProfile = "default"
	}
	fileCfg.Profiles[fileCfg.ActiveProfile] = cfg.Clone()
	return SaveConfigFile(gameDir, fileCfg)
}

// ApplyHardwareSafeStandards detects the host environment and hardware capabilities
// to apply Level 1 safe standard defaults (Gamescope on Wayland, prime-run on hybrid GPUs, P-core pinning on hybrid CPUs).
func ApplyHardwareSafeStandards(cfg *GameConfig) {
	if topo, err := hardware.DetectCPUTopology(); err == nil && topo.IsHybrid {
		cfg.UsePCores = true
		cfg.PCoresMask = topo.PCoresMask
	} else {
		cfg.UsePCores = false
		cfg.PCoresMask = ""
	}

	if gpu, err := hardware.DetectGPU(); err == nil {
		cfg.UsePrimeRun = gpu.HasPrimeRun
		cfg.GamescopeOutput = "auto"
	} else {
		cfg.UsePrimeRun = false
		cfg.GamescopeOutput = "auto"
	}

	// Default geometry to auto (0 = let gamescope match display)
	cfg.GamescopeWidth = 0
	cfg.GamescopeHeight = 0
	cfg.GamescopeRefresh = 0

	// Safe standard for Wayland: enable Gamescope if binary is present
	isWayland := os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland"
	if _, err := exec.LookPath("gamescope"); err == nil && isWayland {
		cfg.UseGamescope = true
	} else {
		cfg.UseGamescope = false
	}

	if _, err := exec.LookPath("powerprofilesctl"); err == nil {
		cfg.ManagePower = true
	} else {
		cfg.ManagePower = false
	}
}
