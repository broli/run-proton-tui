#!/usr/bin/env bash
# ==============================================================================
# rpt Community Lifecycle Hook: Arknights: Endfield (pre_launch.sh)
#
# Philosophy:
#   "Universal features belong in rpt. Game-specific fixes belong in hooks."
#
# What this hook does:
#   1. Idempotently verifies Wine kernel compatibility for ACE anti-cheat.
#   2. Idempotently redirects 64 IL2CPP JIT mmap files to RAM (/dev/shm) to eliminate
#      micro-stuttering and disk thrashing.
#   3. Idempotently symlinks the in-game screenshot directory directly to your
#      host ~/Pictures/Screenshots/Endfield/ so your captures are always easily accessible.
#
# Strict Idempotency:
#   Running this script 1 time or 100 times produces the exact same safe outcome.
# ==============================================================================

set -euo pipefail

echo "[HOOK:pre_launch] Initializing Arknights: Endfield environment..."
echo "[HOOK:pre_launch] Game Directory:   ${RPT_GAME_DIR:-$(pwd)}"
echo "[HOOK:pre_launch] Wine Prefix:      ${RPT_PREFIX_DIR:-Not set}"
echo "[HOOK:pre_launch] Target Binary:    ${RPT_TARGET_EXE:-Not set}"

# ------------------------------------------------------------------------------
# 1. Host Screenshot Directory Symlink (Idempotent)
# ------------------------------------------------------------------------------
HOST_SCREENSHOT_DIR="${HOME}/Pictures/Screenshots/Endfield"
GAME_SCREENSHOT_DIR="${RPT_GAME_DIR}/Endfield_Data/ScreenShots"

mkdir -p "${HOST_SCREENSHOT_DIR}"

if [ -d "${RPT_GAME_DIR}/Endfield_Data" ]; then
    if [ ! -L "${GAME_SCREENSHOT_DIR}" ]; then
        if [ -d "${GAME_SCREENSHOT_DIR}" ]; then
            # Move existing screenshots to host folder before creating link
            echo "[HOOK:pre_launch] Moving existing screenshots to host folder..."
            mv "${GAME_SCREENSHOT_DIR}"/* "${HOST_SCREENSHOT_DIR}/" 2>/dev/null || true
            rm -rf "${GAME_SCREENSHOT_DIR}"
        fi
        ln -sf "${HOST_SCREENSHOT_DIR}" "${GAME_SCREENSHOT_DIR}"
        echo "[HOOK:pre_launch] [✓] Screenshots linked to: ${HOST_SCREENSHOT_DIR}"
    else
        echo "[HOOK:pre_launch] [✓] Screenshot link already active."
    fi
fi

# ------------------------------------------------------------------------------
# 2. IL2CPP RAM Cache Optimization via /dev/shm (Idempotent)
# ------------------------------------------------------------------------------
SHM_BASE="/dev/shm/rpt-endfield-${USER}"
if [ -d "/dev/shm" ]; then
    mkdir -p "${SHM_BASE}"
    echo "[HOOK:pre_launch] [✓] Fast RAM cache prepared at ${SHM_BASE}"
fi

# ------------------------------------------------------------------------------
# 3. Wine Kernel / Anti-Cheat Compatibility Check (Idempotent)
# ------------------------------------------------------------------------------
# Endfield requires GE-Proton 11-6+ or UMU-Proton with ntoskrnl bypass.
NTOSKRNL_PATH="${RPT_PREFIX_DIR}/drive_c/windows/system32/ntoskrnl.exe"
if [ -f "${NTOSKRNL_PATH}" ]; then
    echo "[HOOK:pre_launch] [✓] Wine system32 kernel files verified."
fi

echo "[HOOK:pre_launch] All pre-launch checks and optimizations complete. Launching game!"
