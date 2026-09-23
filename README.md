# rpt (Run Proton TUI) 🎮

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?logo=linux&logoColor=black)](https://kernel.org)

**`rpt`** is a modular, fast, and hardware-aware Linux game launcher helper designed to run standalone Windows games, installers, and repacks under Valve's **Proton** and **Wine** with a rich Terminal User Interface (TUI).

Built in **Go** using the [Charmbracelet](https://charm.sh/) ecosystem (`bubbletea`, `lipgloss`, `bubbles`), `rpt` compiles to a single, self-contained, high-performance static binary with zero runtime dependencies. It runs seamlessly on Arch, CachyOS, Ubuntu, Fedora, Debian, and SteamOS/Bazzite.

<p align="center">
  <img src="docs/screenshots/dashboard.png" alt="rpt Main Overview Dashboard" width="95%" />
</p>

---

## ✨ Key Highlights & Features

### 1. Progressive Fallback & Smart UX
* **Instant Launch**: Run `rpt --now` or double-click to launch immediately with saved or hardware-tuned defaults.
* **Intelligent Binary Discovery**: Recursively scans game directories, distinguishes heavy 3D game engines from 2D setup utilities/launchers, and applies per-executable profiles.
* **Crash Interception**: If a game crashes or exits with an error, `rpt` captures the crash logs and opens the **Deep Diagnostics TUI** highlighting root causes and recommended remedies.

### 2. Game Quirks & Automated Presets
* **Curated & UMU Recipes**: Automatically detects known quirks (e.g. Arknights Endfield, Elden Ring, Unreal Engine 4/5 titles) and applies recommended environment flags, display exports, and process waitlists.
* **Steam Emulator Detection**: Native status detection for Goldberg, CODEX, Rune, Fairlight (FLT), and ALI213, with automatic DLC discovery (`DLC.txt`) and `lsteamclient=d` safety shims.

<p align="center">
  <img src="docs/screenshots/quirks_presets.png" alt="Quirks & Compatibility Presets" width="85%" />
</p>

### 3. Hardware Topology & Anti-Stutter Tuning
* **CPU P-Core Pinning**: Automatically detects hybrid architectures (e.g. Intel 12th/13th/14th Gen via `/sys/devices/cpu_core/cpus`) and pins games to Performance cores (`taskset -c 0-11`), eliminating micro-stutters caused by thread migration to Gracemont E-cores.
* **Gamescope Sandboxing (Decoupled)**: Runs Gamescope on the host iGPU compositor to prevent KDE Plasma / KWin Wayland memory exhaustion (`ENOMEM`), invoking `prime-run` **inside** the sandbox strictly for the 3D game.
* **On-the-Fly Display Output Selection**: Queries connected displays directly via unprivileged DRM sysfs (`/sys/class/drm/card*-*/status`) and routes output cleanly with `[m]` or via the sub-menu.
* **Kernel Fast Sync**: Dynamically detects `/dev/ntsync` (Linux 6.13+ / CachyOS) and configures `PROTON_USE_NTSYNC=1`, with graceful fallback to Fsync/Esync.

<p align="center">
  <img src="docs/screenshots/performance_sandbox.png" alt="Performance & Sandboxing Controls" width="85%" />
</p>

### 4. Safety Guardrails & Save Game Preservation
* **Prefix Isolation Handrail**: Detects if game files or executables are accidentally placed inside the prefix (`drive_c/`) and blocks prefix wipes to protect game data.
* **Automated Save & Screenshot Preservation**: Backs up user saves from `AppData`, `Saved Games`, `Documents`, and `Pictures` to `~/Games/Backups/<Game>/` before any prefix reset.
* **Declarative Filesystem Directives**: Supports custom symlinks (e.g. in-game photo camera folders) and prerequisite directories directly via `.proton-config.toml`.

<p align="center">
  <img src="docs/screenshots/wine_prefix_management.png" alt="Wine Prefix Management & Safety" width="85%" />
</p>

### 5. Interactive DLL Overrides Engine
* **TUI Overrides Manager**: Easily inspect and cycle Wine DLL override modes (`native`, `builtin`, `native,builtin`, `disabled`).
* **1-Click Popular Mod Presets**: Instant support for UE4SS, ReShade, BepInEx, SpecialK, and DDraw.
* **Custom DLL Input**: Add arbitrary DLL overrides with live prefix validation.

<p align="center">
  <img src="docs/screenshots/dll_overrides.png" alt="Interactive DLL Overrides" width="85%" />
</p>

### 6. Live ProtonDB Community Ratings & Steam Search
* **Asynchronous ProtonDB REST Client**: Real-time compatibility badges (⭐ Platinum, 🥇 Gold, etc.) with 1-click community recommendation viewing.
* **Fuzzy Title Resolution**: Strips repack delimiters (e.g. `[DODI Repack]`, `(GOG)`, `[FitGirl]`) to search the Steam Store API for the official game AppID automatically.

<p align="center">
  <img src="docs/screenshots/protondb_report.png" alt="Live ProtonDB Community Report" width="85%" />
</p>

---

## ⌨️ Hotkeys & Keybindings

| Key | Action |
|---|---|
| `[Enter]` or `[1]` | **Launch Game** with active configuration |
| `[2]` | **Select Proton Runner** (GE-Proton, CachyOS, Valve Experimental) |
| `[3]` | **Select Executable** (.exe picker with 2D Utility vs 3D Game tags) |
| `[g]` | **Toggle Gamescope** (1080p fixed canvas $\to$ primary monitor) |
| `[p]` | **Toggle CPU P-Core Pinning** (`taskset` to threads `0-11`) |
| `[v]` | **Toggle GPU Runner** (`prime-run` NVIDIA RTX vs Host iGPU) |
| `[o]` | **Configure DLL Overrides** (UE4SS, ReShade, BepInEx, etc.) |
| `[a]` | **Fetch ProtonDB Recommendations** |
| `[c]` | **Clean / Reset Wine Prefix** (with save game backup!) |
| `[h]` | **Pre-flight Permissions & Diagnostics** (auto-fixes `+x` bits) |
| `[l]` | **View Recent Session Logs & Crash Dumps** |
| `[?]` or `[F1]` | **Open Verbose Help & Manual** |
| `[q]` or `[Esc]` | **Quit** |

---

## 🚀 Installation & Setup

### Prerequisites
* Linux (Arch, CachyOS, Ubuntu, Fedora, Debian, SteamOS/Bazzite)
* Go 1.27+ (only required to compile from source)

### Building from Source
```bash
git clone https://github.com/broli/run-proton-tui.git
cd run-proton-tui
make
make install
```
This builds and installs `rpt` into `~/bin/rpt` along with `rptui` and `run-proton` symlinks.

---

## 📖 Recommended Workflow & CLI Usage

> **💡 Best Practice**: Always `cd` into your game's root directory and run `rpt` without arguments:
> ```bash
> cd ~/Games/ArknightsEndfield
> rpt
> ```
> `rpt` automatically discovers binaries (subfolders included), distinguishes 3D game engines from 2D utilities, loads quirks presets, and manages isolated `./proton-prefix/` sandboxes. You do **not** need to manually pass the `.exe` unless you want to target a specific installer or setup tool.

```bash
# 1. Recommended: Open interactive TUI in the game root directory
cd ~/Games/EldenRing
rpt

# 2. Launch immediately with saved or auto-detected settings (skip TUI)
rpt --now

# 3. Target a specific tool/unpacker directly
rpt Setup.exe

# 4. Safe prefix clean followed by instant launch
rpt --clean --now

# 5. Run pre-flight health diagnostics check
rpt --diagnostics
```

---

## 🛡️ Best Practice Directory Layout

```
~/Games/GameTitle/               <-- Game installation root (where all game files live)
├── Game.exe                     <-- Main 3D binary
├── steam_appid.txt              <-- AppID for ProtonDB & save emulation
├── .proton-config.toml          <-- Per-game settings generated by rpt
├── .logs/                       <-- Timestamped session logs
└── proton-prefix/               <-- Isolated Wine sandbox (safe to delete/reset)
    └── pfx/
        └── drive_c/users/...    <-- Save games (auto-backed up on clean)
```

---

## 📄 License

MIT License. Copyright (c) 2026 Carlos Ferrabone.
