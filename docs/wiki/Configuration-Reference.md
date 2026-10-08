# Configuration Reference (rpt.toml)

Every game directory managed by `rpt` stores its configuration in a clean, human-readable `rpt.toml` (or hidden `.rpt.toml`) file.

`rpt` organizes configurations into **named, self-contained profiles** under the `[profiles.<name>]` table, with an `active_profile` pointer at the root.

---

## 📑 Full TOML Schema Reference

```toml
# ==============================================================================
# rpt Configuration Specification (rpt.toml)
# ==============================================================================

# Currently active profile name
active_profile = "default"

# ------------------------------------------------------------------------------
# Profile: "default" (Self-Contained Configuration)
# ------------------------------------------------------------------------------
[profiles.default]

# --- 1. Core Execution Settings ---
target_exe = "Game.exe"               # Target Windows binary (relative to game dir)
proton_path = "/path/to/proton"       # Absolute path to proton runner executable
app_id = "0"                          # Steam AppID for UMU fixes and ProtonDB queries
extra_args = ["-vulkan"]              # Extra CLI flags passed directly to game executable

# --- 2. Hardware, Sandbox & Performance Controls ---
use_gamescope = true                  # Run inside Gamescope nested micro-compositor
gamescope_output = "auto"             # Target DRM display ("auto" lets host compositor decide)
gamescope_width = 0                   # Resolution canvas width (0 = auto/native)
gamescope_height = 0                  # Resolution canvas height (0 = auto/native)
gamescope_refresh = 0                 # Refresh rate in Hz (0 = native/untouched)
use_prime_run = true                  # Offload 3D rendering to NVIDIA dGPU via prime-run
use_pcores = true                     # Pin execution to Intel Performance cores (taskset)
pcores_mask = "0-11"                  # Thread affinity mask for Performance cores
manage_power = true                   # Automatically switch KDE power profile to performance
use_xalia = false                     # Enable Proton Xalia UI accessibility bridge
enable_logging = false                # Capture verbose Proton, DXVK, and VKD3D logs (.logs/)
opaque_backdrop = true                # TUI visual styling: opaque or transparent

# --- 3. Save Game & Screenshot Preservation ---
backup_dir = "~/winegames/saves"      # Preservation destination for prefix resets/cleanups

# --- 4. Multi-Stage Process Supervision ---
# When a launcher delegates to a patcher/updater, rpt supervises these binaries
wait_processes = ["Updater.exe", "7zg.exe", "Patch.exe"]

# --- 5. External Socket & Display Export ---
display_file = "/tmp/gamescope-game-display"

# --- 6. Custom Environment Variables ---
[profiles.default.env_vars]
WINE_CANONICAL_HOLE = "skip_volatile_check" # Prevents anti-cheat page fault crashes
VKD3D_CONFIG = "no_upload_hvv"              # Prevents 8GB VRAM thrashing
UMU_ID = "umu-custom"                       # Optional UMU game database identifier

# --- 7. WINEDLLOVERRIDES Mappings ---
[profiles.default.dll_overrides]
dxgi = "n,b"
d3d11 = "n,b"
version = "n,b"

# --- 8. Declarative Filesystem Directives ---
[profiles.default.filesystem]
ensure_dirs = [
  "{PREFIX}/pfx/drive_c/users/steamuser/AppData/LocalLow",
  "{PREFIX}/pfx/drive_c/users/steamuser/Pictures"
]

symlinks = [
  { source = "~/Pictures/MyGame", target = "{PREFIX}/pfx/drive_c/users/steamuser/Pictures/MyGame" }
]

extra_backup_paths = ["drive_c/GameData/Saves"]

# --- 9. Lifecycle Shell Hooks (Explicit Paths) ---
pre_launch_hook = "./hooks/pre_launch.sh"
post_exit_hook = "./hooks/post_exit.sh"
hook_dirs = ["./custom_scripts"]

# ------------------------------------------------------------------------------
# Profile: "launcher" (Independent 2D Web Launcher Profile)
# ------------------------------------------------------------------------------
[profiles.launcher]
target_exe = "launcher/Launcher.exe"
proton_path = "/path/to/proton"
app_id = "0"
use_gamescope = false                 # 2D web launcher runs natively on host desktop
gamescope_output = "auto"
use_prime_run = false                 # Runs on host iGPU (prevents Qt5/CEF NVAPI crash)
use_pcores = false
manage_power = false
use_xalia = false
enable_logging = false
opaque_backdrop = true
extra_args = []
```

---

## 🔍 Profile Management & Switching

Every profile in `rpt.toml` is **completely self-contained**. This means there are no mysterious hidden inheritances or conflicting partial overrides: what you see in the profile table is exactly what runs.

### Managing Profiles in the TUI:
- Press **`[P]`** on the dashboard to open the **Profile Manager**.
- View all profiles, see which one is active, inspect its target executable and runner.
- Press **`[Enter]`** to switch the active profile.
- Press **`[n]`** to create or clone a profile under a new name.
- Press **`[x]`** to delete a profile (cannot delete the last profile).

### Managing Profiles via CLI:
- Run with a specific profile:
  ```bash
  rpt --profile launcher
  ```
- Run headless directly with a chosen profile:
  ```bash
  rpt --profile game --now
  ```
- Inspect execution command without running:
  ```bash
  rpt --profile game --print-cmd
  ```

---

## 🎮 Real-World Case Studies & Examples
- [Arknights: Endfield Production Configuration](Example-Config-Arknights-Endfield): Complete setup with decoupled Gamescope, persistent screenshot symlinking, and Tencent Anti-Cheat Expert parameters.
