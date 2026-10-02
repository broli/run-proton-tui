#!/usr/bin/env bash
# ==============================================================================
# 🎮 rpt Lifecycle Hook Template: post_exit.sh
# ==============================================================================
# This template is executed immediately AFTER the game exits.
# Use it for cloud save backups, syncing saves with rclone, and cleanup.
#
# 💡 PHILOSOPHY:
#   - Generic teardown (wineserver -k, wineserver -w, prefix process cleanup,
#     and /tmp/.wine-<UID> stale lock removal) is handled automatically by rpt.
#   - Only game-specific cleanup (cloud save sync, RAM cache removal, external
#     daemons) belongs in this script!
#
# 🚫 WHAT NOT TO DO IN THIS HOOK:
#   - Do NOT run 'wineserver -k', 'wineserver -w', or kill Wine processes.
#   - Do NOT delete or tamper with /tmp/.wine-<UID> lock files.
#   rpt automatically flushes the prefix, isolates lingering processes, and
#   safely cleans dead socket locks in /tmp/.wine-<UID> upon game exit!
#
# 📦 ENVIRONMENT VARIABLES INJECTED BY rpt:
#   $RPT_GAME_DIR           Absolute path to game root directory
#   $RPT_PREFIX_DIR         Absolute path to Wine prefix (.prefix)
#   $RPT_TARGET_EXE         Relative path to target executable
#   $RPT_PROTON_PATH        Absolute path to selected Proton runner
#   $RPT_HOOK_TYPE          Phase name: "post_exit"
#   $RPT_CHILD_PID          Process ID of the game process (if available)
# ==============================================================================

set -euo pipefail

echo "[HOOK:post_exit] Game process finished. Running post-exit tasks..."

# ------------------------------------------------------------------------------
# RECIPE 1: Post-Exit Cloud Save Push (rclone)
# Automatically push your updated save games to Google Drive, Dropbox, or Nextcloud!
# ------------------------------------------------------------------------------
# SAVE_DIR="${RPT_PREFIX_DIR}/drive_c/users/steamuser/AppData/Local/MyGame/Saved"
# if command -v rclone >/dev/null 2>&1 && [ -d "${SAVE_DIR}" ]; then
#     echo "[HOOK:post_exit] Uploading new save files to cloud storage..."
#     # rclone sync "${SAVE_DIR}" "gdrive:GameSaves/MyGame" || true
#     echo "[HOOK:post_exit] [✓] Cloud save sync complete!"
# fi

# ------------------------------------------------------------------------------
# RECIPE 2: Local Timestamped Save Backup
# Create a local tar archive of your save data for safety.
# ------------------------------------------------------------------------------
# BACKUP_DIR="${RPT_GAME_DIR}/.rpt_backups"
# if [ -d "${SAVE_DIR:-}" ]; then
#     mkdir -p "${BACKUP_DIR}"
#     TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
#     tar -czf "${BACKUP_DIR}/saves_${TIMESTAMP}.tar.gz" -C "${SAVE_DIR}" . 2>/dev/null || true
#     echo "[HOOK:post_exit] [✓] Local backup created: saves_${TIMESTAMP}.tar.gz"
# fi

# ------------------------------------------------------------------------------
# RECIPE 3: Transient RAM Cleanup
# Clean up temporary mmap files created in /dev/shm
# ------------------------------------------------------------------------------
# SHM_DIR="/dev/shm/rpt-cache-${USER}"
# if [ -d "${SHM_DIR}" ]; then
#     rm -rf "${SHM_DIR}"
#     echo "[HOOK:post_exit] [✓] RAM cache cleaned up."
# fi

echo "[HOOK:post_exit] [✓] Post-exit cleanup finished."
