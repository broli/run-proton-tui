# Real-World Example: Arknights: Endfield

This page provides the battle-tested, complete `.proton-config.toml` specification for **Arknights: Endfield** on modern hybrid laptops (Intel Raptor Lake + NVIDIA RTX 40-series mobile, CachyOS / Arch Wayland).

It demonstrates how all advanced features of `rpt` come together:
- **Decoupled Gamescope architecture** (Host iGPU compositor + in-sandbox `prime-run`)
- **Anti-Cheat Expert (`ACE-BASE.sys`) kernel compatibility parameters**
- **In-game camera screenshot preservation** via host directory symlinking (`~/Pictures/ENDFIELD`)
- **Multi-stage updater supervision** (`Updater.exe`, `7zg.exe`, `Patch.exe`)
- **Dual 2D/3D executable profiles** to prevent Qt5/WebEngine memory corruption crashes

---

## 📄 Complete Configuration (`.proton-config.toml`)

Place this file into your Endfield root folder (e.g., `~/Games/ArknightsEndfield/.proton-config.toml`):

```toml
# ==============================================================================
# rpt Configuration: Arknights: Endfield
# Target: Endfield.exe (Unity IL2CPP + Tencent Anti-Cheat Expert)
# Runner: GE-Proton11-6
# ==============================================================================

target_exe = "Endfield.exe"
proton_path = "/home/testuser/.local/share/Steam/compatibilitytools.d/GE-Proton11-6/proton"
app_id = "0"
preset_name = "Arknights: Endfield"
umu_id = "umu-endfield"

# ------------------------------------------------------------------------------
# Hardware, Sandbox & Display Routing
# ------------------------------------------------------------------------------
use_gamescope = true
gamescope_output = "auto"
gamescope_width = 1920
gamescope_height = 1080
gamescope_refresh = 75
use_prime_run = true
use_pcores = true
pcores_mask = "0-11"
manage_power = true
use_xalia = false
enable_logging = false
opaque_backdrop = true

# ------------------------------------------------------------------------------
# Launch Directives & Process Supervision
# ------------------------------------------------------------------------------
extra_args = ["-vulkan"]
display_file = "/tmp/gamescope-endfield-display"
wait_processes = ["Updater.exe", "7zg.exe", "Patch.exe", "Games.exe", "ACE-Service64.exe"]

# ------------------------------------------------------------------------------
# Anti-Cheat & Runtime Environment Variables
# ------------------------------------------------------------------------------
[env_vars]
WINE_CANONICAL_HOLE = "skip_volatile_check"
PROTON_USE_XALIA = "0"
PROTON_ENABLE_NVAPI = "1"
DXVK_ENABLE_NVAPI = "1"
VKD3D_CONFIG = "no_upload_hvv"
UMU_USE_STEAM = "0"

# ------------------------------------------------------------------------------
# Declarative Filesystem & In-Game Camera Photo Storage
# ------------------------------------------------------------------------------
[filesystem]
ensure_dirs = [
  "{PREFIX}/pfx/drive_c/users/steamuser/AppData/LocalLow",
  "{PREFIX}/pfx/drive_c/users/steamuser/Pictures"
]

[[filesystem.symlinks]]
source = "~/Pictures/ENDFIELD"
target = "{PREFIX}/pfx/drive_c/users/steamuser/Pictures/ENDFIELD"

# ------------------------------------------------------------------------------
# 2D Launcher & Utility Profiles (Host iGPU Mode)
# ------------------------------------------------------------------------------
[profiles."PlatformProcess.exe"]
target_exe = "PlatformProcess.exe"
use_gamescope = false
use_prime_run = false
use_pcores = false

[profiles."Launcher.exe"]
target_exe = "Launcher.exe"
use_gamescope = false
use_prime_run = false
use_pcores = false

[profiles."Games.exe"]
target_exe = "Games.exe"
use_gamescope = false
use_prime_run = false
use_pcores = false
```

---

## 🔍 Technical Deep-Dive

### 1. In-Game Screenshot & Photo Storage (`[filesystem.symlinks]`)
Endfield accounts save player progress and inventory directly to server databases, so local game save files are not needed. However, the in-game photo camera saves high-resolution screenshots to:
```
C:\users\steamuser\Pictures\ENDFIELD\
```
The declarative symlink directive:
```toml
[[filesystem.symlinks]]
source = "~/Pictures/ENDFIELD"
target = "{PREFIX}/pfx/drive_c/users/steamuser/Pictures/ENDFIELD"
```
ensures that before launch, `rpt` verifies `~/Pictures/ENDFIELD` on the host Linux system and links it into the Wine prefix. Any screenshot taken in the game is written directly to your host's `~/Pictures/ENDFIELD/` folder, completely immune to Wine prefix cleans or resets.

### 2. Decoupled Gamescope Sandboxing
- **Gamescope on Host iGPU**: Gamescope executes as an unprivileged client on the Intel integrated GPU (KWin compositor), avoiding the Intel kernel GEM execution buffer exhaustion (`ENOMEM`) crash that kills the desktop when running `prime-run gamescope`.
- **In-Sandbox `prime-run`**: `prime-run` is placed strictly inside Gamescope, directing the dedicated NVIDIA RTX GPU to render the game frames directly to Gamescope's virtual display.
- **Physical Connector Routing**: `gamescope_output = "auto"` automatically detects connected external monitors via `/sys/class/drm/` (e.g. `HDMI-A-1`), eliminating PCIe double-bounce bus stutter.

### 3. Anti-Cheat Expert (`ACE-BASE.sys`) Parameters
- `WINE_CANONICAL_HOLE = "skip_volatile_check"`: Prevents Wine kernel drivers from throwing unexpected page fault exceptions during volatile memory scans.
- `UMU_USE_STEAM = "0"`: Prevents the UMU wrapper from attempting to delegate execution to Steam, preventing premature wrapper exits.
- `VKD3D_CONFIG = "no_upload_hvv"`: Prevents Vulkan Host-Visible Video memory staging allocation stalls.
- `-vulkan`: Instructs the Unity IL2CPP runtime engine to select the native Vulkan rendering backend.

### 4. 2D Utility Isolation
Games often package helper binaries alongside the main game (e.g. `PlatformProcess.exe`, `Launcher.exe`, `Games.exe`).
- When run on NVIDIA with NVAPI enabled, Qt5/Chromium WebEngine instances frequently encounter glibc memory corruption (`double free or corruption`).
- `rpt` automatically strips NVAPI and Gamescope for these profiles, allowing background downloaders or setup utilities to run smoothly on the host iGPU without interfering with the 3D game.
