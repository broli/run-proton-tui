---
description: RCA discipline gates and methodology for diagnosing Linux gaming issues and configuring Proton/Wine settings with rpt
globs: ["**/*.toml", "hooks/**", "cmd/**", "internal/**", "docs/**"]
---

# Game Troubleshooting & Configuration Discipline Protocol

When assisting users with diagnosing game crashes, configuring `.proton-config.toml`, or writing lifecycle hooks (`hooks/pre_launch.sh`, `hooks/post_exit.sh`), all AI agents operating in this workspace MUST follow this 4-gate Root Cause Analysis (RCA) protocol.

## 🛑 The "Combinatorial Rabbit Hole" Failure Mode
Never guess multiple environment variables, inject speculative DLL overrides, edit registry values, and add engine flags simultaneously ($A + B + C + D$). Stacking unverified workarounds causes prefix corruption, invalidates test feedback, and obscures the actual root cause.

---

## ⚙️ The 3-Tier Settings Hierarchy
1. **Level 0: Clean Zero Baseline** (`rpt --clean-zero --now`): Pure upstream Proton with zero extra wrappers (`use_gamescope = false`, `use_prime_run = false`, `use_pcores = false`, `manage_power = false`, `dll_overrides = {}`, `extra_args = []`).
2. **Level 1: Safe Standards & Hardware-Aware Defaults**: Sane defaults derived dynamically at runtime (Gamescope on Wayland, prime-run on NVIDIA hybrid laptops, P-core pinning on Intel hybrid CPUs, `gamescope_output = "auto"`, `refresh = 0`).
3. **Level 2: Curated Quirks & Profiles**: Per-game workarounds (e.g. 2D launcher isolation, `/dev/shm` RAM caches, anti-cheat stubs).

---

## 🛡️ The 4 Mandatory RCA Discipline Gates

### 1. The "Clean Zero" Baseline Gate
* **Rule**: Always verify baseline behavior with pure upstream defaults and zero extra flags before attempting deep diagnostics, custom patches, or registry tweaks.
* **Execution**:
  - Test the game with pure baseline using `rpt --clean-zero --now` or a minimal `.proton-config.toml` (`target_exe` and standard `proton_path`).
  - If unexpected behavior occurs, verify if the issue reproduces on a clean prefix (`rpt --clean --clean-zero --now`).
  - Only introduce custom environment variables, performance flags, or DLL overrides if the pure baseline fails.

### 2. Single-Variable Isolation (Strict Delta Testing)
* **Rule**: Never stack multiple workarounds or launch arguments ($A + B + C + D$) at the same time.
* **Execution**:
  - Formulate a single hypothesis and test exactly **one variable at a time**.
  - If Hypothesis $A$ does not resolve the issue or alters the symptom, **revert variable $A$ completely** before testing Hypothesis $B$.
  - Do not retain failed workarounds in the configuration "just in case."

### 3. Verify Exception Causality
* **Rule**: Never assume a logged Wine exception, `fixme:` stub, or warning caused a process exit.
* **Execution**:
  - Wine routinely outputs hundreds of benign `fixme:` messages and handled first-chance exceptions during normal execution.
  - Correlate exact timestamps between `.logs/runner.log` and process termination down to the second.
  - Check thread IDs: verify if the thread that logged the error is the thread that actually exited.
  - Check `coredumpctl list` and `journalctl -b 0 -e` to verify if the OS kernel, compositor (`kwin_wayland`), or OOM killer terminated the process.

### 4. Compositor Sandboxing Over Engine Flags
* **Rule**: When dealing with fractional scaling, resolution mismatches, aspect ratio stretching, or mouse cursor trapping under Wayland, prefer external compositor isolation (Gamescope) over forcing internal engine render flags (e.g., `-vulkan`, `-force-d3d11`, `-screen-width`).
* **Execution**:
  - Engine flags frequently bypass DXVK/VKD3D optimizations, destabilize swapchains, or trigger anti-cheat driver faults.
  - Configure Gamescope in `.proton-config.toml`:
    ```toml
    use_gamescope = true
    gamescope_width = 1920
    gamescope_height = 1080
    gamescope_output = "auto"
    ```
  - Let Gamescope provide a clean virtual X11 canvas while the Wayland compositor scales it.

---

## 📚 Escalation Path: When Simple Fixes Fail

If baseline verification and single-variable delta tuning do not resolve the issue:
1. Run `rpt --dump-spec` (or `rpt --helpdump`) to inspect system topology, connected display outputs, P-core masks, and runner versions.
2. Review `docs/wiki/AI-Agent-Integration.md` and `docs/wiki/Hardware-and-Wayland-Architecture.md`.
3. Check `docs/ENDFIELD_LESSONS_AND_QUIRKS_SPEC.md` for historical deep-dives (e.g., Wine `ntoskrnl.exe` anti-cheat kernel stubs, Unity IL2CPP RAM caching, 2D launcher isolation).
4. Review existing working recipes in `community-hooks/`.
