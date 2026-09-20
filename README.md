# rpt (Run Proton TUI) 🎮

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?logo=linux&logoColor=black)](https://kernel.org)

**`rpt`** is a modular, fast, and hardware-aware Linux game launcher helper designed to run standalone Windows games, installers, and repacks under Valve's **Proton** and **Wine** with a rich Terminal User Interface (TUI).

Built in **Go** using the [Charmbracelet](https://charm.sh/) ecosystem (`bubbletea`, `lipgloss`, `bubbles`), `rpt` compiles to a single, self-contained, high-performance static binary with zero runtime dependencies. It runs seamlessly on Arch, CachyOS, Ubuntu, Fedora, Debian, and SteamOS/Bazzite.

---

## ✨ Key Highlights & Features

1. **Progressive Fallback & Smart UX**:
   - **Instant Launch**: Run `rpt --now` or double-click to launch immediately with saved or hardware-tuned defaults.
   - **Crash Interception**: If a game crashes or exits with an error within 15 seconds, `rpt` captures the crash logs and opens the **Deep Diagnostics TUI** highlighting the failure cause and recommended remedies.
2. **First-Class Steam Emulator Support**:
   - Built-in detection and status reporting for **Goldberg Steam Emulator**, **CODEX**, **Rune**, **Fairlight (FLT)**, and **ALI213**.
   - Automatic discovery of `steam_appid.txt`, `configs.app.ini`, `steam_emu.ini`, and DLC definitions (`DLC.txt`).
   - Forces `WINEDLLOVERRIDES="lsteamclient=d"` by default to prevent Wine from delegating to native Linux Steam (eliminating store popups and crash loops).
3. **Hardware Topology & Anti-Stutter Tuning**:
   - **CPU P-Core Pinning**: Automatically detects hybrid architectures (e.g. Intel 12th/13th/14th Gen via `/sys/devices/cpu_core/cpus`) and pins games to Performance cores (`taskset -c 0-11`), eliminating stutter caused by thread migration to Gracemont E-cores.
   - **Gamescope Sandboxing (Decoupled)**: Runs Gamescope on the host iGPU compositor to prevent KDE Plasma / KWin Wayland memory exhaustion (`ENOMEM`), invoking `prime-run` **inside** the sandbox strictly for the 3D game.
   - **Kernel Fast Sync**: Dynamically detects `/dev/ntsync` (CachyOS / Linux 6.13+) and configures `PROTON_USE_NTSYNC=1`, with graceful fallback to Fsync/Esync.
   - **Power Profile Integration**: Temporarily boosts system power profile to `performance` via `powerprofilesctl` and automatically restores the original profile on exit.
4. **Safety Guardrails & Save Game Preservation**:
   - **Prefix Isolation Handrail**: Detects if game files or executables are accidentally placed inside the prefix (`drive_c/`) and blocks prefix wipes to protect game files.
   - **Automated Save Backups**: Backs up user saves from `AppData` and `Saved Games` to `~/Games/Backups/<Game>/` before any prefix reset.
   - **Socket & Lock Cleanup**: Cleans abandoned Wine socket directories in `/tmp/.wine-<UID>/` using non-blocking flock.
5. **Community Ratings & AI Diagnostics**:
   - Asynchronous **ProtonDB REST API** client displaying real-time compatibility badges (⭐ Platinum, 🥇 Gold, etc.) with 1-click community recommendation application.
   - Fallback Steam Store search API to automatically detect AppIDs from folder names.
   - Optional Gemini AI integration for instant crash log diagnosis.
6. **Verbose In-Depth Manual (`?` / `Shift + /`)**:
   - Comprehensive scrollable manual and troubleshooting encyclopedia built right into the TUI.

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

## 📖 CLI Usage

```bash
# Open interactive TUI in the game directory
cd ~/Games/EldenRing
rpt

# Launch immediately with saved or auto-detected settings (skip TUI)
rpt --now

# Specify target executable directly
rpt Game.exe

# Safe prefix clean followed by instant launch
rpt --clean --now

# Run pre-flight health diagnostics check
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
