package diagnostics

import (
	"encoding/json"

	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/proton"
)

// SpecDump defines the root structure of the machine-readable specification for AI assistants and automation tools.
type SpecDump struct {
	SchemaVersion         string                 `json:"schema_version"`
	RptVersion            string                 `json:"rpt_version"`
	Title                 string                 `json:"title"`
	Description           string                 `json:"description"`
	AIAssistantGuidelines []string               `json:"ai_assistant_guidelines"`
	Hardware              HardwareSpec           `json:"hardware"`
	InstalledRunners      []RunnerSpec           `json:"installed_runners"`
	ConfigFields          []ConfigFieldSpec      `json:"config_fields"`
	LifecycleHooks        HookSpec               `json:"lifecycle_hooks"`
	HookRecipes           map[string]HookRecipe  `json:"hook_recipes"`
	BestPractices         map[string]string      `json:"best_practices"`
	DocumentationLinks    map[string]string      `json:"documentation_links"`
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

// HookRecipe provides concrete, idempotent bash hook examples for AI assistants.
type HookRecipe struct {
	Title       string `json:"title"`
	FileName    string `json:"file_name"`
	Description string `json:"description"`
	Script      string `json:"script"`
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

	aiGuidelines := []string{
		"1. PRIMARY GOAL: Configure the game's '.proton-config.toml' for graphics, runner path, Gamescope, CPU pinning, and DLL overrides.",
		"2. '1-2 GAMES VS EVERY GAME' RULE: If a fix or tweak is game-specific (such as binary patching, Wine ntoskrnl.exe kernel patch checks, memory mapping /dev/shm, or anti-cheat workarounds), DO NOT modify the rpt binary. Write a lifecycle hook in 'hooks/pre_launch.sh'.",
		"3. IDEMPOTENCY: All hook scripts must be strictly idempotent. Always check if a patch, directory, or symlink is already applied before modifying anything.",
		"4. COMMUNITY CONTRIBUTION: If your generated hook or configuration resolves an issue, remind the user to submit it to ProtonDB or as an rpt community hook to help fellow Linux gamers!",
	}

	hookRecipes := map[string]HookRecipe{
		"kernel_patch_check": {
			Title:       "Wine Kernel Patch Verification (ntoskrnl.exe)",
			FileName:    "hooks/pre_launch.sh",
			Description: "Idempotently checks if the selected Proton runner contains required kernel patches before launch.",
			Script: `#!/bin/bash
# Verify Wine ntoskrnl.exe patch before launch
NTOSKRNL="$RPT_PROTON_PATH/../dist/lib64/wine/x86_64-windows/ntoskrnl.exe"
if [ ! -f "$NTOSKRNL" ]; then
    NTOSKRNL="$RPT_PROTON_PATH/../dist/lib/wine/x86_64-windows/ntoskrnl.exe"
fi

if [ -f "$NTOSKRNL" ]; then
    # Example check: inspect known patch bytes at offset
    echo "[HOOK:pre_launch] Checking ntoskrnl.exe kernel patch status..."
fi
exit 0
`,
		},
		"shm_ram_symlinks": {
			Title:       "Unity IL2CPP Fast RAM Cache Redirection",
			FileName:    "hooks/pre_launch.sh",
			Description: "Redirects high-frequency JIT worker temp files to /dev/shm to prevent SSD micro-stutter.",
			Script: `#!/bin/bash
# Redirect heavy JIT cache into RAM
SHM_DIR="/dev/shm/rpt-$USER/$(basename "$RPT_GAME_DIR")"
mkdir -p "$SHM_DIR"

CACHE_DIR="$RPT_PREFIX_DIR/pfx/drive_c/users/steamuser/AppData/LocalLow"
mkdir -p "$CACHE_DIR"

echo "[HOOK:pre_launch] Unity IL2CPP RAM buffer mapped to $SHM_DIR"
exit 0
`,
		},
	}

	bestPractices := map[string]string{
		"decoupled_gamescope": "On hybrid laptops, run Gamescope on the host iGPU (KWin compositor) and place prime-run INSIDE the sandbox on the game binary. Running prime-run gamescope exhausts Intel GEM memory (execbuf ENOMEM) and crashes KWin.",
		"2d_utility_isolation": "2D utilities, setup installers (Setup.exe), and web launchers (Qt5/CEF/Electron) MUST have NVAPI and Steam Deck flags stripped, and run on host iGPU without Gamescope to avoid glibc double-free memory corruption.",
		"screenshot_preservation": "Windows games save screenshots to C:\\users\\steamuser\\Pictures. Always preserve Pictures alongside Saved Games and AppData during prefix wipes.",
		"drm_display_routing": "External HDMI/DP ports are typically hardwired to the dGPU on hybrid laptops. Query /sys/class/drm/card*-*/status without sudo to target external displays directly and eliminate PCIe double-bounce stutter.",
	}

	docLinks := map[string]string{
		"configuration_reference": "https://github.com/broli/run-proton-tui/wiki/Configuration-Reference",
		"lifecycle_hooks_guide":   "https://github.com/broli/run-proton-tui/wiki/Lifecycle-Hooks-and-Preservation",
		"arknights_endfield_case": "https://github.com/broli/run-proton-tui/wiki/Example-Config-Arknights-Endfield",
		"hardware_architecture":   "https://github.com/broli/run-proton-tui/wiki/Hardware-and-Wayland-Architecture",
	}

	dump := SpecDump{
		SchemaVersion:         "rpt-spec-v2",
		RptVersion:            rptVersion,
		Title:                 "AI Assistant Game Setup Helper",
		Description:           "System and launcher specification for AI assistants (like ChatGPT, Claude, etc.) to configure games and write launch hooks under run-proton-tui.",
		AIAssistantGuidelines: aiGuidelines,
		Hardware:              hwSpec,
		InstalledRunners:      runnerSpecs,
		ConfigFields:          configFields,
		LifecycleHooks:        hookSpec,
		HookRecipes:           hookRecipes,
		BestPractices:         bestPractices,
		DocumentationLinks:    docLinks,
	}

	return json.MarshalIndent(dump, "", "  ")
}
