package config

import (
	"path/filepath"
	"strings"
)

// SymlinkDirective defines a declarative persistent or setup symlink.
type SymlinkDirective struct {
	Source string `toml:"source"`
	Target string `toml:"target"`
}

// FilesystemConfig defines declarative filesystem operations before launch.
type FilesystemConfig struct {
	EnsureDirs       []string           `toml:"ensure_dirs,omitempty"`
	Symlinks         []SymlinkDirective `toml:"symlinks,omitempty"`
	ExtraBackupPaths []string           `toml:"extra_backup_paths,omitempty"`
}

// CloudSyncConfig defines optional cloud save synchronization parameters.
type CloudSyncConfig struct {
	Backend      string `toml:"backend,omitempty"`       // "rclone", "syncthing", "rsync"
	RemotePath   string `toml:"remote_path,omitempty"`   // e.g. "gdrive:GameSaves/MyGame"
	LocalPath    string `toml:"local_path,omitempty"`    // e.g. "{PREFIX}/drive_c/users/steamuser/Saved Games"
	AutoSyncPre  bool   `toml:"auto_sync_pre,omitempty"` // pull saves before launch
	AutoSyncPost bool   `toml:"auto_sync_post,omitempty"`// push saves after exit
}

// ExecutableProfile allows fine-grained overrides for specific binaries within the same game directory.
// For example, Setup.exe (2D utility) vs Game.exe (3D Vulkan engine).
type ExecutableProfile struct {
	TargetExe        string            `toml:"target_exe,omitempty"`
	UseGamescope     *bool             `toml:"use_gamescope,omitempty"`
	GamescopeOutput     string            `toml:"gamescope_output,omitempty"`
	GamescopeWidth      int               `toml:"gamescope_width,omitempty"`
	GamescopeHeight     int               `toml:"gamescope_height,omitempty"`
	GamescopeRefresh    int               `toml:"gamescope_refresh,omitempty"`
	GamescopeScaling    string            `toml:"gamescope_scaling,omitempty"`
	GamescopeFilter     string            `toml:"gamescope_filter,omitempty"`
	GamescopeSharpness  int               `toml:"gamescope_sharpness,omitempty"`
	GamescopeWindowMode string            `toml:"gamescope_window_mode,omitempty"`
	GamescopeAdaptiveSync *bool           `toml:"gamescope_adaptive_sync,omitempty"`
	GamescopeMangoApp   *bool             `toml:"gamescope_mangoapp,omitempty"`
	GamescopeHDR        *bool             `toml:"gamescope_hdr,omitempty"`
	GamescopeFPSLimit   int               `toml:"gamescope_fps_limit,omitempty"`
	GamescopeExtraArgs  []string          `toml:"gamescope_extra_args,omitempty"`
	UsePCores           *bool             `toml:"use_pcores,omitempty"`
	PCoresMask          string            `toml:"pcores_mask,omitempty"`
	UsePrimeRun         *bool             `toml:"use_prime_run,omitempty"`
	ManagePower         *bool             `toml:"manage_power,omitempty"`
	UseXalia            *bool             `toml:"use_xalia,omitempty"`
	EnableLogging       *bool             `toml:"enable_logging,omitempty"`
	EnableLocalTelemetry *bool            `toml:"enable_local_telemetry,omitempty"`
	OpaqueBackdrop      *bool             `toml:"opaque_backdrop,omitempty"`
	ExtraArgs           []string          `toml:"extra_args,omitempty"`
	DLLOverrides        map[string]string `toml:"dll_overrides,omitempty"`
	EnvVars             map[string]string `toml:"env_vars,omitempty"`
	UmuID               string            `toml:"umu_id,omitempty"`
	WaitProcesses       []string          `toml:"wait_processes,omitempty"`
	DisplayFile         string            `toml:"display_file,omitempty"`
	PreLaunchHook       string            `toml:"pre_launch_hook,omitempty"`
	PostExitHook        string            `toml:"post_exit_hook,omitempty"`
	Filesystem          *FilesystemConfig `toml:"filesystem,omitempty"`
	CloudSync           *CloudSyncConfig  `toml:"cloud_sync,omitempty"`
}

// GameConfig represents the persistent per-game configuration stored in .proton-config.toml.
type GameConfig struct {
	TargetExe           string                        `toml:"target_exe"`
	ProtonPath          string                        `toml:"proton_path"`
	AppID               string                        `toml:"app_id"`
	UseGamescope        bool                          `toml:"use_gamescope"`
	GamescopeOutput     string                        `toml:"gamescope_output"`
	GamescopeWidth      int                           `toml:"gamescope_width"`
	GamescopeHeight     int                           `toml:"gamescope_height"`
	GamescopeRefresh    int                           `toml:"gamescope_refresh"`
	GamescopeScaling    string                        `toml:"gamescope_scaling,omitempty"`
	GamescopeFilter     string                        `toml:"gamescope_filter,omitempty"`
	GamescopeSharpness  int                           `toml:"gamescope_sharpness,omitempty"`
	GamescopeWindowMode string                        `toml:"gamescope_window_mode,omitempty"`
	GamescopeAdaptiveSync bool                        `toml:"gamescope_adaptive_sync,omitempty"`
	GamescopeMangoApp   bool                          `toml:"gamescope_mangoapp,omitempty"`
	GamescopeHDR        bool                          `toml:"gamescope_hdr,omitempty"`
	GamescopeFPSLimit   int                           `toml:"gamescope_fps_limit,omitempty"`
	GamescopeExtraArgs  []string                      `toml:"gamescope_extra_args,omitempty"`
	UsePCores           bool                          `toml:"use_pcores"`
	PCoresMask          string                        `toml:"pcores_mask"`
	UsePrimeRun         bool                          `toml:"use_prime_run"`
	ManagePower         bool                          `toml:"manage_power"`
	UseXalia            bool                          `toml:"use_xalia"`
	EnableLogging       bool                          `toml:"enable_logging"`
	EnableLocalTelemetry bool                         `toml:"enable_local_telemetry,omitempty"`
	OpaqueBackdrop      bool                          `toml:"opaque_backdrop"`
	DLLOverrides        map[string]string             `toml:"dll_overrides,omitempty"`
	ExtraArgs           []string                      `toml:"extra_args"`
	// Non-standard game & repack extensions
	PresetName          string                        `toml:"preset_name,omitempty"`
	UmuID               string                        `toml:"umu_id,omitempty"`
	EnvVars             map[string]string             `toml:"env_vars,omitempty"`
	WaitProcesses       []string                      `toml:"wait_processes,omitempty"`
	DisplayFile         string                        `toml:"display_file,omitempty"`
	BackupDir           string                        `toml:"backup_dir,omitempty"`
	HookDirs            []string                      `toml:"hook_dirs,omitempty"`
	PreLaunchHook       string                        `toml:"pre_launch_hook,omitempty"`
	PostExitHook        string                        `toml:"post_exit_hook,omitempty"`
	Filesystem          FilesystemConfig              `toml:"filesystem,omitempty"`
	CloudSync           *CloudSyncConfig              `toml:"cloud_sync,omitempty"`
	Profiles            map[string]*ExecutableProfile `toml:"profiles,omitempty"`
}

// NewDefaultConfig returns a sane default configuration.
func NewDefaultConfig() *GameConfig {
	return &GameConfig{
		TargetExe:           "",
		ProtonPath:          "",
		AppID:               "0",
		UseGamescope:        true,
		GamescopeOutput:     "HDMI-A-1",
		GamescopeWidth:      1920,
		GamescopeHeight:     1080,
		GamescopeRefresh:    0, // 0 = Native / Untouched
		GamescopeScaling:    "fit",
		GamescopeFilter:     "linear",
		GamescopeWindowMode: "fullscreen",
		GamescopeExtraArgs:  make([]string, 0),
		UsePCores:           true,
		PCoresMask:          "0-11",
		UsePrimeRun:         true,
		ManagePower:         true,
		UseXalia:            false,
		EnableLogging:       false,
		EnableLocalTelemetry: false,
		OpaqueBackdrop:      true,
		DLLOverrides:        make(map[string]string),
		ExtraArgs:           make([]string, 0),
		EnvVars:             make(map[string]string),
		WaitProcesses:       make([]string, 0),
		Profiles:            make(map[string]*ExecutableProfile),
	}
}

// Clone creates a deep copy of the GameConfig.
func (c *GameConfig) Clone() *GameConfig {
	if c == nil {
		return nil
	}
	cpy := *c
	if c.DLLOverrides != nil {
		cpy.DLLOverrides = make(map[string]string, len(c.DLLOverrides))
		for k, v := range c.DLLOverrides {
			cpy.DLLOverrides[k] = v
		}
	}
	if c.ExtraArgs != nil {
		cpy.ExtraArgs = make([]string, len(c.ExtraArgs))
		copy(cpy.ExtraArgs, c.ExtraArgs)
	}
	if c.EnvVars != nil {
		cpy.EnvVars = make(map[string]string, len(c.EnvVars))
		for k, v := range c.EnvVars {
			cpy.EnvVars[k] = v
		}
	}
	if c.WaitProcesses != nil {
		cpy.WaitProcesses = make([]string, len(c.WaitProcesses))
		copy(cpy.WaitProcesses, c.WaitProcesses)
	}
	if c.Profiles != nil {
		cpy.Profiles = make(map[string]*ExecutableProfile, len(c.Profiles))
		for k, v := range c.Profiles {
			if v != nil {
				profCpy := *v
				if v.ExtraArgs != nil {
					profCpy.ExtraArgs = make([]string, len(v.ExtraArgs))
					copy(profCpy.ExtraArgs, v.ExtraArgs)
				}
				if v.GamescopeExtraArgs != nil {
					profCpy.GamescopeExtraArgs = make([]string, len(v.GamescopeExtraArgs))
					copy(profCpy.GamescopeExtraArgs, v.GamescopeExtraArgs)
				}
				if v.DLLOverrides != nil {
					profCpy.DLLOverrides = make(map[string]string, len(v.DLLOverrides))
					for dk, dv := range v.DLLOverrides {
						profCpy.DLLOverrides[dk] = dv
					}
				}
				if v.EnvVars != nil {
					profCpy.EnvVars = make(map[string]string, len(v.EnvVars))
					for ek, ev := range v.EnvVars {
						profCpy.EnvVars[ek] = ev
					}
				}
				if v.WaitProcesses != nil {
					profCpy.WaitProcesses = make([]string, len(v.WaitProcesses))
					copy(profCpy.WaitProcesses, v.WaitProcesses)
				}
				if v.Filesystem != nil {
					fsCpy := *v.Filesystem
					if v.Filesystem.EnsureDirs != nil {
						fsCpy.EnsureDirs = make([]string, len(v.Filesystem.EnsureDirs))
						copy(fsCpy.EnsureDirs, v.Filesystem.EnsureDirs)
					}
					if v.Filesystem.Symlinks != nil {
						fsCpy.Symlinks = make([]SymlinkDirective, len(v.Filesystem.Symlinks))
						copy(fsCpy.Symlinks, v.Filesystem.Symlinks)
					}
					if v.Filesystem.ExtraBackupPaths != nil {
						fsCpy.ExtraBackupPaths = make([]string, len(v.Filesystem.ExtraBackupPaths))
						copy(fsCpy.ExtraBackupPaths, v.Filesystem.ExtraBackupPaths)
					}
					profCpy.Filesystem = &fsCpy
				}
				if v.CloudSync != nil {
					csCpy := *v.CloudSync
					profCpy.CloudSync = &csCpy
				}
				cpy.Profiles[k] = &profCpy
			}
		}
	}
	if c.CloudSync != nil {
		csCpy := *c.CloudSync
		cpy.CloudSync = &csCpy
	}
	if c.GamescopeExtraArgs != nil {
		cpy.GamescopeExtraArgs = make([]string, len(c.GamescopeExtraArgs))
		copy(cpy.GamescopeExtraArgs, c.GamescopeExtraArgs)
	}
	if c.HookDirs != nil {
		cpy.HookDirs = make([]string, len(c.HookDirs))
		copy(cpy.HookDirs, c.HookDirs)
	}
	if c.Filesystem.EnsureDirs != nil {
		cpy.Filesystem.EnsureDirs = make([]string, len(c.Filesystem.EnsureDirs))
		copy(cpy.Filesystem.EnsureDirs, c.Filesystem.EnsureDirs)
	}
	if c.Filesystem.Symlinks != nil {
		cpy.Filesystem.Symlinks = make([]SymlinkDirective, len(c.Filesystem.Symlinks))
		copy(cpy.Filesystem.Symlinks, c.Filesystem.Symlinks)
	}
	if c.Filesystem.ExtraBackupPaths != nil {
		cpy.Filesystem.ExtraBackupPaths = make([]string, len(c.Filesystem.ExtraBackupPaths))
		copy(cpy.Filesystem.ExtraBackupPaths, c.Filesystem.ExtraBackupPaths)
	}
	return &cpy
}

// GetEffectiveConfig returns an active copy of GameConfig with any per-executable
// profile overrides merged in for targetExe.
func (c *GameConfig) GetEffectiveConfig(targetExe string) *GameConfig {
	eff := c.Clone()
	if targetExe == "" {
		targetExe = c.TargetExe
	}
	if targetExe == "" || len(c.Profiles) == 0 {
		return eff
	}

	// Lookup profile by full path or basename (case-insensitive)
	var prof *ExecutableProfile
	targetBase := strings.ToLower(filepath.Base(targetExe))
	if p, ok := c.Profiles[targetExe]; ok {
		prof = p
	} else {
		for pKey, pVal := range c.Profiles {
			if strings.EqualFold(pKey, targetExe) ||
				strings.EqualFold(filepath.Base(pKey), targetBase) ||
				(pVal.TargetExe != "" && (strings.EqualFold(pVal.TargetExe, targetExe) || strings.EqualFold(filepath.Base(pVal.TargetExe), targetBase))) {
				prof = pVal
				break
			}
		}
	}

	if prof == nil {
		return eff
	}

	if prof.TargetExe != "" {
		eff.TargetExe = prof.TargetExe
	}
	if prof.UseGamescope != nil {
		eff.UseGamescope = *prof.UseGamescope
	}
	if prof.GamescopeOutput != "" {
		eff.GamescopeOutput = prof.GamescopeOutput
	}
	if prof.GamescopeWidth > 0 {
		eff.GamescopeWidth = prof.GamescopeWidth
	}
	if prof.GamescopeHeight > 0 {
		eff.GamescopeHeight = prof.GamescopeHeight
	}
	if prof.GamescopeRefresh > 0 {
		eff.GamescopeRefresh = prof.GamescopeRefresh
	}
	if prof.GamescopeScaling != "" {
		eff.GamescopeScaling = prof.GamescopeScaling
	}
	if prof.GamescopeFilter != "" {
		eff.GamescopeFilter = prof.GamescopeFilter
	}
	if prof.GamescopeSharpness > 0 {
		eff.GamescopeSharpness = prof.GamescopeSharpness
	}
	if prof.GamescopeWindowMode != "" {
		eff.GamescopeWindowMode = prof.GamescopeWindowMode
	}
	if prof.GamescopeAdaptiveSync != nil {
		eff.GamescopeAdaptiveSync = *prof.GamescopeAdaptiveSync
	}
	if prof.GamescopeMangoApp != nil {
		eff.GamescopeMangoApp = *prof.GamescopeMangoApp
	}
	if prof.GamescopeHDR != nil {
		eff.GamescopeHDR = *prof.GamescopeHDR
	}
	if prof.GamescopeFPSLimit > 0 {
		eff.GamescopeFPSLimit = prof.GamescopeFPSLimit
	}
	if len(prof.GamescopeExtraArgs) > 0 {
		eff.GamescopeExtraArgs = make([]string, len(prof.GamescopeExtraArgs))
		copy(eff.GamescopeExtraArgs, prof.GamescopeExtraArgs)
	}
	if prof.UsePCores != nil {
		eff.UsePCores = *prof.UsePCores
	}
	if prof.PCoresMask != "" {
		eff.PCoresMask = prof.PCoresMask
	}
	if prof.UsePrimeRun != nil {
		eff.UsePrimeRun = *prof.UsePrimeRun
	}
	if prof.ManagePower != nil {
		eff.ManagePower = *prof.ManagePower
	}
	if prof.UseXalia != nil {
		eff.UseXalia = *prof.UseXalia
	}
	if prof.EnableLogging != nil {
		eff.EnableLogging = *prof.EnableLogging
	}
	if prof.EnableLocalTelemetry != nil {
		eff.EnableLocalTelemetry = *prof.EnableLocalTelemetry
	}
	if prof.OpaqueBackdrop != nil {
		eff.OpaqueBackdrop = *prof.OpaqueBackdrop
	}
	if prof.UmuID != "" {
		eff.UmuID = prof.UmuID
	}
	if prof.DisplayFile != "" {
		eff.DisplayFile = prof.DisplayFile
	}
	if prof.PreLaunchHook != "" {
		eff.PreLaunchHook = prof.PreLaunchHook
	}
	if prof.PostExitHook != "" {
		eff.PostExitHook = prof.PostExitHook
	}
	if prof.Filesystem != nil {
		fsCpy := *prof.Filesystem
		if prof.Filesystem.EnsureDirs != nil {
			fsCpy.EnsureDirs = make([]string, len(prof.Filesystem.EnsureDirs))
			copy(fsCpy.EnsureDirs, prof.Filesystem.EnsureDirs)
		}
		if prof.Filesystem.Symlinks != nil {
			fsCpy.Symlinks = make([]SymlinkDirective, len(prof.Filesystem.Symlinks))
			copy(fsCpy.Symlinks, prof.Filesystem.Symlinks)
		}
		if prof.Filesystem.ExtraBackupPaths != nil {
			fsCpy.ExtraBackupPaths = make([]string, len(prof.Filesystem.ExtraBackupPaths))
			copy(fsCpy.ExtraBackupPaths, prof.Filesystem.ExtraBackupPaths)
		}
		eff.Filesystem = fsCpy
	}
	if prof.CloudSync != nil {
		csCpy := *prof.CloudSync
		eff.CloudSync = &csCpy
	}
	if len(prof.ExtraArgs) > 0 {
		eff.ExtraArgs = make([]string, len(prof.ExtraArgs))
		copy(eff.ExtraArgs, prof.ExtraArgs)
	}
	if len(prof.WaitProcesses) > 0 {
		eff.WaitProcesses = make([]string, len(prof.WaitProcesses))
		copy(eff.WaitProcesses, prof.WaitProcesses)
	}
	for k, v := range prof.DLLOverrides {
		if eff.DLLOverrides == nil {
			eff.DLLOverrides = make(map[string]string)
		}
		eff.DLLOverrides[k] = v
	}
	for k, v := range prof.EnvVars {
		if eff.EnvVars == nil {
			eff.EnvVars = make(map[string]string)
		}
		eff.EnvVars[k] = v
	}

	return eff
}
