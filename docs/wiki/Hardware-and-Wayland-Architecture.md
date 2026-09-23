# Hardware & Wayland Architecture

`rpt` is engineered specifically to ensure rock-solid stability and zero-stutter performance on modern Linux gaming hardware, especially modern hybrid laptops (Intel/AMD iGPU + NVIDIA RTX dGPU) running Wayland compositors (KDE Plasma 6, GNOME, Hyprland).

---

## 💥 The KWin Wayland Memory Hazard (`execbuf ENOMEM`)

### The Symptom
On modern hybrid laptops, launching games with `prime-run gamescope ...` often causes the entire desktop session to crash back to SDDM/GDM within seconds, terminating all running applications.

```
kwin_wayland[1513]: intel: the execbuf ioctl keeps returning ENOMEM
systemd-coredump[1620]: Process 1513 (kwin_wayland) terminated abnormally with signal 6/ABRT
```

### The Root Cause
1. The host desktop compositor (`kwin_wayland`) executes on the integrated GPU (e.g. Intel Iris Xe / `libgallium`).
2. Putting `prime-run` in front of `gamescope` forces the Gamescope Wayland/Vulkan process itself to initialize as an NVIDIA client.
3. Gamescope continuously submits rendered framebuffers across the PCIe bus as DMA-BUFs directly into Intel's KWin compositor.
4. When games recreate Direct3D swapchains or change resolutions, the sudden burst of buffer allocations exhausts the Intel kernel GEM execution buffer pool (`ENOMEM`), crashing KWin.

### The Decoupled Gamescope Architecture
`rpt` strictly adheres to decoupled sandboxing:
- **Gamescope MUST run as an unprivileged client on the host iGPU** driving your compositor.
- **`prime-run` MUST be placed INSIDE the sandbox**, executing strictly on the Windows game binary.

```
Host Wayland Session (Intel iGPU / KWin)
  └── Gamescope Sandbox (Runs natively on Intel iGPU)
        └── Wrapper Script (.rpt_runner.sh)
              └── prime-run taskset -c 0-11 Proton -> Game.exe (NVIDIA dGPU)
```

---

## 🖥️ Zero-ACL World-Readable DRM Sysfs Display Routing

To route Gamescope to external monitors, older tools rely on external desktop-specific utilities (`kscreendoctor`, `xrandr`, `wlr-randr`) or require elevated privileges.

### How `rpt` Solves It
`rpt` inspects the Linux kernel DRM subsystem directly:
```
/sys/class/drm/card*-*/status
```
- **Permissions**: These files have standard file permissions **`0444` (`-r--r--r--`)**.
- **Zero Root / Zero Sudo**: Completely unprivileged and accessible by any user space application.
- **Desktop Agnostic**: Works identically under KDE Plasma, GNOME, Hyprland, Sway, XFCE, and headless sessions.

### Eliminating the PCIe "Double-Bounce"
On many hybrid laptops, the external HDMI or DisplayPort is wired directly to the NVIDIA dGPU, while the laptop screen is wired to the Intel iGPU.
- If Gamescope is forced to the wrong GPU output:
  `Game (NVIDIA) -> PCIe -> Gamescope (Intel) -> PCIe -> HDMI Output (NVIDIA)`
  Frames cross the PCIe bus twice ("double-bounce"), causing 100% bus congestion and severe FPS stutter.
- By reading `/sys/class/drm/` and passing `--prefer-output HDMI-A-1`, `rpt` routes frames directly to the hardwired display at its native refresh rate.

---

## 🛠️ 2D Utilities vs 3D Game Isolation

### The Qt5/CEF WebEngine Crash
Many game launchers, updaters, and repack installers (`Setup.exe`, `Launcher.exe`, `Games.exe`) are built with Qt 5/6 WebEngine or Chromium Embedded Framework (CEF).

When executed with NVIDIA prime offloading and `PROTON_ENABLE_NVAPI=1` on Wayland:
1. Chromium's GPU process attempts to query NVAPI via Wine.
2. Wine's NVAPI returns an unhandled error state.
3. Chromium's error handler corrupts its heap during parser teardown:
   ```
   double free or corruption (!prev)
   pthread_kill + 299
   ```
4. Glibc aborts the process, worker threads freeze, and the launcher window never renders.

### `rpt`'s Universal Rule
Whenever an executable is classified as a 2D utility or installer, or when `UsePrimeRun` is false:
- `rpt` automatically strips:
  - `PROTON_ENABLE_NVAPI`
  - `DXVK_ENABLE_NVAPI`
  - `PROTON_ENABLE_NGX_UPDATER`
  - `STEAMOS` and `STEAMDECK`
- The utility executes cleanly on the host iGPU without Gamescope overhead.

---

## ⚡ Intel Hybrid CPU Thread Pinning (P-Cores vs E-Cores)

Intel 12th, 13th, and 14th Gen processors feature a hybrid architecture combining high-power Performance cores (P-cores) with low-power Efficiency cores (E-cores).

When the Linux kernel thread scheduler migrates game worker threads or audio threads to E-cores, games experience sudden framerate drops and micro-stutters.

`rpt` queries:
```
/sys/devices/cpu_core/cpus
```
If a hybrid architecture is detected, `rpt` discovers the P-core thread range (e.g. `0-11`) and executes the game under:
```bash
taskset -c 0-11 prime-run ...
```
This restricts all 3D game threads strictly to high-performance cores, completely eliminating thread migration stutter.
