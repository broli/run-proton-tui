# Lifecycle Hooks & Save/Screenshot Preservation

`rpt` provides two complementary mechanisms to keep games and user data safe:
1. **The Preservation Engine**: Safely archiving player saves and screenshots when a prefix is cleaned or reset.
2. **Standardized Lifecycle Hooks**: Executing arbitrary shell scripts before launch and after exit.

---

## 🛡️ The Preservation Engine (Not Just a "Backup")

When gaming on Wine and Proton, troubleshooting often requires wiping a corrupt prefix (`rpt -c` or `rpt --clean`). Standard Wine tools delete the entire `proton-prefix/` directory, wiping all saved progress and player photos.

### What `rpt` Preserves Automatically
`rpt` searches `drive_c/users/<user>/` and safely archives the 5 standard Windows user profile folders:
- **`Pictures`**: In-game camera photos, photo-mode captures, and screenshots (frequently overlooked by other launchers, leading to permanent photo loss!).
- **`Saved Games`**: Native Windows game save storage.
- **`Documents`**: Classic Windows game save and config directory.
- **`AppData/Local`**: Modern Unity, Unreal, and indie game save files.
- **`AppData/Roaming`**: Application settings and saves.

### Configurable Preservation Directory
By default, preserved files are saved to:
```
~/Games/Backups/<GameName>/<YYYY-MM-DD_HH-MM-SS>/
```
You can customize where your preserved files are placed in `.proton-config.toml`:
```toml
# Save preserved files directly to your preferred folder:
backup_dir = "~/winegames/saves"
# or:
backup_dir = "~/MySaves"
```

### Extra Preservation Paths
For non-standard games that save progress in arbitrary locations outside Windows user folders (e.g. directly in `drive_c/GameSaves/`), declare them in `.proton-config.toml`:
```toml
[filesystem]
extra_backup_paths = [
  "drive_c/GameSaves",
  "drive_c/users/steamuser/AppData/LocalLow/CustomStudio"
]
```

---

## ⚡ Lifecycle Shell Hooks

For complex operations that no static configuration file can anticipate, `rpt` provides **Lifecycle Shell Hooks**.

### Hook Types
- **`pre_launch.sh`**: Runs immediately before Gamescope and Proton start.
- **`post_exit.sh`**: Runs after the game and any supervised background processes terminate, right before prefix lock cleanup.

### Search Priority Order
`rpt` searches for hooks in the following locations, prioritizing your unpacked game folder:
1. Explicitly configured path in `.proton-config.toml`: `pre_launch_hook = "..."`
2. **Unpacked game directory**: `$PWD/hooks/<type>.sh`
3. **Unpacked game directory**: `$PWD/.rpt/hooks/<type>.sh`
4. **Unpacked game directory root**: `$PWD/<type>.sh`
5. Custom directories declared in config: `hook_dirs = [...]`
6. User global config directory: `$HOME/.config/rpt/hooks/<type>.sh`
7. User legacy directory: `$HOME/.rpt/hooks/<type>.sh`

---

## 🌐 Exported Environment Variables

Whenever `rpt` executes a lifecycle hook, it exports standardized environment variables into the script:

| Variable | Description | Example |
| :--- | :--- | :--- |
| **`$RPT_GAME_DIR`** | Absolute path to the game directory | `/home/user/Games/Endfield` |
| **`$RPT_PREFIX_DIR`** | Absolute path to the game's Wine prefix | `/home/user/Games/Endfield/proton-prefix` |
| **`$RPT_TARGET_EXE`** | Relative path to target Windows executable | `games/Endfield/Endfield.exe` |
| **`$RPT_PROTON_PATH`** | Path to the active Proton runner binary | `/home/user/.../GE-Proton11-6/proton` |
| **`$RPT_GAMESCOPE_DISPLAY`** | Nested Xwayland display number | `:1` (or empty if native) |
| **`$RPT_HOOK_TYPE`** | Lifecycle phase of current hook | `pre_launch` or `post_exit` |

---

## 💡 Practical Examples of Lifecycle Hooks

### Example 1: Redirecting JIT Memory-Mapped Files to RAM (`/dev/shm`)
Engines with heavy JIT compilation (like Unity IL2CPP) generate bursts of temporary worker files that cause SSD micro-stutters. A `pre_launch.sh` hook redirects them to RAM:

```bash
#!/bin/bash
# Place in: ./hooks/pre_launch.sh
LOCAL_LOW="$RPT_PREFIX_DIR/pfx/drive_c/users/steamuser/AppData/LocalLow"
mkdir -p "$LOCAL_LOW"

for i in $(seq 0 63); do
    f="HG_IL2CPP_MMAP$i"
    [ -e "$LOCAL_LOW/$f" ] && [ ! -L "$LOCAL_LOW/$f" ] && rm -f "$LOCAL_LOW/$f"
    [ ! -L "$LOCAL_LOW/$f" ] && ln -sf "/dev/shm/$f" "$LOCAL_LOW/$f"
done
```

### Example 2: Persistent Photo-Mode Symlink
Symlink in-game camera photos directly to your Linux desktop pictures directory:

```bash
#!/bin/bash
# Place in: ./hooks/pre_launch.sh
USER_PICS="$RPT_PREFIX_DIR/pfx/drive_c/users/steamuser/Pictures"
mkdir -p "$USER_PICS"
mkdir -p "$HOME/Pictures/MyGamePhotos"

if [ ! -L "$USER_PICS/MyGame" ]; then
    rm -rf "$USER_PICS/MyGame"
    ln -sf "$HOME/Pictures/MyGamePhotos" "$USER_PICS/MyGame"
fi
```

### Example 3: Running a Companion Auto-Typer or Clipboard Bridge Daemon
```bash
#!/bin/bash
# Place in: ./hooks/pre_launch.sh
if [ -n "$RPT_GAMESCOPE_DISPLAY" ] && [ -x "$HOME/bin/gamescope-clip-bridge" ]; then
    "$HOME/bin/gamescope-clip-bridge" &
    echo $! > /tmp/rpt-clip-bridge.pid
fi
```
And terminate it in `post_exit.sh`:
```bash
#!/bin/bash
# Place in: ./hooks/post_exit.sh
if [ -f /tmp/rpt-clip-bridge.pid ]; then
    kill "$(cat /tmp/rpt-clip-bridge.pid)" 2>/dev/null
    rm -f /tmp/rpt-clip-bridge.pid
fi
```

### Example 4: Verifying & Auto-Applying Proton Binary Patches (e.g. Anti-Cheat `ntoskrnl.exe`)

Because `rpt` exports the exact path of the active Proton binary in **`$RPT_PROTON_PATH`**, you can write a `pre_launch.sh` hook that checks whether the current Proton runner has required kernel patches (such as `ntoskrnl.exe` patches for Tencent Anti-Cheat Expert / ACE), and applies them automatically if Proton was freshly downloaded or updated.

```bash
#!/bin/bash
# Place in: ./hooks/pre_launch.sh
set -e

# Derive the Proton installation root from $RPT_PROTON_PATH
PROTON_DIR="$(dirname "$RPT_PROTON_PATH")"
NTOSKRNL="$PROTON_DIR/files/lib/wine/x86_64-windows/ntoskrnl.exe"

# Check standard Proton paths
[ ! -f "$NTOSKRNL" ] && NTOSKRNL="$PROTON_DIR/lib/wine/x86_64-windows/ntoskrnl.exe"

if [ -f "$NTOSKRNL" ]; then
    echo "[rpt:hook] Inspecting Proton kernel binary: $NTOSKRNL"
    
    # Check if the kernel binary is already patched; if unpatched, patch automatically
    python3 - << 'EOF' "$NTOSKRNL"
import sys, os, shutil

path = sys.argv[1]
with open(path, "r+b") as f:
    data = bytearray(f.read())
    
# Target offset 0x43a8 for PsGetProcessExitProcessCalled
offset = 0x43a8
payload = bytes.fromhex("31c0c3" + "90" * 21)

if data[offset:offset+len(payload)] == payload:
    print("[rpt:hook] ✓ ntoskrnl.exe is already patched for Anti-Cheat Expert.")
else:
    print("[rpt:hook] ! ntoskrnl.exe is unpatched; auto-applying binary patch...")
    if not os.path.exists(path + ".orig"):
        shutil.copy2(path, path + ".orig")
    data[offset:offset+len(payload)] = payload
    
    # Target offset 0x4f90 for SeQueryInformationToken
    sec_offset = 0x4f90
    sec_payload = bytes.fromhex("b8020000c0c3" + "90" * 18)
    data[sec_offset:sec_offset+len(sec_payload)] = sec_payload
    
    f.seek(0)
    f.write(data)
    f.truncate()
    print("[rpt:hook] ✓ Successfully applied Proton kernel binary patches!")
EOF
fi
```

This guarantees that even when ProtonUp-Qt or Steam updates your GE-Proton build in the background, your game will always launch with verified, working kernel binaries without manual re-patching!

