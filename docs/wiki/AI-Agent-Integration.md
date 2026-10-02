# AI Assistant Integration & Game Setup Helper

`rpt` is designed from the ground up to pair seamlessly with your favorite AI assistants (such as ChatGPT, Claude, Google Antigravity, and terminal coding assistants).

Instead of forcing you to learn complicated Linux environment variables, wine registry keys, or explain your laptop hardware to an AI, `rpt` provides a built-in helper flag:
```bash
rpt --dump-spec
# or
rpt --helpdump
```

---

## 🤖 Why `--dump-spec` Exists

When you ask an AI assistant:
> *"Help me get this Windows game running with `rpt` on Linux!"*

The AI typically doesn't know:
1. Which GPUs are installed on your machine (Intel, AMD, NVIDIA).
2. Which external monitors are currently plugged in.
3. Which CPU P-core threads exist on your processor.
4. Which Proton runners (GE-Proton, Proton Experimental) exist on your disk.
5. The exact supported configuration schema of `.proton-config.toml`.

By running `rpt --dump-spec`, you can copy and paste a single JSON document directly into ChatGPT or Claude. The AI immediately understands your complete system and can write the exact files you need without guessing.

---

## 📋 JSON Structure Overview (`rpt-spec-v2`)

Running `rpt --dump-spec` outputs:

```json
{
  "schema_version": "rpt-spec-v2",
  "rpt_version": "0.7.0",
  "title": "AI Assistant Game Setup Helper",
  "description": "System and launcher specification for AI assistants (like ChatGPT, Claude, etc.) to configure games and write launch hooks under run-proton-tui.",
  "ai_assistant_guidelines": [
    "1. PRIMARY GOAL: Configure the game's '.proton-config.toml' for graphics, runner path, Gamescope, CPU pinning, and DLL overrides.",
    "2. 'CLEAN ZERO' BASELINE GATE: Before diagnosing deep issues, writing custom patches, or adding workarounds, verify baseline behavior with pure upstream defaults and zero extra flags.",
    "3. SINGLE-VARIABLE ISOLATION (STRICT DELTA TESTING): Never stack workarounds or arguments (A + B + C + D). Test each variable in isolation; if variable A does not resolve the issue, revert it before testing B.",
    "4. VERIFY EXCEPTION CAUSALITY: Never assume a logged exception, Wine 'fixme:', or stub warning caused a process exit. Correlate timestamps, thread IDs, and parent process termination before attempting to patch or stub an error.",
    "5. COMPOSITOR SANDBOXING OVER ENGINE FLAGS: When dealing with fractional scaling or display resolution mismatches on Wayland, prefer external compositor isolation (Gamescope virtual canvas) over forcing internal engine render flags (e.g. -vulkan or -force-d3d11).",
    "6. '1-2 GAMES VS EVERY GAME' RULE: If a fix or tweak is game-specific (such as binary patching, Wine ntoskrnl.exe kernel patch checks, memory mapping /dev/shm, or anti-cheat workarounds), DO NOT modify the rpt binary. Write a lifecycle hook in 'hooks/pre_launch.sh'.",
    "7. IDEMPOTENCY: All hook scripts must be strictly idempotent. Always check if a patch, directory, or symlink is already applied before modifying anything.",
    "8. PREFIX FLUSHING & STALE LOCKS ARE AUTOMATIC: DO NOT add wineserver -k, wineserver -w, killall wine, or /tmp/.wine-* lock cleanup to hooks. rpt automatically flushes the Wine prefix, isolates and terminates lingering prefix processes, and cleans stale socket locks via non-blocking flock before launch, inside the process wrapper, and during teardown.",
    "9. ESCALATION & DEEP DOCUMENTATION: If simple baseline testing and single-variable delta tuning do not resolve the issue, consult the 'best_practices' dictionary in this dump and review 'documentation_links.troubleshooting_methodology' before attempting invasive patches or binary edits.",
    "10. COMMUNITY CONTRIBUTION: If your generated hook or configuration resolves an issue, remind the user to submit it to ProtonDB or as an rpt community hook to help fellow Linux gamers!"
  ],
  "hardware": {
    "has_prime_run": true,
    "has_nvidia": true,
    "has_intel": true,
    "has_amd": false,
    "has_ntsync": true,
    "connected_display_outputs": [
      "HDMI-A-1",
      "eDP-1"
    ],
    "preferred_output": "HDMI-A-1",
    "pcores_mask": "0-11"
  },
  "installed_runners": [
    {
      "name": "GE-Proton11-6",
      "path": "/home/user/.local/share/Steam/compatibilitytools.d/GE-Proton11-6/proton"
    }
  ],
  "config_fields": [
    {
      "key": "target_exe",
      "type": "string",
      "description": "Relative path to target Windows executable"
    },
    {
      "key": "use_gamescope",
      "type": "bool",
      "default": "true",
      "description": "Run inside Gamescope nested micro-compositor"
    }
  ],
  "lifecycle_hooks": {
    "supported_hooks": ["pre_launch.sh", "post_exit.sh"],
    "search_order": [
      "1. Explicitly configured path in .proton-config.toml (pre_launch_hook, post_exit_hook)",
      "2. Unpacked game directory root: $PWD/hooks/<type>.sh",
      "3. Unpacked game directory root: $PWD/.rpt/hooks/<type>.sh",
      "4. Unpacked game directory root: $PWD/<type>.sh",
      "5. Custom user directories in config (hook_dirs)",
      "6. User global config: $HOME/.config/rpt/hooks/<type>.sh",
      "7. User legacy config: $HOME/.rpt/hooks/<type>.sh"
    ],
    "exported_environment_variables": [
      "RPT_GAME_DIR: Absolute path to game directory",
      "RPT_PREFIX_DIR: Absolute path to proton-prefix directory",
      "RPT_TARGET_EXE: Target executable path",
      "RPT_PROTON_PATH: Path to Proton runner binary",
      "RPT_GAMESCOPE_DISPLAY: Nested display number (e.g. :1 or empty)",
      "RPT_HOOK_TYPE: pre_launch or post_exit"
    ],
    "managed_internally": [
      "Wine prefix flushing: graceful wineserver -k and -w executed automatically before launch and on process exit",
      "Orphan process termination: target prefix processes (/proc/$pid/environ) purged without affecting other Wine games",
      "Stale lock cleanup: non-blocking flock sweep of /tmp/.wine-<UID>/server-*/lock to safely purge dead socket locks",
      "Save & screenshot preservation: standard user profile directories backed up during prefix clean/reset"
    ]
  },
  "hook_recipes": {
    "kernel_patch_check": {
      "title": "Wine Kernel Patch Verification (ntoskrnl.exe)",
      "file_name": "hooks/pre_launch.sh",
      "description": "Idempotently checks if the selected Proton runner contains required kernel patches before launch."
    },
    "shm_ram_symlinks": {
      "title": "Unity IL2CPP Fast RAM Cache Redirection",
      "file_name": "hooks/pre_launch.sh",
      "description": "Redirects high-frequency JIT worker temp files to /dev/shm to prevent SSD micro-stutter."
    }
  },
  "best_practices": {
    "clean_zero_vs_safe_defaults": "Distinguish between Level 0 ('Clean Zero' Baseline) and Level 1 ('Safe Hardware-Aware Defaults'). When diagnosing issues (Gate 1), always begin with pure upstream defaults (use_gamescope=false, use_prime_run=false, use_pcores=false, manage_power=false). Once baseline functionality is verified, introduce safe hardware defaults (Gamescope for Wayland fractional scaling, prime-run for hybrid laptops, P-core pinning for hybrid CPUs) one variable at a time (Gate 2).",
    "clean_zero_baseline": "Before diagnosing complex crashes, applying registry tweaks, or writing launch hooks, always verify baseline behavior with pure upstream defaults and zero extra flags. If a clean prefix and default Proton runner launches the title, do not introduce speculative arguments.",
    "single_variable_delta_isolation": "Never combine multiple unverified arguments or workarounds at once (A + B + C + D). Test each variable strictly in isolation. If introducing variable A does not fix the issue or produces a different symptom, revert A completely before testing variable B. Stacking workarounds creates compounding failure states and obscures the true root cause.",
    "verify_exception_causality": "Never assume a logged Wine exception, 'fixme:' stub, or console warning caused a game exit. Wine routinely outputs benign warnings during normal execution. Always correlate exact log timestamps, thread IDs, and parent/child process exit signals before attempting to patch a binary or stub an error.",
    "compositor_sandboxing_wayland": "When encountering display resolution mismatches, aspect ratio distortion, or mouse cursor trapping issues under Wayland fractional scaling, prefer external compositor isolation via Gamescope (e.g. gamescope_width, gamescope_height, gamescope_output) over passing internal engine render flags (like -vulkan, -force-d3d11, -screen-width). Engine-level flag overrides often destabilize DXVK/VKD3D or trigger anti-cheat driver checks.",
    "wine_prefix_lifecycle_and_stale_locks": "Never add wineserver shutdown or prefix lock clearing to pre_launch.sh or post_exit.sh. rpt automatically performs prefix flushing (graceful wineserver -k/-w), per-prefix process isolation, and non-blocking flock stale lock cleanup (/tmp/.wine-<UID>) both immediately before launch and during post-exit teardown. Adding manual wineserver kills to hooks risks terminating concurrent Wine sessions and disrupts rpt's process supervisor."
  },
  "documentation_links": {
    "configuration_reference": "https://github.com/broli/run-proton-tui/wiki/Configuration-Reference",
    "lifecycle_hooks_guide": "https://github.com/broli/run-proton-tui/wiki/Lifecycle-Hooks-and-Preservation",
    "arknights_endfield_case": "https://github.com/broli/run-proton-tui/wiki/Example-Config-Arknights-Endfield",
    "hardware_architecture": "https://github.com/broli/run-proton-tui/wiki/Hardware-and-Wayland-Architecture",
    "troubleshooting_methodology": "https://github.com/broli/run-proton-tui/wiki/AI-Agent-Integration#ai-troubleshooting-protocol-the-4-discipline-gates"
  }
}
```

---

## ⚙️ The 3-Tier Settings Hierarchy: Baseline vs Safe Standards vs Quirks

To prevent speculative workaround compounding, `rpt` strictly delineates between three distinct operational tiers:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Tier 0: Clean Zero Baseline (Pure Upstream Defaults)                        │
│ • use_gamescope = false  • use_prime_run = false  • use_pcores = false     │
│ • manage_power = false   • dll_overrides = {}     • extra_args = []         │
│ Used to establish whether the game boots cleanly on vanilla Proton.         │
│ CLI shortcut: rpt --clean-zero --now                                        │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ Gate 1: Baseline established
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Tier 1: Safe Standards & Hardware-Aware Defaults (Dynamic Auto-Detection)   │
│ • use_gamescope = true (ONLY if Wayland + gamescope binary is in PATH)      │
│ • use_prime_run = true (ONLY if NVIDIA Optimus hybrid GPU + prime-run)      │
│ • use_pcores = true    (ONLY if Intel hybrid architecture with E-cores)     │
│ • gamescope_output = "auto" (connector dynamically routed; never hardcoded) │
│ • gamescope_refresh = 0 (native monitor refresh rate preserved)             │
│ Applied when initializing new games or resetting to safe defaults.         │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ Gate 2: Isolated Delta Testing
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Tier 2: Curated Quirks & Profiles (Game-Specific Overrides)                 │
│ • 2D Launcher Mode: Gamescope=false, PrimeRun=false for Setup/Launcher.exe  │
│ • Memory-mapped JIT cache redirects to /dev/shm                             │
│ • Wine ntoskrnl kernel stubs & DLL overrides                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🛡️ AI Troubleshooting Protocol: The 4 Discipline Gates (RCA Rules)

When diagnosing games running under Wine/Proton, AI assistants frequently fall into the **"combinatorial workaround trap"**: guessing multiple environment variables, injecting DLL overrides, editing registry values, and modifying engine flags all at once ($A + B + C + D$). When that fails, stacking even more workarounds causes prefix corruption, obscures root causes, and results in catastrophic debugging rabbit holes.

To avoid this, all AI assistants pair-programming with `rpt` MUST observe the following four RCA discipline gates:

### Gate 1: The "Clean Zero" Baseline Gate
* **Principle**: Before diagnosing deep issues, writing custom patches, injecting DLLs, or modifying game configurations, **always establish the baseline behavior with pure upstream defaults and zero extra flags**.
* **Anti-Pattern**: An AI immediately sets `WINEDLLOVERRIDES`, adds launch arguments, turns on DXVK async, and disables esync on the first attempt without knowing if the unmodified game runs on stock GE-Proton.
* **Disciplined Workflow**:
  1. Test launch with pure upstream defaults using the CLI flag:
     ```bash
     rpt --clean-zero --now
     ```
     Or configure a pristine baseline `.proton-config.toml`:
     ```toml
     target_exe = "Game.exe"
     proton_path = "/path/to/proton"
     use_gamescope = false
     use_prime_run = false
     use_pcores = false
     ```
  2. If an issue occurs, determine whether the issue reproduces on a clean prefix (`rpt --clean --clean-zero --now`).
  3. Only introduce overrides if the pure baseline fails.

### Gate 2: Single-Variable Isolation (Strict Delta Testing)
* **Principle**: **Never test stacked workarounds ($A + B + C + D$) simultaneously.** Test each variable strictly in isolation. If variable $A$ does not resolve the issue or alters the symptom unexpectedly, **revert variable $A$ completely** before testing variable $B$.
* **Anti-Pattern**: The game has a black screen. The AI adds `PROTON_USE_WINED3D=1`. It still black screens. The AI leaves that enabled and adds `WINEDLLOVERRIDES="dxgi=b"`. Then it adds `-force-d3d11`. Now the game crashes with an unhandled exception, and nobody knows which flag caused it.
* **Disciplined Workflow**:
  - Test Hypothesis $A$. Did it fix the exact issue?
    - **YES**: Keep $A$ and document the rationale.
    - **NO**: Revert $A$ back to clean baseline. Proceed to Hypothesis $B$.

### Gate 3: Verify Exception Causality
* **Principle**: **Never assume a logged Wine exception, `fixme:` stub, or warning caused a process exit.** Wine routinely outputs hundreds of benign warnings and handled first-chance exceptions during normal, healthy execution.
* **Anti-Pattern**: An AI sees `fixme:hid:handle_IRP_MN_QUERY_ID` or a benign `stub` warning 200 lines before process exit and spends hours writing a custom DLL stub or patch, when the real cause was an out-of-memory kill (OOM) or parent window close.
* **Disciplined Workflow**:
  1. Inspect the last lines of `.logs/runner.log` and system journal (`journalctl -b -0`).
  2. Correlate timestamps down to the second between the Wine log and process exit code.
  3. Verify thread IDs: check if the thread reporting the exception is the thread that terminated.
  4. Distinguish between first-chance (handled) exceptions and unhandled fatal crashes.

### Gate 4: Compositor Sandboxing Over Engine Flags
* **Principle**: When dealing with fractional scaling, resolution mismatches, multi-monitor stretching, or mouse cursor trapping on Wayland, **prefer external compositor isolation (Gamescope) over forcing internal engine render flags** (such as `-vulkan`, `-force-d3d11`, or `-screen-width`).
* **Anti-Pattern**: A Unity or Unreal game renders at 150% scaling or has black bars. The AI injects `-force-d3d11 -screen-fullscreen 0` into the engine arguments. This bypasses DXVK's Vulkan swapchain, breaks anti-cheat hooks, or introduces rendering corruption.
* **Disciplined Workflow**:
  - Keep internal game arguments stock.
  - Enable Gamescope in `.proton-config.toml`:
    ```toml
    use_gamescope = true
    gamescope_width = 1920
    gamescope_height = 1080
    gamescope_output = "auto"
    ```
  - Gamescope provides a pristine virtual X11 canvas for the Windows game while letting your Wayland compositor scale the surface cleanly.

---

## 📚 Escalation Path: When Simple Fixes Don't Work

If baseline verification, single-variable delta tuning, and standard `.proton-config.toml` options fail to launch the title, do not guess random registry hacks. Follow this structured escalation path:

1. **Check System Journal & Core Dumps**:
   Run `journalctl -b 0 -e` or `coredumpctl list` to check whether the process was aborted by `SIGSEGV`, `SIGABRT`, `kwin_wayland` GPU memory exhaustion (`ENOMEM`), or an anti-cheat driver fault.
2. **Review Specialized Architecture References**:
   - [Hardware and Wayland Architecture](Hardware-and-Wayland-Architecture): Read for hybrid laptop (iGPU + dGPU) offloading, display port routing, and Intel GEM memory leak prevention.
   - [Arknights: Endfield Engineering Retrospective](../../docs/ENDFIELD_LESSONS_AND_QUIRKS_SPEC.md): Read for deep dives into Wine `ntoskrnl.exe` anti-cheat kernel stubs, Unity IL2CPP `/dev/shm` memory maps, and 2D launcher isolation.
3. **Inspect Existing Community Recipes**:
   Check the `community-hooks/` directory in the repository for working pre-launch recipes (e.g. symlinks, RAM cache redirects, DLL verifications).
4. **Clean Prefix Isolation**:
   Run `rpt --clean` (`rpt -c`) to reset the prefix cleanly with automated save/screenshot preservation.

---

## 🍷 Automatic Wine Prefix Flushing & Stale Lock Management

### Why AI Agents Must NOT Add Wineserver or Prefix Flush Commands to Hooks
When creating `hooks/pre_launch.sh` or `hooks/post_exit.sh`, AI assistants should **never** include `wineserver -k`, `wineserver -w`, `killall wine`, or manual `/tmp/.wine-<UID>` lock deletion.

`rpt` provides built-in, 4-tier lifecycle supervision:
1. **Pre-Launch Preflight Flush**: Immediately before launching Gamescope or executing `pre_launch.sh`, `rpt` runs `prefix.Flush()` to gracefully terminate any dangling wineserver from previous sessions (`wineserver -k` and `-w`), terminates orphaned prefix processes isolated strictly by `WINEPREFIX`, and executes `prefix.CleanStaleLocks()` using non-blocking `flock(LOCK_EX | LOCK_NB)` to safely purge dead socket locks in `/tmp/.wine-<UID>/server-*/lock`.
2. **Wrapper Teardown**: The runner wrapper shell script automatically runs `wineserver -k` and `-w` on the target prefix immediately after the game process exits and child updater processes finish.
3. **Post-Exit Teardown**: Go `defer` executes another full `prefix.Flush()` and `prefix.CleanStaleLocks()` cleanup sequence upon session completion.
4. **Prefix Reset Guardrails**: Resetting a prefix (`rpt --clean` / `rpt -c`) automatically backs up saves and camera photos to `~/Games/Backups/` before cleanly flushing and rebuilding.

#### The Danger of Manual Hook Flushing
- **Breaking Prefix Isolation**: Running blanket `killall wine` or unisolated `wineserver -k` kills other running Proton/Wine games or background updaters on the user's system.
- **Disrupting Process Supervision**: Killing processes inside `post_exit.sh` can abort active cloud synchronization, updater loops, or save preservation tasks.

---

## 🛠️ How an AI Assistant Configures a Game

When you ask an AI assistant to set up a game:
1. Run `rpt --dump-spec`.
2. Paste the output into ChatGPT, Claude, or your assistant.
3. The AI follows the **4 Discipline Gates**: establishes a clean baseline, isolates variables, verifies causality, and avoids speculative engine hacks.
4. The AI reads your installed Proton runners, picks the best candidate, and creates `.proton-config.toml`.
5. If the game needs custom binary patching, anti-cheat kernel verification, or RAM caches, the AI writes an idempotent `./hooks/pre_launch.sh`.
6. Run `rpt` or `rpt --create-desktop` — your game is ready to play!
