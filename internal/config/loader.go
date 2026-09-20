package config

import (
	"os"
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
		if err := toml.Unmarshal(data, cfg); err == nil {
			return cfg, nil
		}
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
	if topo, err := hardware.DetectCPUTopology(); err == nil && topo.IsHybrid {
		cfg.UsePCores = true
		cfg.PCoresMask = topo.PCoresMask
	}

	if gpu, err := hardware.DetectGPU(); err == nil {
		cfg.UsePrimeRun = gpu.HasPrimeRun
		if gpu.PreferredOutput != "" {
			cfg.GamescopeOutput = gpu.PreferredOutput
		}
	}

	return cfg, nil
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
