# Configuration Reference (.proton-config.toml)

Every game directory managed by `rpt` stores its configuration in a clean, human-readable `.proton-config.toml` file.

---

## 📑 Full TOML Schema Reference

```toml
# ==============================================================================
# rpt Configuration Specification (.proton-config.toml)
# ==============================================================================

# ------------------------------------------------------------------------------
# 1. Core Execution Settings
# ------------------------------------------------------------------------------
target_exe = "Game.exe"               # Target Windows binary (relative to game dir)
proton_path = "/path/to/proton"       # Absolute path to proton runner executable
app_id = "0"                          # Steam AppID for UMU fixes and ProtonDB queries
extra_args = ["-vulkan"]              # Extra CLI flags passed directly to game executable

# ------------------------------------------------------------------------------
# 2. Hardware, Sandbox & Performance Controls
# ------------------------------------------------------------------------------
use_gamescope = true                  # Run inside Gamescope nested micro-compositor
gamescope_output = "HDMI-A-1"         # Target DRM display (or "auto" for external preferred)
gamescope_width = 1920                # Resolution canvas width
gamescope_height = 1080               # Resolution canvas height
gamescope_refresh = 75                # Refresh rate in Hz
use_prime_run = true                  # Offload 3D rendering to NVIDIA dGPU via prime-run
use_pcores = true                     # Pin execution to Intel Performance cores (taskset)
pcores_mask = "0-11"                  # Thread affinity mask for Performance cores
manage_power = true                   # Automatically switch KDE power profile to performance
use_xalia = false                     # Enable Proton Xalia UI accessibility bridge
enable_logging = false                # Capture verbose Proton, DXVK, and VKD3D logs (.logs/)
opaque_backdrop = true                # TUI visual styling: opaque or transparent

# ------------------------------------------------------------------------------
# 3. Save Game & Screenshot Preservation
# ------------------------------------------------------------------------------
# Destination directory for preserved files during prefix resets/cleanups.
# Standard Windows folders (Saved Games, Documents, AppData, Pictures) are preserved.
backup_dir = "~/winegames/saves"

# ------------------------------------------------------------------------------
# 4. Multi-Stage Process Supervision
# ------------------------------------------------------------------------------
# When a game launcher delegates to a patcher/updater and exits, rpt supervises
# these process patterns, preventing premature prefix teardown.
wait_processes = ["Updater.exe", "7zg.exe", "Patch.exe"]

# ------------------------------------------------------------------------------
# 5. External Socket & Display Export
# ------------------------------------------------------------------------------
# Path where Gamescope nested display number is written (e.g. for companion tools)
display_file = "/tmp/gamescope-game-display"

# ------------------------------------------------------------------------------
# 6. Custom Environment Variables
# ------------------------------------------------------------------------------
[env]
WINE_CANONICAL_HOLE = "skip_volatile_check" # Prevents anti-cheat page fault crashes
VKD3D_CONFIG = "no_upload_hvv"              # Prevents 8GB VRAM thrashing
UMU_ID = "umu-custom"                       # Optional UMU game database identifier

# ------------------------------------------------------------------------------
# 7. WINEDLLOVERRIDES Mappings
# ------------------------------------------------------------------------------
[dll_overrides]
dxgi = "n,b"
d3d11 = "n,b"
version = "n,b"

# ------------------------------------------------------------------------------
# 8. Declarative Filesystem Directives
# ------------------------------------------------------------------------------
[filesystem]
# Directories that must exist before launch
ensure_dirs = [
  "{PREFIX}/pfx/drive_c/users/steamuser/AppData/LocalLow",
  "{PREFIX}/pfx/drive_c/users/steamuser/Pictures"
]

# Persistent symlinks established prior to launch
# Supports {PREFIX}, {GAME_DIR}, and ~
symlinks = [
  { source = "~/Pictures/MyGame", target = "{PREFIX}/pfx/drive_c/users/steamuser/Pictures/MyGame" }
]

# Additional non-standard paths preserved during prefix resets
extra_backup_paths = ["drive_c/GameData/Saves"]

# ------------------------------------------------------------------------------
# 9. Lifecycle Shell Hooks (Explicit Paths)
# ------------------------------------------------------------------------------
# Optional: explicitly specify hook scripts (overrides auto-discovery)
pre_launch_hook = "./hooks/pre_launch.sh"
post_exit_hook = "./hooks/post_exit.sh"
hook_dirs = ["./custom_scripts"]

# ------------------------------------------------------------------------------
# 10. Per-Executable Profile Overrides
# ------------------------------------------------------------------------------
# When the active target matches a profile key (by full path or basename),
# these settings dynamically override the global config while sharing the prefix.
[profiles."Launcher.exe"]
target_exe = "launcher/Launcher.exe"
use_gamescope = false                 # 2D web launcher runs natively on host desktop
use_prime_run = false                 # Runs on host iGPU (prevents Qt5/CEF NVAPI crash)
use_pcores = false
extra_args = []
```

---

## 🔍 How Profile Matching Works

Games frequently package a 2D web updater/installer (`Launcher.exe`, `Setup.exe`) and the actual 3D game (`Game.exe`, `Shipping.exe`) in the same root folder or inside subdirectories.

When you switch executables in `rpt` (using `[3]` or passing `rpt Setup.exe`):
1. `rpt` checks if an exact match exists in `[profiles."..."]`.
2. If not found, it compares by **basename** (e.g., target `launcher/Launcher.exe` matches profile `Launcher.exe`).
3. If matched, the profile's flags (`use_gamescope`, `use_prime_run`, `extra_args`, etc.) automatically override the defaults.
4. When you switch back to `Game.exe`, the 3D performance settings are restored automatically!

---

## 🎮 Real-World Case Studies & Examples
- [Arknights: Endfield Production Configuration](Example-Config-Arknights-Endfield.md): Complete setup with decoupled Gamescope, persistent screenshot symlinking, and Tencent Anti-Cheat Expert parameters.

