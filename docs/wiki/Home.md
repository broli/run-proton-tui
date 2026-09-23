# Welcome to the rpt (run-proton-tui) Wiki

**`rpt`** is a generic, high-performance Linux gaming launcher helper and interactive Terminal User Interface (TUI) designed to run non-Steam and standalone Windows games under Valve's Proton, GE-Proton, or DW-Proton.

<p align="center">
  <img src="https://raw.githubusercontent.com/broli/run-proton-tui/main/docs/screenshots/dashboard.png" alt="rpt Main Overview Dashboard" width="95%" />
</p>

---

## 📖 Table of Contents
1. [Overview & Philosophy](#overview--philosophy)
2. [Interface Screenshots](#interface-screenshots)
3. [Quick Start & Recommended Workflow](#quick-start--recommended-workflow)
4. [Wiki Documentation Links](#wiki-documentation-links)
5. [Keyboard Shortcuts Quick Reference](#keyboard-shortcuts-quick-reference)

---

## Interface Screenshots

| Overview Dashboard | Performance & Sandboxing |
|:---:|:---:|
| <img src="https://raw.githubusercontent.com/broli/run-proton-tui/main/docs/screenshots/dashboard.png" width="100%" /> | <img src="https://raw.githubusercontent.com/broli/run-proton-tui/main/docs/screenshots/performance_sandbox.png" width="100%" /> |

| Quirks & Compatibility Presets | Live ProtonDB Community Ratings |
|:---:|:---:|
| <img src="https://raw.githubusercontent.com/broli/run-proton-tui/main/docs/screenshots/quirks_presets.png" width="100%" /> | <img src="https://raw.githubusercontent.com/broli/run-proton-tui/main/docs/screenshots/protondb_report.png" width="100%" /> |

| Wine Prefix Management | Interactive DLL Overrides |
|:---:|:---:|
| <img src="https://raw.githubusercontent.com/broli/run-proton-tui/main/docs/screenshots/wine_prefix_management.png" width="100%" /> | <img src="https://raw.githubusercontent.com/broli/run-proton-tui/main/docs/screenshots/dll_overrides.png" width="100%" /> |


---

## Overview & Philosophy

Modern Windows PC gaming on Linux has evolved rapidly, but hybrid hardware setups (e.g. Intel/AMD integrated graphics + NVIDIA discrete GPUs) and Wayland compositors (KDE Plasma 6, GNOME 46+, Hyprland) present unique architectural hazards:
- Gamescope crashes when forced onto NVIDIA dGPU on Wayland (`kwin_wayland intel execbuf ENOMEM`).
- 2D setup utilities and Qt5/CEF WebEngine launchers crash with glibc `double free or corruption` when probing NVAPI.
- Save files and screenshots (in `drive_c/users/steamuser/Pictures`) are often deleted by generic launchers during prefix resets.
- Multi-executable games (e.g. a 2D updater `Launcher.exe` vs a 3D game `Game.exe`) have conflicting hardware requirements while needing to share the exact same Wine prefix.

**`rpt` solves these problems with a 100% game-agnostic architecture:**
- **Zero hardcoded game logic**: All game-specific behaviors are declared in `.proton-config.toml` or handled via user shell hooks.
- **Decoupled Gamescope**: Gamescope runs natively on the host compositor's iGPU; `prime-run` runs strictly inside the sandbox on the Windows binary.
- **World-Readable Display Detection**: Queries Linux DRM sysfs directly without `sudo`, `kscreendoctor`, or external tools.
- **AI-Agent Ready**: Includes `--dump-spec` to output machine-readable runtime capabilities for autonomous AI pair programmers.

---

## Quick Start & Recommended Workflow

The recommended workflow is to **navigate to your game's directory and run `rpt`**:

```bash
# 1. Enter game folder
cd ~/Games/MyGame

# 2. Launch rpt
rpt
```

`rpt` will:
1. Scan the directory and discover executables.
2. Separate 3D game engines from 2D setup tools/launchers.
3. Automatically configure an isolated prefix (`./proton-prefix/`).
4. Detect connected displays and recommend peak performance settings.
5. Present the dashboard for instant launch (`[Enter]`).

### Command-Line Execution
```bash
# Instant launch using saved/auto-detected settings (bypasses TUI)
rpt --now

# Clean/reset prefix safely with automatic save & screenshot preservation
rpt --clean --now

# Switch to a specific setup/installer tool
rpt Setup.exe

# Output machine-readable JSON architecture specification for AI assistants
rpt --dump-spec
```

---

## Wiki Documentation Links

- **[Configuration Reference](Configuration-Reference)**: Complete documentation of `.proton-config.toml` fields, profiles, and environment overrides.
- **[Real-World Case Study: Arknights: Endfield](Example-Config-Arknights-Endfield)**: Complete production configuration with screenshot symlinking, Gamescope, and anti-cheat settings.
- **[Lifecycle Hooks & Preservation](Lifecycle-Hooks-and-Preservation)**: How to write `pre_launch.sh` scripts, hook search paths, and save/screenshot preservation.
- **[Hardware & Wayland Architecture](Hardware-and-Wayland-Architecture)**: Decoupled Gamescope sandboxing, DRM sysfs discovery, P-Core CPU pinning, and Wayland stability.
- **[AI Agent Integration](AI-Agent-Integration)**: Using `rpt --dump-spec` / `--helpdump` with LLM coding agents.

---

## Keyboard Shortcuts Quick Reference

| Hotkey | Action | Scope |
| :--- | :--- | :--- |
| **`Enter` / `1`** | **Launch Game** with active settings | Dashboard |
| **`2`** | **Proton & Game Setup** sub-menu (Runner, Target Exe, Quirks, ProtonDB) | Dashboard |
| **`3`** | **Performance & Sandbox** sub-menu (Gamescope, P-Cores, GPU, Displays) | Dashboard |
| **`4`** | **Prefix & Overrides** sub-menu (DLL overrides, Reset, Health) | Dashboard |
| **`5`** | **Logs & Diagnostics** sub-menu (Session logging, Log viewer, Pre-flight) | Dashboard |
| **`6`** | **Settings & Help** sub-menu (Backdrop styling, Manual, Quit) | Dashboard |
| **`g`** | Toggle Gamescope Sandboxing on/off | Instant / Global |
| **`p`** | Toggle CPU P-Core Thread Pinning (taskset) | Instant / Global |
| **`v`** | Toggle GPU Runner (`prime-run` NVIDIA offload) | Instant / Global |
| **`m`** | Cycle Target Display Connector (`HDMI-A-1`, `eDP-1`, `auto`) | Instant / Global |
| **`d`** | Inspect Active Quirks Preset & UMU Compatibility | Instant / Global |
| **`a`** | View live ProtonDB community tier report | Instant / Global |
| **`L`** | Toggle Proton/DXVK session logging (.logs/) | Instant / Global |
| **`c`** | Safe prefix clean with automatic save preservation | Instant / Global |
| **`?` / `F1`** | Open full interactive documentation manual | Instant / Global |
| **`q` / `Esc`** | Return from modal or quit rpt | Instant / Global |
