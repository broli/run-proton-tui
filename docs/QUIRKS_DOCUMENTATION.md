# run-proton-tui (rpt) Quirks & Optimizations Reference Archive

This document provides a comprehensive audit and technical catalog of all game-specific quirks, environment variables, launch arguments, and platform shims historically integrated into `run-proton-tui`.

In accordance with the **Clean Zero Baseline Gate**, `rpt` does **not** silently inject these flags without user consent. Instead, they are presented with plain-English explanations and clear source attribution in the **Recommended Optimizations** menu or loaded via declarative `.rpt/quirks.toml` manifests.

---

## 1. Game-Specific Curated Presets

### Arknights: Endfield
* **Target Detection**: Directory contains `endfield` or `gryphlink`, or binary matches `Endfield.exe` / `Launcher.exe`.
* **Source**: Curated Game Quirks Engine / Gryphline Community Research.
* **Flags & Directives**:
  ```toml
  umu_id = "umu-endfield"
  extra_args = ["-vulkan"]
  display_file = "/tmp/gamescope-endfield-display"
  wait_processes = ["Updater.exe", "7zg.exe", "Patch.exe", "Games.exe", "ACE-Service64.exe"]

  [env_vars]
  WINE_CANONICAL_HOLE = "skip_volatile_check"
  PROTON_USE_XALIA = "0"
  VKD3D_CONFIG = "no_upload_hvv"
  ```
* **Technical Justifications**:
  1. `WINE_CANONICAL_HOLE="skip_volatile_check"`: Arknights: Endfield employs the Anti-Cheat Expert (ACE) kernel-level anti-cheat driver. On 64-bit Wine/Proton, ACE performs volatile memory scans across canonical 48-bit address space holes. Without this Wine Staging patch variable, Wine throws unhandled page fault exceptions, terminating the process immediately upon anti-cheat initialization.
  2. `extra_args = ["-vulkan"]`: The Windows Unity IL2CPP game binary defaults to DirectX 11/12 rendering. Passing `-vulkan` instructs the Unity engine to initialize its native Vulkan rendering backend directly, eliminating the overhead and potential shader stalls of D3D translation layers (DXVK/VKD3D).
  3. `VKD3D_CONFIG="no_upload_hvv"`: For scenarios where DirectX 12 is initialized, disabling Host-Visible Video Memory (HVV / ReBAR) staging allocations prevents severe VRAM thrashing and stutter on GPUs with 8GB of VRAM.
  4. `PROTON_USE_XALIA="0"`: Proton's Xalia accessibility automation bridge attempts to inspect UI hierarchies. On Chromium Embedded Framework (CEF) and Qt WebEngine launchers (like `Launcher.exe`), this causes memory corruption and double-free crashes.
  5. `umu_id="umu-endfield"`: Activates upstream ProtonFixes recipe redirecting 64 IL2CPP JIT mmap files into `/dev/shm` (RAM disk), preventing disk I/O bottlenecks.
  6. `wait_processes`: The Endfield launcher delegates downloading, patching, and 7-zip decompression to child executables (`Updater.exe`, `7zg.exe`, `Patch.exe`). Supervising these prevents `rpt` from tearing down the session while updates are being unpacked.

---

### Genshin Impact
* **Target Detection**: Directory or executable contains `genshin`.
* **Source**: Curated Game Quirks Engine / UMU ProtonFixes Database.
* **Flags & Directives**:
  ```toml
  umu_id = "umu-genshin"
  wait_processes = ["launcher.exe"]

  [env_vars]
  WINE_CANONICAL_HOLE = "skip_volatile_check"
  PROTON_USE_XALIA = "0"
  ```
* **Technical Justifications**:
  1. `WINE_CANONICAL_HOLE="skip_volatile_check"`: Suppresses Wine memory inspection crashes during anti-cheat security scans.
  2. `PROTON_USE_XALIA="0"`: Prevents crash loops inside the HoYoPlay web launcher.
  3. `wait_processes = ["launcher.exe"]`: Prevents premature session termination when the game launches from the HoYoPlay launcher.

---

### Goddess of Victory: NIKKE
* **Target Detection**: Directory or executable contains `nikke`.
* **Source**: Curated Game Quirks Engine / UMU Database.
* **Flags & Directives**:
  ```toml
  umu_id = "umu-nikke"

  [env_vars]
  WINE_CANONICAL_HOLE = "skip_volatile_check"
  PROTON_USE_XALIA = "0"
  ```
* **Technical Justifications**:
  1. `WINE_CANONICAL_HOLE="skip_volatile_check"`: Mitigates ACE/anti-cheat volatile memory scans.
  2. `PROTON_USE_XALIA="0"`: Stabilizes the web-based launcher window.

---

### Repack / Standalone Setup Installers
* **Target Detection**: Executable name is `setup.exe` or `installer.exe`.
* **Source**: Installer Detection Engine.
* **Flags & Directives**:
  ```toml
  [profiles."Setup.exe"]
  use_gamescope = false
  use_prime_run = false
  use_pcores = false
  ```
* **Technical Justifications**:
  1. `use_gamescope = false`: InnoSetup and NSIS installers often use non-standard window types or multi-window dialogs that behave poorly inside a fixed micro-compositor canvas. Running on the host desktop ensures dialog boxes are accessible.
  2. `use_prime_run = false`: Setup wizards do not require high-performance dedicated GPU 3D rendering.

---

## 2. Platform & Hardware Optimizations

### Gamescope Micro-Compositor
* **Settings**: `use_gamescope`, `gamescope_width`, `gamescope_height`, `gamescope_refresh`, `gamescope_scaling`, `gamescope_filter`.
* **Source**: Host Environment (Wayland Session).
* **Technical Justification**: On Wayland compositors (KDE Plasma, GNOME, Hyprland), fractional scaling and multi-monitor setups can lead to aspect ratio distortion, stutter, or mouse cursor escape. Gamescope creates an isolated virtual Xwayland canvas that handles scaling, fullscreening, and refresh rate management smoothly.
* **Clean Zero Default**: Disabled (`use_gamescope = false`) and geometry set to Auto (`0x0`). When geometry is blank (`0x0`), Gamescope naturally negotiates with the host display.

### Prime-Run (NVIDIA GPU Offload)
* **Setting**: `use_prime_run`.
* **Source**: Hardware Detection (NVIDIA Hybrid GPU / Optimus).
* **Technical Justification**: On laptops and desktops with hybrid graphics (Intel/AMD iGPU + NVIDIA dGPU), launching with `prime-run` sets `__NV_PRIME_RENDER_OFFLOAD=1` and `__GLX_VENDOR_LIBRARY_NAME=nvidia`, ensuring 3D workloads render on the dedicated graphics card.

### CPU Affinity / P-Core Pinning (`taskset`)
* **Settings**: `use_pcores`, `pcores_mask`.
* **Source**: Hardware Detection (Intel Alder Lake / Raptor Lake / Arrow Lake Hybrid Topologies).
* **Technical Justification**: Linux kernel schedulers can occasionally schedule latency-sensitive game worker threads onto low-power Efficiency (E) cores instead of Performance (P) cores, causing micro-stutters. Pinning threads to P-cores eliminates scheduling jitter.

---

## 3. Declarative Manifest Schema (`.rpt/quirks.toml`)

Users or title maintainers can define custom quirks declaratively without modifying `rpt` source code:

```toml
[[quirks]]
name = "My Custom Title"
match_exe = ["game.exe", "Game-Shipping.exe"]
summary_notes = "Enables anti-cheat hole skip and native Vulkan renderer"
extra_args = ["-vulkan"]

[quirks.env_vars]
WINE_CANONICAL_HOLE = "skip_volatile_check"
```

Candidate locations searched in order:
1. Game folder: `<game_dir>/.rpt/quirks.toml` or `<game_dir>/quirks.toml`
2. Game quirks dir: `<game_dir>/.rpt/quirks/*.toml`
3. User global: `~/.config/rpt/quirks/*.toml`
4. System-wide: `/usr/share/rpt/quirks/*.toml`
