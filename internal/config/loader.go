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
	ConfigFileName       = ".proton-config.toml"
	LegacyConfigFileName = ".proton-config"
)

// LoadConfig loads the game configuration from .proton-config.toml.
// If not found, it checks for legacy .proton-config and migrates it.
// If neither exists, it generates a hardware-aware default.
func LoadConfig(gameDir string) (*GameConfig, error) {
	tomlPath := filepath.Join(gameDir, ConfigFileName)
	if data, err := os.ReadFile(tomlPath); err == nil {
		cfg := NewDefaultConfig()
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", ConfigFileName, err)
		}
		return cfg, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read %s: %w", ConfigFileName, err)
	}

	// Check for legacy Fish configuration
	legacyPath := filepath.Join(gameDir, LegacyConfigFileName)
	if _, err := os.Stat(legacyPath); err == nil {
		if legacyCfg, err := ParseLegacyFishConfig(legacyPath); err == nil {
			_ = SaveConfig(gameDir, legacyCfg)
			return legacyCfg, nil
		}
	}

	// Generate hardware-aware default
	cfg := NewDefaultConfig()
	ApplyHardwareSafeStandards(cfg)
	return cfg, nil
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
		if gpu.PreferredOutput != "" {
			cfg.GamescopeOutput = gpu.PreferredOutput
		} else {
			cfg.GamescopeOutput = "auto"
		}
	} else {
		cfg.UsePrimeRun = false
		cfg.GamescopeOutput = "auto"
	}

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

// SaveConfig serializes the GameConfig into .proton-config.toml in the target game directory.
func SaveConfig(gameDir string, cfg *GameConfig) error {
	tomlPath := filepath.Join(gameDir, ConfigFileName)
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(tomlPath, data, 0644)
}
