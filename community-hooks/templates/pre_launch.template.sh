#!/usr/bin/env bash
# ==============================================================================
# 🎮 rpt Lifecycle Hook Template: pre_launch.sh
# ==============================================================================
# This template is designed for gamers and AI Assistants (ChatGPT, Claude, etc.)
# to implement game-specific fixes, tweaks, and optimizations before the game starts.
#
# 💡 PHILOSOPHY:
#   - Generic, universal features (Gamescope, CPU pinning, Wine prefix creation)
#     are handled automatically by rpt.
#   - Quirks needed by only 1 or 2 games (anti-cheat shims, binary patches, RAM
#     caches, mod managers) belong in this script!
#
# 🚫 WHAT NOT TO DO IN THIS HOOK:
#   - Do NOT run 'wineserver -k', 'wineserver -w', or kill Wine processes.
#   - Do NOT delete or tamper with /tmp/.wine-<UID> lock files.
#   rpt automatically flushes the prefix, isolates lingering processes, and
#   cleans stale socket locks via non-blocking flock before executing this hook!
#
# ⚠️ RULE OF IDEMPOTENCY:
#   This hook runs every single time the game launches. It MUST be idempotent:
#   executing it once or 100 times must produce the exact same safe outcome!
#
# 📦 ENVIRONMENT VARIABLES INJECTED BY rpt:
#   $RPT_GAME_DIR           Absolute path to game root directory
#   $RPT_PREFIX_DIR         Absolute path to Wine prefix (.prefix)
#   $RPT_TARGET_EXE         Relative path to target executable (e.g. Game.exe)
#   $RPT_PROTON_PATH        Absolute path to selected Proton runner
#   $RPT_GAMESCOPE_DISPLAY  Gamescope display socket (e.g. :1, or empty)
#   $RPT_HOOK_TYPE          Phase name: "pre_launch"
# ==============================================================================

set -euo pipefail

echo "[HOOK:pre_launch] Executing pre-launch hook for: ${RPT_TARGET_EXE}"

# ------------------------------------------------------------------------------
# RECIPE 1: Idempotent Binary or File Patching
# Check if the target has already been patched BEFORE modifying anything!
# ------------------------------------------------------------------------------
# TARGET_FILE="${RPT_PREFIX_DIR}/drive_c/windows/system32/example.dll"
# if [ -f "${TARGET_FILE}" ]; then
#     # Check if already patched by verifying an offset or hash:
#     if ! cmp -s "${TARGET_FILE}" "${TARGET_FILE}.patched" 2>/dev/null; then
#         echo "[HOOK:pre_launch] Applying compatibility patch..."
#         cp -n "${TARGET_FILE}" "${TARGET_FILE}.bak" # Safe backup
#         # apply patch here...
#     else
#         echo "[HOOK:pre_launch] [✓] Patch already present, skipping."
#     fi
# fi

# ------------------------------------------------------------------------------
# RECIPE 2: Fast RAM Cache via /dev/shm
# Move heavy shader caches, JIT files, or temporary indices to RAM for stutter-free gaming.
# ------------------------------------------------------------------------------
# SHM_DIR="/dev/shm/rpt-cache-${USER}"
# if [ -d "/dev/shm" ]; then
#     mkdir -p "${SHM_DIR}"
#     echo "[HOOK:pre_launch] [✓] RAM cache ready at ${SHM_DIR}"
# fi

# ------------------------------------------------------------------------------
# RECIPE 3: Idempotent Host Symlinking (Saves, Screenshots, Mods)
# Safely link internal prefix folders to standard host locations.
# ------------------------------------------------------------------------------
# HOST_DIR="${HOME}/Pictures/Screenshots/MyGame"
# INTERNAL_DIR="${RPT_PREFIX_DIR}/drive_c/users/steamuser/AppData/Local/MyGame/Saved"
# mkdir -p "${HOST_DIR}"
# if [ -d "$(dirname "${INTERNAL_DIR}")" ] && [ ! -L "${INTERNAL_DIR}" ]; then
#     ln -sf "${HOST_DIR}" "${INTERNAL_DIR}"
#     echo "[HOOK:pre_launch] [✓] Linked internal directory to ${HOST_DIR}"
# fi

# ------------------------------------------------------------------------------
# RECIPE 4: Pre-Launch Cloud Save Pull (rclone)
# Pull latest save files from cloud storage before launching.
# ------------------------------------------------------------------------------
# if command -v rclone >/dev/null 2>&1; then
#     echo "[HOOK:pre_launch] Syncing cloud saves from Google Drive / Nextcloud..."
#     # rclone sync "gdrive:GameSaves/MyGame" "${INTERNAL_DIR}" || true
# fi

echo "[HOOK:pre_launch] [✓] Pre-launch tasks completed successfully."
