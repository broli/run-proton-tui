# AI Agent Integration & Machine-Readable Spec

`rpt` is designed from the ground up to pair seamlessly with modern AI coding assistants (such as Google Antigravity, Claude, GitHub Copilot, and terminal agents).

Instead of forcing users to manually learn complex Linux environment flags or explain hardware topology to their AI, `rpt` provides a native discovery flag:
```bash
rpt --dump-spec
# or
rpt --helpdump
```

---

## 🤖 Why `--dump-spec` Exists

When an AI agent is asked by a user:
> *"Help me get this non-standard Windows game running with `rpt`!"*

The agent typically does not know:
1. Which GPUs are installed on the user's laptop.
2. Which external monitors are currently plugged in.
3. What CPU P-core threads exist on the machine.
4. Which Proton runners (GE-Proton, DW-Proton, Proton Experimental) exist on the user's filesystem.
5. The exact schema of `.proton-config.toml`.

By running `rpt --dump-spec`, the AI agent immediately receives a single, unified JSON document with complete system and launcher awareness.

---

## 📋 JSON Structure Overview

Running `rpt --dump-spec` outputs:

```json
{
  "schema_version": "rpt-spec-v1",
  "rpt_version": "0.5.0-alpha",
  "description": "Machine-readable runtime specification for AI agents configuring games under run-proton-tui.",
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
    ]
  },
  "best_practices": {
    "2d_utility_isolation": "2D utilities, setup installers (Setup.exe), and web launchers (Qt5/CEF/Electron) MUST have NVAPI and Steam Deck flags stripped, and run on host iGPU without Gamescope to avoid glibc double-free memory corruption.",
    "decoupled_gamescope": "On hybrid laptops, run Gamescope on the host iGPU (KWin compositor) and place prime-run INSIDE the sandbox on the game binary. Running prime-run gamescope exhausts Intel GEM memory (execbuf ENOMEM) and crashes KWin.",
    "drm_display_routing": "External HDMI/DP ports are typically hardwired to the dGPU on hybrid laptops. Query /sys/class/drm/card*-*/status without sudo to target external displays directly and eliminate PCIe double-bounce stutter.",
    "screenshot_preservation": "Windows games save screenshots to C:\\users\\steamuser\\Pictures. Always preserve Pictures alongside Saved Games and AppData during prefix wipes."
  }
}
```

---

## 🛠️ How an Agent Configures a Game

When a user introduces a new game, the AI agent can simply:
1. Execute `rpt --dump-spec`.
2. Inspect the installed Proton runners and pick the best candidate (e.g. GE-Proton).
3. Generate the tailored `.proton-config.toml` in the game's directory with appropriate environment variables and profile separation.
4. If the game needs special setup (e.g. `/dev/shm` RAM symlinks or custom registry overrides), write `./hooks/pre_launch.sh`.
5. Tell the user: *"Done! Simply run `rpt` and hit Enter."*
