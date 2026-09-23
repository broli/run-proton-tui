package diagnostics

import (
	"encoding/json"

	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/proton"
)

// SpecDump defines the root structure of the machine-readable specification.
type SpecDump struct {
	SchemaVersion    string            `json:"schema_version"`
	RptVersion       string            `json:"rpt_version"`
	Description      string            `json:"description"`
	Hardware         HardwareSpec      `json:"hardware"`
	InstalledRunners []RunnerSpec      `json:"installed_runners"`
	ConfigFields     []ConfigFieldSpec `json:"config_fields"`
	LifecycleHooks   HookSpec          `json:"lifecycle_hooks"`
	BestPractices    map[string]string `json:"best_practices"`
}

type HardwareSpec struct {
	HasPrimeRun      bool     `json:"has_prime_run"`
	HasNvidia        bool     `json:"has_nvidia"`
	HasIntel         bool     `json:"has_intel"`
	HasAMD           bool     `json:"has_amd"`
	HasNTSync        bool     `json:"has_ntsync"`
	ConnectedOutputs []string `json:"connected_display_outputs"`
	PreferredOutput  string   `json:"preferred_output"`
	PCoresMask       string   `json:"pcores_mask"`
}

type RunnerSpec struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ConfigFieldSpec struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

type HookSpec struct {
	SearchOrder       []string `json:"search_order"`
	ExportedVariables []string `json:"exported_environment_variables"`
	SupportedHooks    []string `json:"supported_hooks"`
}

// GenerateSpecDump returns a complete machine-readable snapshot of rpt's architecture,
// configuration schema, hardware topology, and runner ecosystem.
func GenerateSpecDump(rptVersion string) ([]byte, error) {
	gpuInfo, _ := hardware.DetectGPU()
	topo, _ := hardware.DetectCPUTopology()
	connectedOutputs := hardware.GetConnectedDisplayOutputs()

	hwSpec := HardwareSpec{
		HasNTSync: hardware.HasNTSync(),
	}
	if gpuInfo != nil {
		hwSpec.HasPrimeRun = gpuInfo.HasPrimeRun
		hwSpec.HasNvidia = gpuInfo.HasNvidia
		hwSpec.HasIntel = gpuInfo.HasIntel
		hwSpec.HasAMD = gpuInfo.HasAMD
		hwSpec.PreferredOutput = gpuInfo.PreferredOutput
		hwSpec.ConnectedOutputs = connectedOutputs
	}
	if topo != nil {
		hwSpec.PCoresMask = topo.PCoresMask
	}

	runners, _ := proton.DiscoverRunners()
	runnerSpecs := make([]RunnerSpec, 0, len(runners))
	for _, r := range runners {
		runnerSpecs = append(runnerSpecs, RunnerSpec{
			Name: r.Name,
			Path: r.Path,
		})
	}

	configFields := []ConfigFieldSpec{
		{"target_exe", "string", "", "Relative path to target Windows executable"},
		{"proton_path", "string", "", "Absolute path to Proton runner executable"},
		{"app_id", "string", "0", "Steam AppID for compatibility and protonfixes lookup"},
		{"use_gamescope", "bool", "true", "Run inside Gamescope nested micro-compositor"},
		{"gamescope_output", "string", "auto", "Target DRM display output connector (e.g. HDMI-A-1, eDP-1, or auto)"},
		{"gamescope_width", "int", "1920", "Gamescope virtual canvas width"},
		{"gamescope_height", "int", "1080", "Gamescope virtual canvas height"},
		{"gamescope_refresh", "int", "75", "Gamescope target display refresh rate (Hz)"},
		{"use_prime_run", "bool", "true", "Execute game with prime-run NVIDIA GPU offloading"},
		{"use_pcores", "bool", "true", "Pin execution to Intel Performance cores via taskset"},
		{"pcores_mask", "string", "0-11", "CPU affinity mask for Performance core threads"},
		{"manage_power", "bool", "true", "Automatically switch KDE power profile to performance"},
		{"use_xalia", "bool", "false", "Enable Proton Xalia UI automation accessibility bridge"},
		{"enable_logging", "bool", "false", "Capture verbose Proton, DXVK, and VKD3D logs to .logs/"},
		{"backup_dir", "string", "~/Games/Backups", "Destination directory to preserve user saves/screenshots"},
		{"hook_dirs", "[]string", "[]", "Extra directories to search for lifecycle hooks"},
		{"wait_processes", "[]string", "[]", "Background/child processes to supervise before teardown"},
		{"display_file", "string", "", "File path where Gamescope DISPLAY number is written"},
		{"pre_launch_hook", "string", "", "Explicit script path executed before Proton launch"},
		{"post_exit_hook", "string", "", "Explicit script path executed after process exit"},
		{"env", "map[string]string", "{}", "Custom environment variables injected into Proton"},
		{"dll_overrides", "map[string]string", "{}", "Custom WINEDLLOVERRIDES mappings (e.g. dxgi = 'n,b')"},
		{"filesystem.ensure_dirs", "[]string", "[]", "Directories to ensure exist before launch"},
		{"filesystem.symlinks", "[]{source, target}", "[]", "Declarative persistent symlinks (supports {PREFIX}, {GAME_DIR}, ~)"},
		{"filesystem.extra_backup_paths", "[]string", "[]", "Extra relative prefix paths to preserve during prefix resets"},
		{"profiles.<exe_name>", "table", "", "Per-executable overrides for multi-binary game folders"},
	}

	hookSpec := HookSpec{
		SupportedHooks: []string{"pre_launch.sh", "post_exit.sh"},
		SearchOrder: []string{
			"1. Explicitly configured path in .proton-config.toml (pre_launch_hook, post_exit_hook)",
			"2. Unpacked game directory root: $PWD/hooks/<type>.sh",
			"3. Unpacked game directory root: $PWD/.rpt/hooks/<type>.sh",
			"4. Unpacked game directory root: $PWD/<type>.sh",
			"5. Custom user directories in config (hook_dirs)",
			"6. User global config: $HOME/.config/rpt/hooks/<type>.sh",
			"7. User legacy config: $HOME/.rpt/hooks/<type>.sh",
		},
		ExportedVariables: []string{
			"RPT_GAME_DIR: Absolute path to game directory",
			"RPT_PREFIX_DIR: Absolute path to proton-prefix directory",
			"RPT_TARGET_EXE: Target executable path",
			"RPT_PROTON_PATH: Path to Proton runner binary",
			"RPT_GAMESCOPE_DISPLAY: Nested display number (e.g. :1 or empty)",
			"RPT_HOOK_TYPE: pre_launch or post_exit",
		},
	}

	bestPractices := map[string]string{
		"decoupled_gamescope": "On hybrid laptops, run Gamescope on the host iGPU (KWin compositor) and place prime-run INSIDE the sandbox on the game binary. Running prime-run gamescope exhausts Intel GEM memory (execbuf ENOMEM) and crashes KWin.",
		"2d_utility_isolation": "2D utilities, setup installers (Setup.exe), and web launchers (Qt5/CEF/Electron) MUST have NVAPI and Steam Deck flags stripped, and run on host iGPU without Gamescope to avoid glibc double-free memory corruption.",
		"screenshot_preservation": "Windows games save screenshots to C:\\users\\steamuser\\Pictures. Always preserve Pictures alongside Saved Games and AppData during prefix wipes.",
		"drm_display_routing": "External HDMI/DP ports are typically hardwired to the dGPU on hybrid laptops. Query /sys/class/drm/card*-*/status without sudo to target external displays directly and eliminate PCIe double-bounce stutter.",
	}

	dump := SpecDump{
		SchemaVersion:    "rpt-spec-v1",
		RptVersion:       rptVersion,
		Description:      "Machine-readable runtime specification for AI agents configuring games under run-proton-tui.",
		Hardware:         hwSpec,
		InstalledRunners: runnerSpecs,
		ConfigFields:     configFields,
		LifecycleHooks:   hookSpec,
		BestPractices:    bestPractices,
	}

	return json.MarshalIndent(dump, "", "  ")
}
