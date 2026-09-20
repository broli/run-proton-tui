# rpt Architecture & Technical Reference

## 1. Hardware Topology & Wayland Stability

### The KWin Wayland Memory Exhaustion Hazard
On modern hybrid GPU laptops (e.g. Intel Iris Xe + NVIDIA RTX 4060):
1. The desktop compositor (`kwin_wayland`) executes on the Intel integrated graphics (`iris` driver).
2. If Gamescope is launched under `prime-run` or forced to use NVIDIA Vulkan device (`10de:28e0`), Gamescope becomes an NVIDIA client that streams DMA-BUFs into Intel KWin.
3. During Direct3D swapchain recreation (game startup, resolution shifts, DXVK reset), Intel's kernel execution buffer driver runs out of memory:
   ```
   kwin_wayland[PID]: intel: the execbuf ioctl keeps returning ENOMEM
   systemd-coredump: Process 1513 (kwin_wayland) terminated abnormally with signal 6/ABRT
   ```
4. This crashes KDE Plasma instantly, terminating all user applications.

### The Decoupled Gamescope Architecture
`rpt` strictly adheres to the following pipeline:
```mermaid
flowchart TD
    Host["Host Wayland Session (Intel iGPU / KWin)"] -->|"Native Wayland Protocol"| GS["Gamescope Sandbox (Host iGPU)"]
    GS -->|"Runs Sandbox Child Wrapper"| Wrapper[".rpt_runner.sh"]
    Wrapper -->|"Executes with prime-run & taskset -c 0-11"| Game["Proton / Wine -> Game.exe (NVIDIA dGPU)"]
    HostClip["Host Clipboard (wl-paste)"] -->|"Socket / Passive Sync"| ClipBridge["gamescope-clip-bridge"]
    ClipBridge -->|"X11 Clipboard (:1)"| Game
```

---

## 2. 2D Utilities vs 3D Game Isolation

Executables are classified into two distinct operational profiles:

### 2D Utilities, Launchers & Installers
- **Signatures**: `*Launcher*.exe`, `*Update*.exe`, `*Setup*.exe`, `*unins*.exe`, `Games.exe`, `*CrashReport*.exe`.
- **Rationale**: Many 2D game launchers and installers are built with Qt 5.15 WebEngine or Chromium Embedded Framework (CEF). When executed with `prime-run` on Wayland sessions, their NVAPI probes trigger glibc `double free or corruption (!prev)` memory corruption.
- **Default Profile**: Host Intel iGPU, Gamescope **OFF**, CPU pinning **OFF**, Power profile **Default**.

### 3D Game Binaries
- **Signatures**: `*Shipping*.exe`, `*Game*.exe`, `Endfield.exe`, or standard game binaries.
- **Default Profile**: NVIDIA RTX 4060 (`prime-run`), Gamescope **ON** (1080p @ 75Hz $\to$ external monitor), CPU pinning **P-Cores (0-11)**, Power profile **Performance**.

---

## 3. Wine Process Isolation & Socket Cleanup

Wine creates IPC server sockets in `/tmp/.wine-<UID>/` using the device and inode numbers of the prefix directory:
`server-<st_dev:x>-<st_ino:x>`

`rpt` manages this lifecycle via `internal/prefix`:
1. **Flock Testing**: Tests `/tmp/.wine-<UID>/server-*/lock` files with non-blocking `flock`. If lock acquisition succeeds, the wineserver process is dead and the directory is safely purged.
2. **Process Pruning**: Inspects `/proc/$pid/environ` for `WINEPREFIX=` before terminating processes, ensuring games running under Steam or Heroic are never touched.
