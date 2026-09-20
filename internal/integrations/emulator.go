package integrations

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// EmulatorInfo contains detected Steam emulator signatures and DLC unlock statuses.
type EmulatorInfo struct {
	Type       string // "Goldberg", "CODEX / Rune", "Fairlight (FLT)", "ALI213", "Standard / Steam"
	AppID      string // Detected AppID from emulator config or steam_appid.txt
	DLCStatus  string // "All Unlocked", "5 DLCs Defined", "Locked (unlock_all=0)", etc.
	ConfigFile string // Path to emulator config file (configs.app.ini, steam_emu.ini, etc.)
	Detected   bool   // True if any Steam emulator was detected
}

// ScanEmulators inspects the game directory tree for known Steam emulators and DLC configurations.
func ScanEmulators(gameDir string) (*EmulatorInfo, error) {
	info := &EmulatorInfo{
		Type:       "Standard / Steam",
		AppID:      "0",
		DLCStatus:  "Default / None",
		ConfigFile: "",
		Detected:   false,
	}

	// 1. Detect AppID via steam_appid.txt
	appIDFiles := findFilesNamed(gameDir, "steam_appid.txt", 6)
	if len(appIDFiles) > 0 {
		if content, err := os.ReadFile(appIDFiles[0]); err == nil {
			id := strings.TrimSpace(string(content))
			if id != "" {
				info.AppID = id
			}
		}
	}

	// 2. Check for Goldberg Steam Emulator (configs.app.ini, steam_settings/, local_save/)
	goldbergConfigs := findFilesNamed(gameDir, "configs.app.ini", 8)
	if len(goldbergConfigs) > 0 {
		info.Detected = true
		info.Type = "Goldberg Emulator"
		info.ConfigFile = goldbergConfigs[0]
		if val, err := readIniKey(goldbergConfigs[0], "unlock_all"); err == nil {
			if val == "1" {
				info.DLCStatus = "All Unlocked (unlock_all=1)"
			} else if val == "0" {
				info.DLCStatus = "Locked (unlock_all=0)"
			}
		}
	}

	if !info.Detected {
		// Goldberg alternate signature: steam_settings directory
		if stat, err := os.Stat(filepath.Join(gameDir, "steam_settings")); err == nil && stat.IsDir() {
			info.Detected = true
			info.Type = "Goldberg Emulator"
		}
	}

	// 3. Check for CODEX / Rune (steam_emu.ini, steam_api64.cdx)
	codexConfigs := findFilesNamed(gameDir, "steam_emu.ini", 8)
	if len(codexConfigs) > 0 {
		info.Detected = true
		info.Type = "CODEX / Rune"
		info.ConfigFile = codexConfigs[0]
		if dlcVal, err := readIniKey(codexConfigs[0], "DLCUnlockall"); err == nil && dlcVal == "1" {
			info.DLCStatus = "All Unlocked (DLCUnlockall=1)"
		}
		if idVal, err := readIniKey(codexConfigs[0], "AppId"); err == nil && idVal != "" && info.AppID == "0" {
			info.AppID = idVal
		}
	}

	// 4. Check for Fairlight (flt.ini)
	fltConfigs := findFilesNamed(gameDir, "flt.ini", 8)
	if len(fltConfigs) > 0 {
		info.Detected = true
		info.Type = "Fairlight (FLT)"
		info.ConfigFile = fltConfigs[0]
	}

	// 5. Check for ALI213 (ali213.ini, SteamConfig.ini)
	aliConfigs := findFilesNamed(gameDir, "ali213.ini", 8)
	if len(aliConfigs) == 0 {
		aliConfigs = findFilesNamed(gameDir, "SteamConfig.ini", 8)
	}
	if len(aliConfigs) > 0 {
		info.Detected = true
		info.Type = "ALI213"
		info.ConfigFile = aliConfigs[0]
	}

	// 6. Check DLC definitions in DLC.txt if not already verified
	if info.DLCStatus == "Default / None" {
		dlcTxtFiles := findFilesNamed(gameDir, "DLC.txt", 8)
		if len(dlcTxtFiles) == 0 {
			dlcTxtFiles = findFilesNamed(gameDir, "dlc.txt", 8)
		}
		if len(dlcTxtFiles) > 0 {
			count := countValidDLCLines(dlcTxtFiles[0])
			if count > 0 {
				info.DLCStatus = strings.TrimSpace(string(rune('0'+count))) + " DLCs Defined (DLC.txt)"
			}
		}
	}

	return info, nil
}

func findFilesNamed(rootDir, targetName string, maxDepth int) []string {
	var matches []string
	baseDepth := strings.Count(rootDir, string(os.PathSeparator))

	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip wine prefix and logs directory
		if info.IsDir() {
			name := info.Name()
			if name == "proton-prefix" || name == ".logs" || name == ".git" {
				return filepath.SkipDir
			}
			curDepth := strings.Count(path, string(os.PathSeparator))
			if curDepth-baseDepth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.EqualFold(info.Name(), targetName) {
			matches = append(matches, path)
		}
		return nil
	})

	return matches
}

func readIniKey(filePath, keyName string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			if strings.EqualFold(k, keyName) {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", nil
}

func countValidDLCLines(filePath string) int {
	f, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, ";") {
			count++
		}
	}
	return count
}
