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

### 5. Automated Proton Kernel Patching via Lifecycle Hook (`pre_launch.sh`)
Tencent Anti-Cheat Expert (`ACE-BASE.sys`) probes Wine's Windows NT kernel emulation (`ntoskrnl.exe`). When running on GE-Proton, it requires two specific function patches (`PsGetProcessExitProcessCalled` and `SeQueryInformationToken`).

Instead of manually re-patching every time Steam or ProtonUp-Qt updates your Proton build, you can create a simple `hooks/pre_launch.sh` script in your Endfield game folder:

```bash
#!/bin/bash
# Place in: ~/Games/ArknightsEndfield/hooks/pre_launch.sh
set -e

PROTON_DIR="$(dirname "$RPT_PROTON_PATH")"
NTOSKRNL="$PROTON_DIR/files/lib/wine/x86_64-windows/ntoskrnl.exe"
[ ! -f "$NTOSKRNL" ] && NTOSKRNL="$PROTON_DIR/lib/wine/x86_64-windows/ntoskrnl.exe"

if [ -f "$NTOSKRNL" ]; then
    python3 - << 'EOF' "$NTOSKRNL"
import sys, os, shutil
path = sys.argv[1]
with open(path, "r+b") as f:
    data = bytearray(f.read())

offset = 0x43a8
payload = bytes.fromhex("31c0c3" + "90" * 21)

if data[offset:offset+len(payload)] == payload:
    print("[rpt:hook] ✓ GE-Proton ntoskrnl.exe is already patched for Anti-Cheat Expert.")
else:
    print("[rpt:hook] ! Fresh or updated Proton detected; applying ntoskrnl.exe patch...")
    if not os.path.exists(path + ".orig"):
        shutil.copy2(path, path + ".orig")
    data[offset:offset+len(payload)] = payload
    sec_offset = 0x4f90
    sec_payload = bytes.fromhex("b8020000c0c3" + "90" * 18)
    data[sec_offset:sec_offset+len(sec_payload)] = sec_payload
    f.seek(0)
    f.write(data)
    f.truncate()
    print("[rpt:hook] ✓ Kernel patch applied successfully! Proceeding with launch.")
EOF
fi
```

Because `rpt` automatically discovers `hooks/pre_launch.sh` in your unpacked game folder, this check runs before Gamescope and Proton start, ensuring that newly installed or updated Proton versions are verified and patched on the fly!

