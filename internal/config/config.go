package config

// GameConfig represents the persistent per-game configuration stored in .proton-config.toml.
type GameConfig struct {
	TargetExe        string   `toml:"target_exe"`
	ProtonPath       string   `toml:"proton_path"`
	AppID            string   `toml:"app_id"`
	UseGamescope     bool     `toml:"use_gamescope"`
	GamescopeOutput  string   `toml:"gamescope_output"`
	GamescopeWidth   int      `toml:"gamescope_width"`
	GamescopeHeight  int      `toml:"gamescope_height"`
	GamescopeRefresh int      `toml:"gamescope_refresh"`
	UsePCores        bool     `toml:"use_pcores"`
	PCoresMask       string   `toml:"pcores_mask"`
	UsePrimeRun      bool     `toml:"use_prime_run"`
	ManagePower      bool     `toml:"manage_power"`
	UseXalia         bool     `toml:"use_xalia"`
	EnableLogging    bool     `toml:"enable_logging"`
	OpaqueBackdrop   bool              `toml:"opaque_backdrop"`
	DLLOverrides     map[string]string `toml:"dll_overrides,omitempty"`
	ExtraArgs        []string          `toml:"extra_args"`
}

// NewDefaultConfig returns a sane default configuration.
func NewDefaultConfig() *GameConfig {
	return &GameConfig{
		TargetExe:        "",
		ProtonPath:       "",
		AppID:            "0",
		UseGamescope:     true,
		GamescopeOutput:  "HDMI-A-1",
		GamescopeWidth:   1920,
		GamescopeHeight:  1080,
		GamescopeRefresh: 75,
		UsePCores:        true,
		PCoresMask:       "0-11",
		UsePrimeRun:      true,
		ManagePower:      true,
		UseXalia:         false,
		EnableLogging:    false,
		OpaqueBackdrop:   true,
		DLLOverrides:     make(map[string]string),
		ExtraArgs:        make([]string, 0),
	}
}
