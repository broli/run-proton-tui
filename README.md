# rpt (Run Proton TUI) 🎮

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?logo=linux&logoColor=black)](https://kernel.org)
[![Wiki](https://img.shields.io/badge/Docs-GitHub%20Wiki-blueviolet)](https://github.com/broli/run-proton-tui/wiki)

**`rpt` makes running Windows games on Linux as simple as entering the game folder and typing `rpt`.**

Whether you are jumping into the main game, running a setup installer, updating through an official launcher, or applying a game patch, `rpt` handles the entire lifecycle automatically. No wrestling with complicated Wine prefixes, no memorizing dozens of launch flags, and no fear of accidentally deleting your save files. Just a clean, responsive terminal dashboard that gets your games running smoothly in seconds.

<p align="center">
  <img src="docs/screenshots/dashboard.png" alt="rpt Main Overview Dashboard" width="90%" />
</p>

---

## ⚡ The 10-Second Quickstart

You don't need to configure anything or select executables manually. Just open your terminal in any game folder:

```bash
cd ~/Games/MyGame
rpt
```

1. `rpt` automatically discovers your game.
2. It detects your best graphics card and tunes your CPU for maximum performance.
3. **Press `[Enter]` to play!**

Prefer to skip the menu and launch straight into the game?
```bash
rpt --now
```

---

## 🧭 The Three Core Philosophies of `rpt`

Look, let's be honest: at the end of the day, we all just want to launch our game, play, and be done with it. No headaches, no endless troubleshooting. And let's face it—unless you're an actual Proton developer, you're probably going to ask an AI how to fix whatever weird Wine error popped up anyway. So we planned for that! Here is the simple truth behind how `rpt` gets out of your way and makes gaming happen:

1. **🚀 Always-Work Universal Defaults**:  
   `rpt` includes all known, stable configurations for games to run out of the box. Safe prefix isolation, Gamescope sandboxing, CPU P-core pinning, auto-selection of your best GPU, and clean process supervision work automatically with zero setup required.

2. **⚡ Modular Power for Any Game (No Rocket Science Required)**:  
   If a fix or tweak only benefits a small number of games (such as anti-cheat kernel byte patches, memory-mapping IL2CPP caches to RAM, or game installer quirks), `rpt` keeps its core clean and delegates the fix to **modular lifecycle hooks**. This gives you the power to run *any* game, no matter how complicated the setup. Everything is fully documented and structured so an AI assistant (like ChatGPT, Claude, etc.) can write these complex scripts for you — meaning you never have to learn how to launch a rocket to Mars just to play your game.

3. **🎮 "You Do You" (Total Flexibility)**:  
   The application and configuration format are completely open and documented for hands-on gamers who love fine-tuning every single setting manually. But if you just want to play, you can simply tell an AI assistant to read the documentation, run `rpt`, inspect your files, and generate the configurations and hooks for you!

---

## 🌟 Why Gamers Love `rpt`

### 1. 🔍 Seamless Support for Games, Launchers, Installers & Patchers
Never worry about hunting down nested executable paths like `games/Arknights Endfield/Endfield.exe`. `rpt` recursively scans your game directory and intelligently distinguishes the **main 3D game** from **setup installers, game launchers, unpackers, and background patchers**. It automatically switches to lightweight settings for 2D utilities so they don't lag or crash, and supervises background updaters until they finish!

### 2. 🛡️ Never Lose Your Saves or Screenshots
When troubleshooting Wine games, standard online advice often tells you to "delete the prefix" — which accidentally wipes out your hard-earned save data and in-game camera photos!
* `rpt` keeps each game safely self-contained in its own `./proton-prefix/` folder.
* Whenever you clean or reset a prefix, `rpt` **automatically backs up your saves, documents, and photo gallery** to `~/Games/Backups/` first.

<p align="center">
  <img src="docs/screenshots/wine_prefix_management.png" alt="Wine Prefix Management & Safety" width="85%" />
</p>

### 3. 🚀 Smooth, Stutter-Free Performance Out of the Box
Modern gaming on Linux has great tools like Gamescope and hybrid CPU tuning, but setting them up manually can be intimidating. `rpt` handles it all automatically:
* **Picks Your Best GPU**: Runs games on your dedicated graphics card (NVIDIA RTX / AMD Radeon) while keeping your desktop responsive.
* **Eliminates Micro-Stutter**: On newer Intel CPUs (12th–14th Gen), games are automatically pinned to your fast Performance cores so background efficiency cores don't cause framerate dips.
* **Multi-Monitor Friendly**: Easily cycle which monitor your game appears on with a single keypress (`[m]`).

<p align="center">
  <img src="docs/screenshots/performance_sandbox.png" alt="Performance & Sandboxing Controls" width="85%" />
</p>

### 4. 🎯 Automated Game Fixes & Community Ratings
Wondering if a game runs well on Linux? You don't even need to open your web browser:
* **Live ProtonDB Ratings**: See real-time community tier badges (⭐ Platinum, 🥇 Gold, etc.) right on your screen.
* **Instant Game Recipes**: Automatically detects popular titles and applies tested community fixes (anti-cheat memory shims, audio workarounds, and video codecs) without you having to touch a config file.

<p align="center">
  <img src="docs/screenshots/quirks_presets.png" alt="Quirks & Compatibility Presets" width="85%" />
</p>

<p align="center">
  <img src="docs/screenshots/protondb_report.png" alt="Live ProtonDB Community Report" width="85%" />
</p>

### 5. 🧩 1-Click Modding (DLL Overrides)
Want to install mods like **ReShade**, **Unreal Engine Scripting (UE4SS)**, or **BepInEx**?
Forget opening `winecfg` and messing with the Windows registry. Just press `[o]` or choose **Prefix & Overrides** to enable popular mod loaders with a single click.

<p align="center">
  <img src="docs/screenshots/dll_overrides.png" alt="Interactive DLL Overrides" width="85%" />
</p>

### 6. 🩺 Friendly Crash Doctor
If a game fails to start or crashes within a few seconds, `rpt` doesn't just vanish. It captures the crash logs and opens the **Diagnostics Screen**, explaining in plain English what went wrong (e.g. missing executable permissions, conflicting processes, or DirectX errors) and how to resolve it.

### 7. 🎛️ Zero Setup Out of the Box, Total Control When You Need It
Every game is unique. While `rpt` works instantly with zero setup, it also lets you customize per-game behavior with a simple, human-readable `.proton-config.toml` right inside the game directory:
* **Custom Launch Arguments**: Add flags like `-dx11`, `-vulkan`, `-windowed`, or `-skipintro`.
* **Per-Executable Profiles**: Run the heavy 3D game with your dedicated GPU and Gamescope, while keeping the launcher light on battery.
* **Custom Symlinks & Directories**: Automatically link in-game photo cameras to your host `~/Pictures/` folder, or redirect save data to cloud storage.
* **Lifecycle Shell Hooks**: Run custom scripts before launch or after exit to mount archives, launch companion tools, or apply patches.

### 8. 🖥️ 1-Click Desktop & Application Shortcuts
Once your game is configured and tested, you can turn it into an icon on your desktop or application menu with a single keypress (`[s]`). `rpt` generates standard XDG `.desktop` launcher files with the game's icon and working directory configured, installing it straight to `~/.local/share/applications/` (so it appears in your system start menu/dock) or directly onto `~/Desktop/`.

---

## ⌨️ Simple Controls & Hotkeys

All main functions are organized into clean numbered categories, with convenient 1-key shortcuts:

| Key | Category / Action | What It Does |
|:---:|---|---|
| **`[Enter]`** | **Launch Game** | Run game with active configuration |
| **`[e / 2]`** | **Switch Executable** | Instant picker to switch between game binaries, launchers, or setup installers |
| **`[3]`** | **Proton & Game Setup** | Choose a Proton runner, game quirks, or view ProtonDB ratings |
| **`[4]`** | **Performance & Sandbox** | Toggle Gamescope, switch GPUs, or cycle display monitors (`[m]`) |
| **`[5]`** | **Prefix & Overrides** | Manage 1-click DLL mod overrides or safely reset the prefix |
| **`[6]`** | **Logs & Diagnostics** | View recent session/crash logs, AI helper files, and inspect lifecycle hooks |
| **`[s]`** | **Create Shortcut** | Generate a 1-click `.desktop` launcher for your desktop or app menu |
| **`[?]`** | **In-Depth Guide** | Open the complete built-in manual anytime |
| **`[q]`** | **Quit** | Exit `rpt` |

---

## 📥 Installation

Compiling `rpt` takes less than 5 seconds because it is written in pure Go with zero runtime dependencies:

```bash
git clone https://github.com/broli/run-proton-tui.git
cd run-proton-tui
make install
```

This installs `rpt` to your `~/bin/` folder, sets up shell completions for your terminal, and installs the manual page (`man rpt`).

### 📦 Dependencies & Optional Enhancements

`rpt` is a self-contained, statically compiled Go binary with **zero required runtime dependencies**. It automatically discovers your installed Steam/GE Proton runners and executes games out of the box.

To unlock advanced sandboxing, performance tuning, and cloud save synchronization, you can optionally install:

| Package | Purpose | Why You Might Want It |
|---|---|---|
| **`gamescope`** | Micro-compositor sandboxing | Sandboxed virtual display, resolution forcing (1080p, 1440p, 4K), custom refresh rates (75Hz, 144Hz), and multi-monitor routing (`[m]`). |
| **`util-linux` (`taskset`)** | CPU thread pinning | Pins games to fast CPU P-Cores on hybrid architectures (Intel 12th–14th Gen) to eliminate micro-stutter. |
| **`mangohud`** | Performance overlay | Real-time FPS, frame timing, GPU/CPU temperature, and VRAM monitoring. |
| **`rclone`** | Cloud save synchronization | Automatic pre-launch pull and post-exit push to Google Drive, Dropbox, Nextcloud, or OneDrive via `[cloud_sync]`. |

---

## 📚 Deep Dive & Documentation

Looking for technical architecture, config schemas, or developer guides? Check out our dedicated documentation:

* [📖 **GitHub Wiki Home**](https://github.com/broli/run-proton-tui/wiki)
* [⚙️ **Configuration Reference (`.proton-config.toml`)**](https://github.com/broli/run-proton-tui/wiki/Configuration-Reference)
* [🖥️ **Hardware Topology & Wayland Architecture**](https://github.com/broli/run-proton-tui/wiki/Hardware-and-Wayland-Architecture)
* [🔄 **Lifecycle Shell Hooks & File Directives**](https://github.com/broli/run-proton-tui/wiki/Lifecycle-Hooks-and-Preservation)
* [🤖 **AI Agent Integration & `--dump-spec`**](https://github.com/broli/run-proton-tui/wiki/AI-Agent-Integration)
* [🎮 **Example Game Guide (Arknights: Endfield)**](https://github.com/broli/run-proton-tui/wiki/Example-Config-Arknights-Endfield)
* [🗺️ **Project Roadmap & Planned Features**](ROADMAP.md)

---

## 📄 License

MIT License. Copyright (c) 2026 Carlos Ferrabone.
