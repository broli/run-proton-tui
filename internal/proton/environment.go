package proton

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/hardware"
)

// EnvOptions contains settings required to construct the execution environment.
type EnvOptions struct {
	PrefixDir      string
	GameDir        string
	AppID          string
	UseXalia       bool
	EnableLogging  bool
	ExtraOverrides string // Additional DLL overrides string
}

// BuildEnvironment generates the complete environment variable map for Proton execution.
// It incorporates critical bugfixes for Steam emulators, Wayland swapchains,
// and the Proton Python runner logging requirement (SteamGameId).
func BuildEnvironment(opts EnvOptions) map[string]string {
	home, _ := os.UserHomeDir()
	env := make(map[string]string)

	// Copy base process environment
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	appID := opts.AppID
	if appID == "" || appID == "0" {
		appID = "0"
	}

	// 1. Core Steam / Proton compatibility variables
	env["STEAM_COMPAT_DATA_PATH"] = opts.PrefixDir
	env["STEAM_COMPAT_CLIENT_INSTALL_PATH"] = filepath.Join(home, ".local/share/Steam")
	env["STEAM_COMPAT_APP_ID"] = appID
	// CRITICAL FIX: Proton's internal runner requires SteamGameId to create steam-$SteamGameId.log
	env["SteamGameId"] = appID
	env["UMU_ID"] = "umu-default"
	env["UMU_USE_STEAM"] = "0" // Prevents delegating to steam.exe wrapper which exits prematurely

	// 2. Suppress Steam Vulkan implicit layers for non-Steam titles
	env["DISABLE_VK_LAYER_VALVE_steam_overlay_1"] = "1"
	env["DISABLE_VK_LAYER_VALVE_steam_fossilize_1"] = "1"

	// 3. Steam Emulators (Goldberg, CODEX, Rune, FLT, ALI213):
	// Disable Wine's built-in lsteamclient to ensure game-provided emulator DLLs are called
	overrides := "lsteamclient=d"
	if opts.ExtraOverrides != "" {
		overrides = fmt.Sprintf("%s;%s", overrides, opts.ExtraOverrides)
	}
	if existing, ok := env["WINEDLLOVERRIDES"]; ok && existing != "" {
		if !strings.Contains(existing, "lsteamclient") {
			overrides = fmt.Sprintf("%s;%s", overrides, existing)
		} else {
			overrides = existing
		}
	}
	env["WINEDLLOVERRIDES"] = overrides

	// 4. Accessibility & UI Automation control
	if opts.UseXalia {
		env["PROTON_USE_XALIA"] = "1"
	} else {
		env["PROTON_USE_XALIA"] = "0"
	}

	// 5. Kernel NT fast synchronization:
	// Automatically use /dev/ntsync if present on host, else fallback to Fsync/Esync
	if hardware.HasNTSync() {
		env["PROTON_USE_NTSYNC"] = "1"
		delete(env, "PROTON_NO_NTSYNC")
	} else {
		env["PROTON_NO_NTSYNC"] = "1"
		delete(env, "PROTON_USE_NTSYNC")
	}

	// 6. Modern Linux & SteamOS compatibility flags
	env["STEAMOS"] = "1"
	env["STEAMDECK"] = "1"

	// 7. NVIDIA NVAPI, DLSS & NGX
	env["PROTON_ENABLE_NVAPI"] = "1"
	env["DXVK_ENABLE_NVAPI"] = "1"
	env["PROTON_ENABLE_NGX_UPDATER"] = "1"

	// 8. Memory Management: Prevent VRAM thrashing on 8GB GPUs
	if _, ok := env["VKD3D_CONFIG"]; !ok {
		env["VKD3D_CONFIG"] = "no_upload_hvv"
	}

	// 9. Debug & Session Logging
	if opts.EnableLogging {
		logsDir := filepath.Join(opts.GameDir, ".logs")
		_ = os.MkdirAll(logsDir, 0755)

		env["PROTON_LOG"] = "1"
		env["PROTON_LOG_DIR"] = logsDir
		env["WINEDEBUG"] = "+err,+warn"
		env["DXVK_LOG_LEVEL"] = "info"
		env["DXVK_NVAPI_LOG_LEVEL"] = "info"
		env["VKD3D_DEBUG"] = "warn"
		env["VKD3D_SHADER_DEBUG"] = "fixme"
		env["GST_DEBUG"] = "1"
	} else {
		env["WINEDEBUG"] = "-all"
		env["GST_DEBUG"] = "0"
		delete(env, "PROTON_LOG")
		delete(env, "PROTON_LOG_DIR")
	}

	return env
}

// EnvSlice converts the environment map into a []string slice suitable for exec.Cmd.Env.
func EnvSlice(envMap map[string]string) []string {
	slice := make([]string, 0, len(envMap))
	for k, v := range envMap {
		slice = append(slice, fmt.Sprintf("%s=%s", k, v))
	}
	return slice
}
