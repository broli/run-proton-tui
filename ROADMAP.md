# 🗺️ rpt (run-proton-tui) Project Roadmap

This document outlines active development goals, planned architecture improvements, and known field issues targeted for future releases of `rpt`.

---

## 📌 High-Priority Backlog

### 1. 🩺 Premature Termination & Silent Exit Log Analysis Doctor

#### Background & Phenomenon
In testing with complex protected games (*Arknights: Endfield* on GE-Proton11-6 / CachyOS Wayland):
- The game process terminated unexpectedly after **~45 seconds** (and in today's session within **7.9 seconds**) of execution.
- The session returned with a positive or zero exit status (due to Gamescope exiting cleanly when its child window closes, or Wine reporting clean teardown after an anti-cheat heartbeat failure).
- Because `rpt` checked `result.ExitCode != 0` before declaring a crash, **no diagnostic analysis was performed** even though the game aborted after 7.9s:
  ```text
  [gamescope] [Info]  launch: Primary child shut down!
  (EE) failed to read Wayland events: Broken pipe

  [✓] Game session finished cleanly (duration: 7.9s).
  ```

#### Real-World Incident Log (September 23, 2026)
```text
~/Games/GRYPHLINK
❯ rpt

🚀 Launching games/Arknights Endfield/Endfield.exe with proton...
   Gamescope: Active (1920x1080 @ 75Hz -> HDMI-A-1)
   CPU Cores: P-Cores pinned (Threads 0-11)

[gamescope] [Info]  console: gamescope version 3.16.25 (gcc 16.2.1)
No CAP_SYS_NICE, falling back to regular-priority compute and threads.
Performance will be affected.
[gamescope] [Info]  scriptmgr: Loading scripts from: '/usr/share/gamescope/scripts'
[gamescope] [Info]  scriptmgr: Loading scripts from: '/usr/share/gamescope/scripts/00-gamescope'
[gamescope] [Info]  scriptmgr: Loading scripts from: '/usr/share/gamescope/scripts/00-gamescope/common'
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/common/inspect.lua' (id: 0)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/common/modegen.lua' (id: 1)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/common/util.lua' (id: 2)
[gamescope] [Info]  scriptmgr: Loading scripts from: '/usr/share/gamescope/scripts/00-gamescope/displays'
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/asus.rogally.lcd.lua' (id: 3)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/deckhd.steamdeck.deckhd-lcd.lua' (id: 4)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/gpd.win4.lcd.lua' (id: 5)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/lenovo.legiongo.lcd.lua' (id: 6)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/lenovo.legiongos.lcd.lua' (id: 7)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/onexplayer.f1.oled.lua' (id: 8)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/valve.steamdeck.lcd.lua' (id: 9)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/valve.steamdeck.oled.lua' (id: 10)
[gamescope] [Info]  scriptmgr: Running script file '/usr/share/gamescope/scripts/00-gamescope/displays/zotac.zone.oled.lua' (id: 11)
[gamescope] [Info]  scriptmgr: Loading scripts from: '/etc/gamescope/scripts'
[gamescope] [Warn]  scriptmgr: Directory '/etc/gamescope/scripts' does not exist
[gamescope] [Info]  scriptmgr: Loading scripts from: '/home/carlos/.config/gamescope/scripts'
[gamescope] [Warn]  scriptmgr: Directory '/home/carlos/.config/gamescope/scripts' does not exist
[gamescope] [Info]  vulkan: Intel device detected, forcing general queue family instead of compute-only queue
[gamescope] [Info]  vulkan: selecting physical device 'Intel(R) Graphics (RPL-P)': queue family 0 (general queue family 0)
[gamescope] [Info]  vulkan: physical device supports DRM format modifiers
[gamescope] [Info]  wlserver: [backend/headless/backend.c:60] Creating headless backend
[gamescope] [Info]  xdg_backend: Seat name: seat0
[gamescope] [Info]  xdg_backend: Initted Wayland backend
[gamescope] [Info]  vulkan: supported DRM formats for sampling usage:
[gamescope] [Info]  vulkan:   AR24 (0x34325241)
[gamescope] [Info]  vulkan:   XR24 (0x34325258)
[gamescope] [Info]  vulkan:   AB24 (0x34324241)
[gamescope] [Info]  vulkan:   XB24 (0x34324258)
[gamescope] [Info]  vulkan:   RG16 (0x36314752)
[gamescope] [Info]  vulkan:   NV12 (0x3231564E)
[gamescope] [Info]  vulkan:   AB4H (0x48344241)
[gamescope] [Info]  vulkan:   XB4H (0x48344258)
[gamescope] [Info]  vulkan:   AB48 (0x38344241)
[gamescope] [Info]  vulkan:   XB48 (0x38344258)
[gamescope] [Info]  vulkan:   AB30 (0x30334241)
[gamescope] [Info]  vulkan:   XB30 (0x30334258)
[gamescope] [Info]  vulkan:   AR30 (0x30335241)
[gamescope] [Info]  vulkan:   XR30 (0x30335258)
[gamescope] [Info]  wlserver: Using explicit sync when available
[gamescope] [Info]  wlserver: Running compositor on wayland display 'gamescope-0'
[gamescope] [Info]  wlserver: [backend/headless/backend.c:17] Starting headless backend
[gamescope] [Info]  wlserver: Successfully initialized libei for input emulation!
[gamescope] [Info]  wlserver: [xwayland/server.c:107] Starting Xwayland on :1
The XKEYBOARD keymap compiler (xkbcomp) reports:
> Warning:          Multiple symbols for level 1/group 1 on key <FK23>
>                   Using F23, ignoring XF86TouchpadOff
> Warning:          Symbol map for key <FK23> redefined
>                   Using last definition for conflicting fields
> Warning:          Symbol map for key <FK24> redefined
>                   Using last definition for conflicting fields
> Warning:          Could not resolve keysym XF86ElectronicPrivacyScreenOn
> Warning:          Could not resolve keysym XF86ElectronicPrivacyScreenOff
> Warning:          Could not resolve keysym XF86ActionOnSelection
> Warning:          Could not resolve keysym XF86ContextualInsert
> Warning:          Could not resolve keysym XF86ContextualQuery
Errors from xkbcomp are not fatal to the X server
[gamescope] [Info]  pipewire: stream state changed: connecting
[gamescope] [Info]  pipewire: stream state changed: paused
[gamescope] [Info]  pipewire: stream available on node ID: 144
[gamescope] [Info]  xdg_backend: uMaxContentLightLevel: 80
[gamescope] [Info]  xdg_backend: HDR INFO
[gamescope] [Info]  xdg_backend:   cv_hdr_enabled: false
[gamescope] [Info]  xdg_backend:   uMaxLum: 80, uRefLum: 80
[gamescope] [Info]  xdg_backend:   bExposeHDRSupport: false
[gamescope] [Info]  xdg_backend: uMaxContentLightLevel: 166
[gamescope] [Info]  xdg_backend: HDR INFO
[gamescope] [Info]  xdg_backend:   cv_hdr_enabled: false
[gamescope] [Info]  xdg_backend:   uMaxLum: 166, uRefLum: 166
[gamescope] [Info]  xdg_backend:   bExposeHDRSupport: false
[gamescope] [Info]  edid: Patching res 800x1280 -> 1920x1080
[gamescope] [Info]  vblank: Using timerfd.
[gamescope] [Info]  xdg_backend: Post-Initted Wayland backend
ProtonFixes[3946] INFO: Running protonfixes on "GE-Proton11-6", build at 2026-08-28 21:12:12+00:00.
ProtonFixes[3946] INFO: Running checks
ProtonFixes[3946] INFO: All checks successful
ProtonFixes[3946] INFO: Non-steam game Arknights: Endfield (umu-endfield)
ProtonFixes[3946] INFO: No store specified, using UMU database
ProtonFixes[3946] INFO: Using early stage global defaults for Arknights: Endfield (umu-endfield)
ProtonFixes[3946] INFO: Non-steam game Arknights: Endfield (umu-endfield)
ProtonFixes[3946] INFO: No store specified, using UMU database
ProtonFixes[3946] INFO: Using early stage global protonfix for Arknights: Endfield (umu-endfield)
ProtonFixes[3946] INFO: Running protonfixes on "GE-Proton11-6", build at 2026-08-28 21:12:12+00:00.
ProtonFixes[3946] INFO: Running checks
ProtonFixes[3946] INFO: All checks successful
ProtonFixes[3946] INFO: Non-steam game Arknights: Endfield (umu-endfield)
ProtonFixes[3946] INFO: No store specified, using UMU database
ProtonFixes[3946] INFO: Using main stage global defaults for Arknights: Endfield (umu-endfield)
ProtonFixes[3946] INFO: Non-steam game Arknights: Endfield (umu-endfield)
ProtonFixes[3946] INFO: No store specified, using UMU database
ProtonFixes[3946] INFO: Using main stage global protonfix for Arknights: Endfield (umu-endfield)
Proton: Executable is inside wine prefix, launching normally.
ntsync: up and running.
[Gamescope WSI] No application info given.
[Gamescope WSI] Executable name: explorer.exe
The XKEYBOARD keymap compiler (xkbcomp) reports:
> Warning:          Unsupported maximum keycode 709, clipping.
>                   X11 cannot support keycodes above 255.
> Warning:          Virtual modifier Hyper multiply defined
>                   Using 0, ignoring 0
> Warning:          Virtual modifier ScrollLock multiply defined
>                   Using 0, ignoring 0
Errors from xkbcomp are not fatal to the X server
[gamescope] [Info]  xdg_backend: Changed refresh to: 74.973hz
Unable to read VR Path Registry from C:\users\steamuser\AppData\Local\openvr\openvrpaths.vrpath
[Gamescope WSI] Application info:
  pApplicationName: Endfield.exe
  applicationVersion: 1
  pEngineName: DXVK
  engineVersion: 12587008
  apiVersion: 4206592
[Gamescope WSI] Executable name: Endfield.exe
ATTENTION: default value of option vk_wsi_force_swapchain_to_current_extent overridden by environment.
ATTENTION: default value of option vk_wsi_force_swapchain_to_current_extent overridden by environment.
[Gamescope WSI] No application info given.
[gamescope] [Info]  launch: Primary child shut down!
(EE) failed to read Wayland events: Broken pipe

[✓] Game session finished cleanly (duration: 7.9s).
```

#### Root Causes in Current Architecture
1. **Rigid Crash Detection Logic**:
   In [`internal/runner/runner.go`](internal/runner/runner.go):
   ```go
   if !result.AbortedByUser && (result.ExitCode != 0 || (duration < 15*time.Second && result.ExitCode != 0)) {
       result.CrashDetected = true
   }
   ```
   - If the exit code is masked as `0`, `result.CrashDetected` remains `false`.
   - The 15-second window is too short to catch anti-cheat driver heartbeat timeouts, IL2CPP JIT initialization aborts, or Qt5 NVAPI double-free crashes, which frequently occur between 30 and 60 seconds after launch.
2. **Gamescope Exit Code Masking**:
   When Gamescope wraps `.rpt_runner.sh`, Gamescope can return exit code `0` upon window destruction even if the underlying `proton waitforexitandrun` command failed or was terminated by a signal.
3. **Log Analysis Only on Error**:
   The diagnostic analyzer (`runner.AnalyzeSessionLog`) is currently bypassed completely if `CrashDetected` is false, leaving critical errors in `.logs/` uninspected.
4. **Missing Anti-Cheat Signatures**:
   [`internal/runner/analyzer.go`](internal/runner/analyzer.go) lacks pattern matchers for Tencent Anti-Cheat Expert (`ACE-BASE.sys`), `STATUS_WINE_STUB` (`0x80000100`), unpatched `ntoskrnl.exe` `ProbeForWrite` crashes, or memory-mapped JIT aborts.

#### Implementation Status: Completed in v0.5.0
- [x] **False-Positive Friendly Short Exit Interception**:
  - Prefer false positives: if any session terminates under **90 seconds** (or with non-zero exit code), assume the game stalled or aborted.
  - Track child exit codes explicitly via `.rpt_runner.sh` exit status file (`/tmp/rpt-child-exit-$PID`) so Gamescope cannot mask game failure codes.
- [x] **Proactive Log Scanner & Modular Signatures**:
  - Keep universal failure signatures directly in `rpt` (out of memory, page faults, Vulkan device lost, missing files).
  - Support modular external signatures dictionary (`signatures.toml`) for game-specific driver aborts and anti-cheat stubs (`STATUS_WINE_STUB`, `ace-base.sys`, `ProbeForWrite`) that can be updated independently like virus definitions.
  - Proactively scan and report: `[!] Found known issue: [Category]`.
- [x] **AI Assistant Diagnostic Package (`ask-ai-help.txt`)**:
  - Automatically compile a self-contained diagnostic file and output its **absolute path**:
    `📄 /home/user/Games/MyGame/.logs/ask-ai-help.txt`
  - Include game configuration, system hardware, log tail, and direct links to GitHub wiki documentation.
  - Friendly instructions for gamers with zero AI/LLM background:
    *"The AI assistant will do its best to diagnose the issue and suggest settings or launch scripts 😉"*
  - Instruct the AI assistant to provide fixes via `.proton-config.toml` or create lifecycle hooks (`hooks/pre_launch.sh`).
- [x] **Linux Boot-Style `[ OK ]` Launch Indicators**:
  - When launching games, output clean sequential status lines (`[ OK ] Target verified`, `[ OK ] Hook executed`, `[ OK ] CPU affinity set`) for immediate reassurance while slow games initialize.
- [x] **Manual Post-Session Diagnostics**:
  - Option in Category `[6] Logs & Diagnostics` to inspect logs and lifecycle hooks on demand.

---

### 2. 🪝 Interactive Lifecycle Hook Inspector & Built-in Pager

#### Background & Need
`rpt` features a cascading lifecycle hook engine that looks for `pre_launch.sh` and `post_exit.sh` across multiple locations (game directory root, `.rpt/hooks/`, `hooks/`, custom config dirs, and user global directories).

These scripts perform mission-critical setup for games:
- Memory-mapping JIT buffers into `/dev/shm`
- Applying runtime kernel binary patches to Wine's `ntoskrnl.exe`
- Setting up persistent screenshot symlinks
- Launching clipboard bridge daemons

Currently, users have no way inside the `rpt` TUI to verify which hook scripts were resolved, whether they have execute permissions, or what code they contain without switching to an external terminal or editor.

#### Implementation Status: Completed in v0.5.0
- [x] **Hook Discovery & Status Screen**:
  - Dedicated **Lifecycle Hooks** view (accessible via Category `[6] Logs & Diagnostics` or hotkey `[h]`).
  - Displays the resolved status of both phases:
    - `Pre-Launch Hook`: `Active` (Path) or `None detected`
    - `Post-Exit Hook`: `Active` (Path) or `None detected`
  - Shows cascade resolution priority breakdown:
    1. Config explicit override (`.proton-config.toml`)
    2. Local game folder (`./hooks/pre_launch.sh`, `./.rpt/hooks/pre_launch.sh`, `./pre_launch.sh`)
    3. Custom hook paths (`hook_dirs`)
    4. Global user configuration (`~/.config/rpt/hooks/pre_launch.sh`)
- [x] **Built-in Hook Pager with Syntax Highlighting**:
  - Integrated scrollable Bubbletea viewport with Chroma ANSI syntax highlighting for bash scripts (`dracula` / terminal colors).
  - Press `[1]` or `[2]` or `[Enter]` on any discovered hook to view its full script content directly within `rpt`.
  - Display metadata header and quick return hotkeys (`[Esc]`).
- [x] **Exported Environment Variables Reference**:
  - Interactive cheatsheet in the pager view showing all variables injected by `rpt` at runtime:
    - `$RPT_GAME_DIR`, `$RPT_PREFIX_DIR`, `$RPT_TARGET_EXE`, `$RPT_PROTON_PATH`, `$RPT_GAMESCOPE_DISPLAY`, `$RPT_HOOK_TYPE`, `$RPT_CHILD_PID`.
- [x] **CLI Inspection Option**:
  - Added `rpt --inspect-hooks` or `rpt --hooks` flag for terminal and AI assistant inspection without entering interactive TUI mode.

---

## 🔮 Medium-to-Long Term Vision

### 3. 🧩 Declarative Quirks / Plugin Engine (Completed in v0.5.0)
- Move game-specific quirks from hardcoded logic into declarative `.toml` quirks manifests.
- Discovery hierarchy:
  1. Local game override: `$GAME_DIR/.rpt/quirks.toml`
  2. User custom quirks: `~/.config/rpt/quirks/*.toml`
  3. System/bundled quirks: `/usr/share/rpt/quirks/*.toml` or embedded Go assets
  4. Offline UMU fallback: `umu-database.csv`
- Dynamic application of 2D vs 3D environment flags, wait processes, and filesystem directives.

### 4. 🛡️ Hook-Centric Patching & Verification ("1-2 Games vs Every Game" Rule) (Completed in v0.5.0)
- **Architectural Principle**: Binary patching and kernel patch verification belong strictly in **lifecycle hooks** (`hooks/pre_launch.sh`), **NOT** hardcoded in the `rpt` binary.
- Provide heavily commented, idempotent hook templates:
  - Checking `ntoskrnl.exe` hex offset before launch.
  - Safe backup before patching.
  - Applying patches without touching the core launcher.
- Handled by the user or an AI assistant reading `rpt` documentation.

### 5. ☁️ Cloud Save Sync Integration (Completed in v0.5.0)
- Declarative sync hooks for Rclone, Syncthing, or custom rsync backends.
- Automatically execute backup sync before launch and after exit hooks finish.
- Document optional dependencies (`optdepends=('rclone: cloud save synchronization')`) in Makefile and AUR PKGBUILD guidelines.

### 6. 🎮 Wayland Gamescope Passthrough Cleanliness
- Ensure clean controller and input passthrough through Gamescope without SDL event duplication.
- De-scope complex in-app controller mapping (gamers configure gamepads via Steam / system tools; AI assistants can configure launch environment flags).

### 7. 🖥️ 1-Click Desktop & Application Shortcut Generator (`.desktop`) (Completed in v0.5.0)
- Generate standard XDG `.desktop` launcher entries once a game is configured and verified.
- Target destinations:
  - System applications menu (`~/.local/share/applications/<game>.desktop` -> shows in KDE start menu, GNOME Dash, Rofi, and Steam)
  - Desktop (`~/Desktop/<game>.desktop`)
  - Local game directory (`./<game>.desktop`)
- Accessible via TUI hotkey `[s]` or CLI `rpt --create-desktop`.

### 8. 📂 Community Hooks Repository (Completed in v0.5.0)
- Host bundled, tested community hooks in the repository under `community-hooks/` (e.g. `community-hooks/arknights-endfield/pre_launch.sh`).
- Provide annotated templates for community contributions and AI assistant reference (`pre_launch.template.sh`, `post_exit.template.sh`).
