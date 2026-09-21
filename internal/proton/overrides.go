package proton

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/prefix"
)

// DLLPreset represents a common Windows mod or compatibility hook.
type DLLPreset struct {
	Name        string // e.g. "UE4SS Unreal Mod Loader"
	DLL         string // e.g. "dwmapi"
	Mode        string // e.g. "native,builtin"
	Description string // e.g. "Loads Unreal Engine 4/5 Lua & C++ mods"
}

// GetStandardDLLPresets returns the curated list of common DLL overrides.
func GetStandardDLLPresets() []DLLPreset {
	return []DLLPreset{
		{
			Name:        "UE4SS Mod Loader",
			DLL:         "dwmapi",
			Mode:        "native,builtin",
			Description: "Required for Unreal Engine 4 & 5 mods (UE4SS)",
		},
		{
			Name:        "BepInEx / Unity Mod Loader",
			DLL:         "winhttp",
			Mode:        "native,builtin",
			Description: "Required for BepInEx modding in Unity / .NET games",
		},
		{
			Name:        "ReShade / SpecialK",
			DLL:         "dxgi",
			Mode:        "native,builtin",
			Description: "Loads custom DXGI graphics injectors and HDR hooks",
		},
		{
			Name:        "DInput8 Controller Hook",
			DLL:         "dinput8",
			Mode:        "native,builtin",
			Description: "Used by many mod managers and custom controller wrappers",
		},
		{
			Name:        "Version Mod Loader",
			DLL:         "version",
			Mode:        "native,builtin",
			Description: "Generic ASI and mod loader wrapper",
		},
	}
}

// IsWineDefaultDLL returns true if a DLL is an internal Wine/Proton system default
// (such as Visual C++ runtimes, UCRT, Windows API sets, or driver stubs).
func IsWineDefaultDLL(dll string) bool {
	lower := strings.ToLower(strings.TrimSpace(dll))
	if strings.HasPrefix(lower, "msvc") ||
		strings.HasPrefix(lower, "vcomp") ||
		strings.HasPrefix(lower, "vcrunt") ||
		strings.HasPrefix(lower, "vccor") ||
		strings.HasPrefix(lower, "concrt") ||
		strings.HasPrefix(lower, "atl") ||
		strings.HasPrefix(lower, "api-ms-win") ||
		lower == "ucrtbase" ||
		lower == "nvcuda" ||
		lower == "atiadlxx" ||
		lower == "lsteamclient" {
		return true
	}
	return false
}

// ReadRegistryOverrides reads user/game DLL overrides from user.reg, automatically
// filtering out Wine's internal system runtime defaults.
func ReadRegistryOverrides(prefixDir string) (map[string]string, error) {
	userOverrides, _, err := ReadRegistryOverridesSeparated(prefixDir)
	return userOverrides, err
}

// ReadRegistryOverridesSeparated reads active DLL overrides from user.reg and separates
// user/game overrides from Wine's internal system defaults.
func ReadRegistryOverridesSeparated(prefixDir string) (userOverrides, systemDefaults map[string]string, err error) {
	pfx := prefix.GetCanonicalPfx(prefixDir)
	userReg := filepath.Join(pfx, "user.reg")

	f, err := os.Open(userReg)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	userOverrides = make(map[string]string)
	systemDefaults = make(map[string]string)
	inOverridesSection := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[Software\\\\Wine\\\\DllOverrides]") {
			inOverridesSection = true
			continue
		}

		if inOverridesSection {
			if strings.HasPrefix(line, "[") {
				// Entered another section
				break
			}
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], "\"")
				val := strings.Trim(parts[1], "\"")
				if IsWineDefaultDLL(key) {
					systemDefaults[key] = val
				} else {
					userOverrides[key] = val
				}
			}
		}
	}

	return userOverrides, systemDefaults, nil
}


func resolveWineBin(protonPath string) (string, error) {
	if protonPath != "" {
		pDir := filepath.Dir(protonPath)
		wineBin := filepath.Join(pDir, "files/bin/wine")
		if _, err := os.Stat(wineBin); err == nil {
			return wineBin, nil
		}
		wineBin = filepath.Join(pDir, "dist/bin/wine")
		if _, err := os.Stat(wineBin); err == nil {
			return wineBin, nil
		}
	}
	if sysWine, err := exec.LookPath("wine"); err == nil {
		return sysWine, nil
	}
	return "", fmt.Errorf("wine binary not found to execute reg command")
}

// SetRegistryOverride writes a DLL override into the Wine prefix registry using 'wine reg add'.
func SetRegistryOverride(prefixDir string, protonPath string, dllName string, mode string) error {
	pfx := prefix.GetCanonicalPfx(prefixDir)
	wineBin, err := resolveWineBin(protonPath)
	if err != nil {
		return err
	}

	regKey := `HKCU\Software\Wine\DllOverrides`
	cmd := exec.Command(wineBin, "reg", "add", regKey, "/v", dllName, "/d", mode, "/f")
	cmd.Env = append(os.Environ(), "WINEPREFIX="+pfx, "WINEDEBUG=-all")

	return cmd.Run()
}

// DeleteRegistryOverride deletes a specific DLL override under HKCU\Software\Wine\DllOverrides.
func DeleteRegistryOverride(prefixDir string, protonPath string, dllName string) error {
	pfx := prefix.GetCanonicalPfx(prefixDir)
	wineBin, err := resolveWineBin(protonPath)
	if err != nil {
		return err
	}

	regKey := `HKCU\Software\Wine\DllOverrides`
	cmd := exec.Command(wineBin, "reg", "delete", regKey, "/v", dllName, "/f")
	cmd.Env = append(os.Environ(), "WINEPREFIX="+pfx, "WINEDEBUG=-all")

	return cmd.Run()
}

// ClearRegistryOverrides deletes all DLL overrides under HKCU\Software\Wine\DllOverrides.
func ClearRegistryOverrides(prefixDir string, protonPath string) error {
	pfx := prefix.GetCanonicalPfx(prefixDir)
	wineBin, err := resolveWineBin(protonPath)
	if err != nil {
		return err
	}

	regKey := `HKCU\Software\Wine\DllOverrides`
	cmd := exec.Command(wineBin, "reg", "delete", regKey, "/f")
	cmd.Env = append(os.Environ(), "WINEPREFIX="+pfx, "WINEDEBUG=-all")

	return cmd.Run()
}

