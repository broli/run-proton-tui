# Steam Emulators & Compatibility Guide

## 1. What Does `lsteamclient.dll=d` Do?

In Valve's Proton and Wine, `lsteamclient.dll` is an internal bridge library ("shim"). Its purpose is to intercept Windows Steam API calls (`steam_api.dll`, `steam_api64.dll`, `steamclient.dll`) and translate them across the Wine boundary into native Linux Steam client IPC sockets (`~/.local/share/Steam/ubuntu12_32/steam`).

### The Problem in Non-Steam Releases
When running games that use Steam emulators (such as **Goldberg**, **CODEX**, **Rune**, **FLT**, or **ALI213**):
1. The game folder contains custom emulator DLLs designed to emulate the Steam backend locally.
2. If Wine's built-in `lsteamclient.dll` is enabled, Wine intercepts Steam API initialization before it reaches the emulator DLLs.
3. Wine attempts to connect to your real native Linux Steam client.
4. Because the game wasn't purchased or launched through Steam, the official Steam client pops up store pages, returns error codes, or crashes the game engine.

### The Solution
Setting:
```bash
WINEDLLOVERRIDES="lsteamclient=d;$WINEDLLOVERRIDES"
```
disables Wine's internal Linux bridge completely (`=d` means disabled). The game engine is forced to load the emulator DLLs found in its folder, allowing offline achievements, saves, and DLC unlocking to function smoothly.

---

## 2. Supported Steam Emulators

### Goldberg Steam Emulator
* **Configuration Files**: `steam_settings/`, `configs.app.ini`, `local_save/`.
* **DLC Unlocking**: Set `unlock_all=1` in `configs.app.ini`.
* **Save Locations**:
  - `local_save/` in the game directory, OR
  - `%APPDATA%\Goldberg SteamEmu Saves\` inside the Wine prefix.

### CODEX & Rune
* **Configuration Files**: `steam_emu.ini`, `steam_api64.cdx`.
* **DLC Unlocking**: Ensure `DLCUnlockall=1` is set in `steam_emu.ini`.
* **Save Locations**:
  - `%PUBLIC%\Documents\Steam\CODEX\<AppID>\` OR
  - `%APPDATA%\Steam\CODEX\<AppID>\`.

### Fairlight (FLT)
* **Configuration Files**: `flt.ini`.
* **Save Locations**:
  - `%APPDATA%\Fairlight\<AppID>\`.

### ALI213
* **Configuration Files**: `ali213.ini`, `SteamConfig.ini`.
* **Save Locations**:
  - `Profile/` inside the game directory.
