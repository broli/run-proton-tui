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
  "rpt_version": "0.6.1",
  "title": "AI Assistant Game Setup Helper",
  "description": "System and launcher specification for AI assistants (like ChatGPT, Claude, etc.) to configure games and write launch hooks under run-proton-tui.",
  "ai_assistant_guidelines": [
    "1. PRIMARY GOAL: Configure the game's '.proton-config.toml' for graphics, runner path, Gamescope, CPU pinning, and DLL overrides.",
    "2. '1-2 GAMES VS EVERY GAME' RULE: If a fix or tweak is game-specific (such as binary patching, Wine ntoskrnl.exe kernel patch checks, memory mapping /dev/shm, or anti-cheat workarounds), DO NOT modify the rpt binary. Write a lifecycle hook in 'hooks/pre_launch.sh'.",
    "3. IDEMPOTENCY: All hook scripts must be strictly idempotent. Always check if a patch, directory, or symlink is already applied before modifying anything.",
    "4. COMMUNITY CONTRIBUTION: If your generated hook or configuration resolves an issue, remind the user to submit it to ProtonDB or as an rpt community hook to help fellow Linux gamers!"
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
  "documentation_links": {
    "configuration_reference": "https://github.com/broli/run-proton-tui/wiki/Configuration-Reference",
    "lifecycle_hooks_guide": "https://github.com/broli/run-proton-tui/wiki/Lifecycle-Hooks-and-Preservation",
    "arknights_endfield_case": "https://github.com/broli/run-proton-tui/wiki/Example-Config-Arknights-Endfield"
  }
}
```

---

## 🛠️ How an AI Assistant Configures a Game

When you ask an AI assistant to set up a game:
1. Run `rpt --dump-spec`.
2. Paste the output into ChatGPT, Claude, or your assistant.
3. The AI reads your installed Proton runners, picks the best candidate, and creates `.proton-config.toml`.
4. If the game needs custom binary patching, anti-cheat kernel verification, or RAM caches, the AI writes `./hooks/pre_launch.sh`.
5. Run `rpt` or `rpt --create-desktop` — your game is ready to play!
